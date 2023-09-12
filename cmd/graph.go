package cmd
import (
	"github.com/elliotchance/orderedmap/v2"
)


func (g graph) dimensions() (int, int) {
	return getDrawingSize(g.drawing)
}

func mkGraph(data *orderedmap.OrderedMap[string, []string]) graph {
	g := graph{drawing: mkDrawing(0, 0)}
	for el := data.Front(); el != nil; el = el.Next() {
		nodeName := el.Key
		children := el.Value
		parentNode := g.getOrCreateRootNode(nodeName)
		for _, childNodeName := range children {
			childNode := g.getOrCreateChildNode(parentNode, childNodeName)
			e := edge{from: parentNode, to: childNode, text: ""}
			g.drawEdge(e)
			g.edges = append(g.edges, e)
		}
	}
	return g
}

func doDrawingsCollide(drawing1 drawing, drawing2 drawing, offset coord) bool {
	// Check if any of the drawing2 characters overlap with drawing1 characters.
	// The offset is the coord of drawing2 relative to drawing1.
	drawing1Width, drawing1Height := getDrawingSize(drawing1)
	drawing2Width, drawing2Height := getDrawingSize(drawing2)
	for x := 0; x < drawing2Width; x++ {
		for y := 0; y < drawing2Height; y++ {
			// Check if drawing2[x][y] overlaps with drawing1[x+offset.x][y+offset.y]
			if x+offset.x >= 0 && x+offset.x < drawing1Width &&
				y+offset.y >= 0 && y+offset.y < drawing1Height &&
				drawing2[x][y] != " " &&
				drawing1[x+offset.x][y+offset.y] != " " {
				return true
			}
		}
	}
	return false
}