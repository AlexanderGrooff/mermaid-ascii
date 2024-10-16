package cmd

import (
	"errors"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/sirupsen/logrus"
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
	nodes        []*node
	edges        []*edge
	drawing      *drawing
	grid         map[gridCoord]*node
	columnWidth  map[int]int
	rowHeight    map[int]int
	styleClasses map[string]styleClass
	styleType    string
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
	logrus.Debugf("Setting style classes to %v", properties.styleClasses)
	g.styleClasses = *properties.styleClasses
	g.styleType = properties.styleType
	for _, n := range g.nodes {
		if n.styleClassName != "" {
			logrus.Debugf("Setting style class for node %s to %s", n.name, n.styleClassName)
			(*n).styleClass = g.styleClasses[n.styleClassName]
		}
	}
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
	for _, n := range rootNodes {
		var mappingCoord *gridCoord
		if graphDirection == "LR" {
			mappingCoord = g.reserveSpotInGrid(g.nodes[n.index], &gridCoord{x: 0, y: highestPositionPerLevel[0]})
		} else {
			mappingCoord = g.reserveSpotInGrid(g.nodes[n.index], &gridCoord{x: highestPositionPerLevel[0], y: 0})
		}
		logrus.Debugf("Setting mapping coord for rootnode %s to %v", n.name, mappingCoord)
		g.nodes[n.index].gridCoord = mappingCoord
		highestPositionPerLevel[0] = highestPositionPerLevel[0] + 4
	}

	for _, n := range g.nodes {
		log.Debugf("Creating mapping for node %s at %v", n.name, n.gridCoord)
		var childLevel int
		// Next column is 4 coords further. This is because every node is 3 coords wide + 1 coord inbetween.
		if graphDirection == "LR" {
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
			if graphDirection == "LR" {
				mappingCoord = g.reserveSpotInGrid(g.nodes[child.index], &gridCoord{x: childLevel, y: highestPosition})
			} else {
				mappingCoord = g.reserveSpotInGrid(g.nodes[child.index], &gridCoord{x: highestPosition, y: childLevel})
			}
			logrus.Debugf("Setting mapping coord for child %s of parent %s to %v", child.name, n.name, mappingCoord)
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
}

func (g *graph) draw() *drawing {
	// Draw all nodes.
	for _, node := range g.nodes {
		if !node.drawn {
			g.drawNode(node)
		}
	}
	lineDrawings := []*drawing{}
	cornerDrawings := []*drawing{}
	arrowHeadDrawings := []*drawing{}
	labelDrawings := []*drawing{}
	for _, edge := range g.edges {
		line, corners, arrowHead, label := g.drawEdge(edge)
		lineDrawings = append(lineDrawings, line)
		cornerDrawings = append(cornerDrawings, corners)
		arrowHeadDrawings = append(arrowHeadDrawings, arrowHead)
		labelDrawings = append(labelDrawings, label)
	}

	// Draw in order
	g.drawing = mergeDrawings(g.drawing, drawingCoord{0, 0}, lineDrawings...)
	g.drawing = mergeDrawings(g.drawing, drawingCoord{0, 0}, cornerDrawings...)
	g.drawing = mergeDrawings(g.drawing, drawingCoord{0, 0}, arrowHeadDrawings...)
	g.drawing = mergeDrawings(g.drawing, drawingCoord{0, 0}, labelDrawings...)
	return g.drawing
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
	dc := drawingCoord{x: x + g.columnWidth[target.x]/2, y: y + g.rowHeight[target.y]/2}

	return dc
}
