// Copyright (c) 2023 Alexander Grooff
// Copyright (c) 2026 Gregory R. Warnes
// Multi-line label support (<br/> and <br> tags) added by Gregory R. Warnes

package cmd

import (
	"strings"
	"errors"

	"github.com/elliotchance/orderedmap/v2"
	log "github.com/sirupsen/logrus"
)

type genericCoord struct {
	x int
	y int
}

type gridCoord genericCoord
type drawingCoord genericCoord

func (c gridCoord) Equals(other gridCoord) bool {
	return c.x == other.x && c.y == other.y
}
func (c drawingCoord) Equals(other drawingCoord) bool {
	return c.x == other.x && c.y == other.y
}
func (g graph) lineToDrawing(line []gridCoord) []drawingCoord {
	dc := []drawingCoord{}
	for _, c := range line {
		dc = append(dc, g.gridToDrawingCoord(c, nil))
	}
	return dc
}

type graph struct {
	nodes                 []*node
	edges                 []*edge
	drawing               *drawing
	grid                  map[gridCoord]*node
	columnWidth           map[int]int
	rowHeight             map[int]int
	styleClasses          map[string]styleClass
	styleType             string
	paddingX              int
	paddingY              int
	graphDirection        string
	boxBorderPadding      int
	labelWrapWidth        int
	edgeLabelPolicy       string
	edgeLabelMaxWidth     int
	subgraphs             []*subgraph
	offsetX               int
	offsetY               int
	useAscii              bool
	centerMultiLineLabels bool
}

type subgraph struct {
	name     string
	nodes    []*node
	parent   *subgraph
	children []*subgraph
	// Bounding box in drawing coordinates
	minX int
	minY int
	maxX int
	maxY int
}

func mkGraph(data *orderedmap.OrderedMap[string, []textEdge]) graph {
	g := graph{drawing: mkDrawing(0, 0)}
	g.grid = make(map[gridCoord]*node)
	g.columnWidth = make(map[int]int)
	g.rowHeight = make(map[int]int)
	g.styleClasses = make(map[string]styleClass)
	index := 0
	for el := data.Front(); el != nil; el = el.Next() {
		nodeName := el.Key
		children := el.Value
		// Get or create parent node
		parentNode, err := g.getNode(nodeName)
		if err != nil {
			parentNode = &node{name: nodeName, index: index, styleClassName: ""}
			g.appendNode(parentNode)
			index += 1
		}
		for _, textEdge := range children {
			childNode, err := g.getNode(textEdge.child.name)
			if err != nil {
				childNode = &node{name: textEdge.child.name, index: index, styleClassName: textEdge.child.styleClass}
				parentNode.styleClassName = textEdge.parent.styleClass
				g.appendNode(childNode)
				index += 1
			}
			e := edge{from: parentNode, to: childNode, text: textEdge.label}
			g.edges = append(g.edges, &e)
		}
	}
	return g
}

func (g *graph) setStyleClasses(properties *graphProperties) {
	log.Debugf("Setting style classes to %v", properties.styleClasses)
	g.styleClasses = *properties.styleClasses
	g.styleType = properties.styleType
	g.paddingX = properties.paddingX
	g.paddingY = properties.paddingY
	for _, n := range g.nodes {
		if n.styleClassName != "" {
			log.Debugf("Setting style class for node %s to %s", n.name, n.styleClassName)
			(*n).styleClass = g.styleClasses[n.styleClassName]
		}
	}
}

func (g *graph) setLabelLines() {
	for _, n := range g.nodes {
		// Support <br>, <br/>, and \n for multi-line labels
		name := strings.ReplaceAll(strings.ReplaceAll(n.name, "<br/>", "\n"), "<br>", "\n")
		// Always split on \n, wrap only if labelWrapWidth > 0
		n.labelLines = wrapLabelLines(name, g.labelWrapWidth)
		n.labelWidth = maxLineWidth(n.labelLines)
	}
}

func (g *graph) setSubgraphs(textSubgraphs []*textSubgraph) {
	g.subgraphs = []*subgraph{}

	// Convert textSubgraphs to subgraphs with node references
	for _, tsg := range textSubgraphs {
		sg := &subgraph{
			name:     tsg.name,
			nodes:    []*node{},
			children: []*subgraph{},
		}

		// Find and add node references
		for _, nodeName := range tsg.nodes {
			node, err := g.getNode(nodeName)
			if err == nil {
				sg.nodes = append(sg.nodes, node)
			}
		}

		g.subgraphs = append(g.subgraphs, sg)
	}

	// Set up parent-child relationships
	for i, tsg := range textSubgraphs {
		sg := g.subgraphs[i]

		// Set parent
		if tsg.parent != nil {
			for j, parentTsg := range textSubgraphs {
				if parentTsg == tsg.parent {
					sg.parent = g.subgraphs[j]
					break
				}
			}
		}

		// Set children
		for _, childTsg := range tsg.children {
			for j, checkTsg := range textSubgraphs {
				if checkTsg == childTsg {
					sg.children = append(sg.children, g.subgraphs[j])
					break
				}
			}
		}
	}

	log.Debugf("Set %d subgraphs", len(g.subgraphs))
}

func (g *graph) createMapping() {
	// Set mapping coord for every node in the graph
	highestPositionPerLevel := []int{}
	// Init array with 0 values
	// TODO: I'm sure there's a better way of doing this
	for i := 0; i < 100; i++ {
		highestPositionPerLevel = append(highestPositionPerLevel, 0)
	}

	// TODO: should the mapping be bottom-to-top instead of top-to-bottom?
	// Set root nodes to level 0
	nodesFound := make(map[string]bool)
	rootNodes := []*node{}
	for _, n := range g.nodes {
		if _, ok := nodesFound[n.name]; !ok {
			rootNodes = append(rootNodes, n)
		}
		nodesFound[n.name] = true
		for _, child := range g.getChildren(n) {
			nodesFound[child.name] = true
		}
	}

	// Check if we have a mix of external and subgraph root nodes with edges in subgraphs
	// This indicates we should separate them visually in LR layout
	hasExternalRoots := false
	hasSubgraphRootsWithEdges := false
	for _, n := range rootNodes {
		if g.isNodeInAnySubgraph(n) {
			// Check if this node or any node in its subgraph has children
			if len(g.getChildren(n)) > 0 {
				hasSubgraphRootsWithEdges = true
			}
		} else {
			hasExternalRoots = true
		}
	}

	// Separate root nodes by whether they're in subgraphs, but only if we have both types
	// AND there are edges in subgraphs (indicating intentional layout structure)
	shouldSeparate := g.graphDirection == "LR" && hasExternalRoots && hasSubgraphRootsWithEdges

	externalRootNodes := []*node{}
	subgraphRootNodes := []*node{}
	if shouldSeparate {
		for _, n := range rootNodes {
			if g.isNodeInAnySubgraph(n) {
				subgraphRootNodes = append(subgraphRootNodes, n)
			} else {
				externalRootNodes = append(externalRootNodes, n)
			}
		}
	} else {
		// Treat all root nodes the same
		externalRootNodes = rootNodes
	}

	// Place external root nodes first at level 0
	for _, n := range externalRootNodes {
		var mappingCoord *gridCoord
		if g.graphDirection == "LR" {
			mappingCoord = g.reserveSpotInGrid(g.nodes[n.index], &gridCoord{x: 0, y: highestPositionPerLevel[0]})
		} else {
			mappingCoord = g.reserveSpotInGrid(g.nodes[n.index], &gridCoord{x: highestPositionPerLevel[0], y: 0})
		}
		log.Debugf("Setting mapping coord for external rootnode %s to %v", n.name, mappingCoord)
		g.nodes[n.index].gridCoord = mappingCoord
		highestPositionPerLevel[0] = highestPositionPerLevel[0] + 4
	}

	// Place subgraph root nodes at level 4 (one level to the right/down of external nodes)
	// This creates visual separation between external nodes and subgraphs
	if shouldSeparate && len(subgraphRootNodes) > 0 {
		subgraphLevel := 4
		for _, n := range subgraphRootNodes {
			var mappingCoord *gridCoord
			if g.graphDirection == "LR" {
				mappingCoord = g.reserveSpotInGrid(g.nodes[n.index], &gridCoord{x: subgraphLevel, y: highestPositionPerLevel[subgraphLevel]})
			} else {
				mappingCoord = g.reserveSpotInGrid(g.nodes[n.index], &gridCoord{x: highestPositionPerLevel[subgraphLevel], y: subgraphLevel})
			}
			log.Debugf("Setting mapping coord for subgraph rootnode %s to %v", n.name, mappingCoord)
			g.nodes[n.index].gridCoord = mappingCoord
			highestPositionPerLevel[subgraphLevel] = highestPositionPerLevel[subgraphLevel] + 4
		}
	}

	for _, n := range g.nodes {
		log.Debugf("Creating mapping for node %s at %v", n.name, n.gridCoord)
		var childLevel int
		// Next column is 4 coords further. This is because every node is 3 coords wide + 1 coord inbetween.
		if g.graphDirection == "LR" {
			childLevel = n.gridCoord.x + 4
		} else {
			childLevel = n.gridCoord.y + 4
		}
		highestPosition := highestPositionPerLevel[childLevel]
		for _, child := range g.getChildren(n) {
			// Skip if the child already has a mapping coord
			if child.gridCoord != nil {
				continue
			}

			var mappingCoord *gridCoord
			if g.graphDirection == "LR" {
				mappingCoord = g.reserveSpotInGrid(g.nodes[child.index], &gridCoord{x: childLevel, y: highestPosition})
			} else {
				mappingCoord = g.reserveSpotInGrid(g.nodes[child.index], &gridCoord{x: highestPosition, y: childLevel})
			}
			log.Debugf("Setting mapping coord for child %s of parent %s to %v", child.name, n.name, mappingCoord)
			g.nodes[child.index].gridCoord = mappingCoord
			highestPositionPerLevel[childLevel] = highestPosition + 4
		}
	}

	for _, n := range g.nodes {
		g.setColumnWidth(n)
	}

	for _, e := range g.edges {
		g.determinePath(e)
		g.increaseGridSizeForPath(e.path)
		g.determineLabelLine(e)
	}

	// ! Last point before we manipulate the drawing !
	log.Debug("Mapping complete, starting to draw")

	for _, n := range g.nodes {
		dc := g.gridToDrawingCoord(*n.gridCoord, nil)
		g.nodes[n.index].setCoord(&dc)
		g.nodes[n.index].setDrawing(*g)
	}
	g.setDrawingSizeToGridConstraints()

	// Calculate subgraph bounding boxes after nodes are positioned
	g.calculateSubgraphBoundingBoxes()

	// Offset everything if subgraphs have negative coordinates
	g.offsetDrawingForSubgraphs()
}

func (g *graph) calculateSubgraphBoundingBoxes() {
	// Calculate bounding boxes for subgraphs
	// Process innermost subgraphs first (those with no children)
	for _, sg := range g.subgraphs {
		g.calculateSubgraphBoundingBox(sg)
	}

	// Ensure minimum spacing between subgraphs
	g.ensureSubgraphSpacing()
}

func (g *graph) isNodeInAnySubgraph(n *node) bool {
	for _, sg := range g.subgraphs {
		for _, sgNode := range sg.nodes {
			if sgNode == n {
				return true
			}
		}
	}
	return false
}

func (g *graph) getNodeSubgraph(n *node) *subgraph {
	for _, sg := range g.subgraphs {
		for _, sgNode := range sg.nodes {
			if sgNode == n {
				return sg
			}
		}
	}
	return nil
}

func (g *graph) hasIncomingEdgeFromOutsideSubgraph(n *node) bool {
	nodeSubgraph := g.getNodeSubgraph(n)
	if nodeSubgraph == nil {
		return false // Node not in any subgraph
	}

	// Check if any edge targets this node from outside its subgraph
	hasExternalEdge := false
	for _, edge := range g.edges {
		if edge.to == n {
			sourceSubgraph := g.getNodeSubgraph(edge.from)
			// If source is not in the same subgraph (or any subgraph), it's from outside
			if sourceSubgraph != nodeSubgraph {
				hasExternalEdge = true
				break
			}
		}
	}

	if !hasExternalEdge {
		return false
	}

	// Only apply overhead if this is the topmost node in the subgraph with external edges
	// (has the lowest Y coordinate among nodes with external edges)
	for _, otherNode := range nodeSubgraph.nodes {
		if otherNode == n || otherNode.gridCoord == nil {
			continue
		}
		// Check if otherNode also has external edges and is at a lower Y
		otherHasExternal := false
		for _, edge := range g.edges {
			if edge.to == otherNode {
				sourceSubgraph := g.getNodeSubgraph(edge.from)
				if sourceSubgraph != nodeSubgraph {
					otherHasExternal = true
					break
				}
			}
		}
		if otherHasExternal && otherNode.gridCoord.y < n.gridCoord.y {
			// There's another node higher up that has external edges
			return false
		}
	}

	return true
}

func (g *graph) ensureSubgraphSpacing() {
	const minSpacing = 1 // Minimum lines between subgraphs

	// Only check root-level subgraphs (those without parents)
	rootSubgraphs := []*subgraph{}
	for _, sg := range g.subgraphs {
		if sg.parent == nil && len(sg.nodes) > 0 {
			rootSubgraphs = append(rootSubgraphs, sg)
		}
	}

	// Check each pair of root subgraphs for overlaps
	for i := 0; i < len(rootSubgraphs); i++ {
		for j := i + 1; j < len(rootSubgraphs); j++ {
			sg1 := rootSubgraphs[i]
			sg2 := rootSubgraphs[j]

			// Check if they overlap or are too close
			// Vertical overlap check (for TD layout)
			if sg1.minX < sg2.maxX && sg1.maxX > sg2.minX {
				// They share the same X space, check Y spacing
				if sg1.maxY >= sg2.minY-minSpacing && sg1.minY < sg2.minY {
					// sg1 is above sg2 and too close, extend sg2 upward to create space
					newMinY := sg1.maxY + minSpacing + 1
					log.Debugf("Extending subgraph %s minY to %d (from %d) to add spacing from %s", sg2.name, newMinY, sg2.minY, sg1.name)
					sg2.minY = newMinY
				} else if sg2.maxY >= sg1.minY-minSpacing && sg2.minY < sg1.minY {
					// sg2 is above sg1 and too close, extend sg1 upward to create space
					newMinY := sg2.maxY + minSpacing + 1
					log.Debugf("Extending subgraph %s minY to %d (from %d) to add spacing from %s", sg1.name, newMinY, sg1.minY, sg2.name)
					sg1.minY = newMinY
				}
			}

			// Horizontal overlap check (for LR layout)
			if sg1.minY < sg2.maxY && sg1.maxY > sg2.minY {
				// They share the same Y space, check X spacing
				if sg1.maxX >= sg2.minX-minSpacing && sg1.minX < sg2.minX {
					// sg1 is left of sg2 and too close, extend sg2 leftward to create space
					newMinX := sg1.maxX + minSpacing + 1
					log.Debugf("Extending subgraph %s minX to %d (from %d) to add spacing from %s", sg2.name, newMinX, sg2.minX, sg1.name)
					sg2.minX = newMinX
				} else if sg2.maxX >= sg1.minX-minSpacing && sg2.minX < sg1.minX {
					// sg2 is left of sg1 and too close, extend sg1 leftward to create space
					newMinX := sg2.maxX + minSpacing + 1
					log.Debugf("Extending subgraph %s minX to %d (from %d) to add spacing from %s", sg1.name, newMinX, sg1.minX, sg2.name)
					sg1.minX = newMinX
				}
			}
		}
	}
}

func (g *graph) calculateSubgraphBoundingBox(sg *subgraph) {
	if len(sg.nodes) == 0 {
		return
	}

	// Start with impossible bounds
	minX := 1000000
	minY := 1000000
	maxX := -1000000
	maxY := -1000000

	// First, calculate bounding box for all child subgraphs
	for _, child := range sg.children {
		g.calculateSubgraphBoundingBox(child)
		if len(child.nodes) > 0 {
			minX = Min(minX, child.minX)
			minY = Min(minY, child.minY)
			maxX = Max(maxX, child.maxX)
			maxY = Max(maxY, child.maxY)
		}
	}

	// Then include all direct nodes
	for _, node := range sg.nodes {
		if node.drawingCoord == nil || node.drawing == nil {
			continue
		}

		// Get the actual bounds of the node's drawing
		nodeMinX := node.drawingCoord.x
		nodeMinY := node.drawingCoord.y
		nodeMaxX := nodeMinX + len(*node.drawing) - 1
		nodeMaxY := nodeMinY + len((*node.drawing)[0]) - 1

		minX = Min(minX, nodeMinX)
		minY = Min(minY, nodeMinY)
		maxX = Max(maxX, nodeMaxX)
		maxY = Max(maxY, nodeMaxY)
	}

	// Add padding (allow negative coordinates, we'll offset later)
	const subgraphPadding = 2
	const subgraphLabelSpace = 2 // Extra space for label at top
	sg.minX = minX - subgraphPadding
	sg.minY = minY - subgraphPadding - subgraphLabelSpace
	sg.maxX = maxX + subgraphPadding
	sg.maxY = maxY + subgraphPadding

	log.Debugf("Subgraph %s bounding box: (%d,%d) to (%d,%d)", sg.name, sg.minX, sg.minY, sg.maxX, sg.maxY)
}

func (g *graph) offsetDrawingForSubgraphs() {
	if len(g.subgraphs) == 0 {
		return
	}

	// Find the minimum coordinates across all subgraphs
	minX := 0
	minY := 0
	for _, sg := range g.subgraphs {
		minX = Min(minX, sg.minX)
		minY = Min(minY, sg.minY)
	}

	// If we have negative coordinates, we need to offset everything
	offsetX := -minX
	offsetY := -minY

	if offsetX == 0 && offsetY == 0 {
		return
	}

	log.Debugf("Offsetting drawing by (%d, %d)", offsetX, offsetY)

	// Store the offset in the graph so it can be applied during drawing
	g.offsetX = offsetX
	g.offsetY = offsetY

	// Offset all subgraph coordinates
	for _, sg := range g.subgraphs {
		sg.minX += offsetX
		sg.minY += offsetY
		sg.maxX += offsetX
		sg.maxY += offsetY
	}

	// Offset all node coordinates (they were set before offset was calculated)
	for _, n := range g.nodes {
		if n.drawingCoord != nil {
			n.drawingCoord.x += offsetX
			n.drawingCoord.y += offsetY
		}
	}
}

func (g *graph) draw() *drawing {
	// Draw subgraphs first (outermost to innermost) so they appear in the background
	g.drawSubgraphs()

	// Draw all nodes.
	for _, node := range g.nodes {
		if !node.drawn {
			g.drawNode(node)
		}
	}
	lineDrawings := []*drawing{}
	cornerDrawings := []*drawing{}
	arrowHeadDrawings := []*drawing{}
	boxStartDrawings := []*drawing{}
	labelDrawings := []*drawing{}
	for _, edge := range g.edges {
		line, boxStart, arrowHead, corners, label := g.drawEdge(edge)
		lineDrawings = append(lineDrawings, line)
		cornerDrawings = append(cornerDrawings, corners)
		arrowHeadDrawings = append(arrowHeadDrawings, arrowHead)
		boxStartDrawings = append(boxStartDrawings, boxStart)
		labelDrawings = append(labelDrawings, label)
	}

	// Draw in order
	g.drawing = g.mergeDrawings(g.drawing, drawingCoord{0, 0}, lineDrawings...)
	g.drawing = g.mergeDrawings(g.drawing, drawingCoord{0, 0}, cornerDrawings...)
	g.drawing = g.mergeDrawings(g.drawing, drawingCoord{0, 0}, arrowHeadDrawings...)
	g.drawing = g.mergeDrawings(g.drawing, drawingCoord{0, 0}, boxStartDrawings...)
	g.drawing = g.mergeDrawings(g.drawing, drawingCoord{0, 0}, labelDrawings...)

	// Draw subgraph labels LAST so they don't get overwritten by arrows
	g.drawSubgraphLabels()

	return g.drawing
}

func (g *graph) drawSubgraphs() {
	// Sort subgraphs by depth (outermost first)
	// We'll draw parents before children
	sorted := g.sortSubgraphsByDepth()

	for _, sg := range sorted {
		sgDrawing := drawSubgraph(sg, *g)
		// Position the drawing at the subgraph's min coordinates
		offset := drawingCoord{sg.minX, sg.minY}
		g.drawing = g.mergeDrawings(g.drawing, offset, sgDrawing)
	}
}

func (g *graph) drawSubgraphLabels() {
	// Draw labels for all subgraphs
	// No need to sort - labels can be drawn in any order since they don't overlap
	for _, sg := range g.subgraphs {
		if len(sg.nodes) == 0 {
			continue
		}
		labelDrawing, offset := drawSubgraphLabel(sg, *g)
		g.drawing = g.mergeDrawings(g.drawing, offset, labelDrawing)
	}
}

func (g *graph) sortSubgraphsByDepth() []*subgraph {
	// Calculate depth for each subgraph
	depths := make(map[*subgraph]int)
	for _, sg := range g.subgraphs {
		depths[sg] = g.getSubgraphDepth(sg)
	}

	// Sort by depth (lower depth = outermost = drawn first)
	sorted := make([]*subgraph, len(g.subgraphs))
	copy(sorted, g.subgraphs)

	// Simple bubble sort by depth
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if depths[sorted[i]] > depths[sorted[j]] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

func (g *graph) getSubgraphDepth(sg *subgraph) int {
	if sg.parent == nil {
		return 0
	}
	return 1 + g.getSubgraphDepth(sg.parent)
}

func (g *graph) getNode(nodeName string) (*node, error) {
	for _, n := range g.nodes {
		if n.name == nodeName {
			return n, nil
		}
	}
	return &node{}, errors.New("node " + nodeName + " not found")
}

func (g *graph) appendNode(n *node) {
	g.nodes = append(g.nodes, n)
}

func (g graph) getEdgesFromNode(n *node) []edge {
	edges := []edge{}
	for _, edge := range g.edges {
		if (edge.from.name) == (n.name) {
			edges = append(edges, *edge)
		}
	}
	return edges
}

func (g *graph) getChildren(n *node) []*node {
	edges := g.getEdgesFromNode(n)
	children := []*node{}
	for _, edge := range edges {
		if edge.from.name == n.name {
			children = append(children, edge.to)
		}
	}
	return children
}

func (g *graph) gridToDrawingCoord(c gridCoord, dir *direction) drawingCoord {
	x := 0
	y := 0
	var target gridCoord
	if dir == nil {
		target = c
	} else {
		target = gridCoord{x: c.x + dir.x, y: c.y + dir.y}
	}
	for column := 0; column < target.x; column++ {
		x += g.columnWidth[column]
	}
	for row := 0; row < target.y; row++ {
		y += g.rowHeight[row]
	}
	dc := drawingCoord{x: x + g.columnWidth[target.x]/2 + g.offsetX, y: y + g.rowHeight[target.y]/2 + g.offsetY}

	return dc
}
