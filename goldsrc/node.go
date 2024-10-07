package goldsrc

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Original retail version.
const NodeGraphVersion = 16

type NodeFormat int

const (
	NodeFormatValve NodeFormat = iota
	NodeFormatDecay
)

type Graph struct {
	_ [3]int32  // 3 qbool
	_ [3]uint32 // 3 pointers

	NumNodes int32
	NumLinks int32

	_ [8364]byte
}

type Vec3f struct {
	X, Y, Z float32
}

func (vec Vec3f) String() string { // .map-compatible
	return fmt.Sprintf("%f %f %f", vec.X, vec.Y, vec.Z)
}

type Node interface {
	Position(original bool) Vec3f // false to get position after being dropped.
	ClassName() string
}

type ValveNode struct {
	Origin     Vec3f
	OriginPeek Vec3f

	_ [3]byte

	NodeInfo int32

	_ [57]byte
}

type DecayNode struct {
	ValveNode
	_ [8]byte
}

func (node ValveNode) Position(original bool) Vec3f {
	if original {
		return node.Origin
	}

	return node.OriginPeek
}

const (
	NodeTypeLand  int32 = 1 << 0
	NodeTypeAir   int32 = 1 << 1
	NodeTypeWater int32 = 1 << 2
)

func (node ValveNode) ClassName() string {
	// Order matters.
	switch {
	case node.NodeInfo == 256:
		// HACK: No idea where this 256 comes from.
		return "info_node"
	case (node.NodeInfo & NodeTypeLand) != 0, node.NodeInfo == 0:
		return "info_node"
	case (node.NodeInfo & NodeTypeWater) != 0:
		return "info_node_water"
	case (node.NodeInfo & NodeTypeAir) != 0:
		return "info_node_air"
	default:
		return fmt.Sprintf("info_node_unknown_%d", node.NodeInfo)
	}
}

const (
	LinkTypeSmallHull = iota
	LinkTypeHumanHull
	LinkTypeLargeHull
	LinkTypeFlyHull
	LinkTypeDisabledHull

	LinkTypeBitMax = 4
)

func LinkTypeName(id int) string {
	switch id {
	case LinkTypeSmallHull:
		return "small"
	case LinkTypeHumanHull:
		return "human"
	case LinkTypeLargeHull:
		return "large"
	case LinkTypeFlyHull:
		return "fly"
	case LinkTypeDisabledHull:
		return "disabled"
	default:
		return "unknown"
	}
}

type Link struct {
	SrcNode  int32
	DstNode  int32
	_        [8]byte
	LinkInfo int32
	_        [4]byte
}

func ReadNodes(r io.Reader, format NodeFormat) ([]Node, []Link, error) {
	var version int32
	if err := binary.Read(r, binary.LittleEndian, &version); err != nil {
		return nil, nil, fmt.Errorf("unable to parse version: %w", err)
	}
	if version != NodeGraphVersion {
		return nil, nil, fmt.Errorf("unsupported node graph version: %d", version)
	}

	var graph Graph
	if err := binary.Read(r, binary.LittleEndian, &graph); err != nil {
		return nil, nil, fmt.Errorf("unable to read CGraph: %w", err)
	}

	var nodes = make([]Node, 0, graph.NumNodes)
	for i := int32(0); i < graph.NumNodes; i++ {
		node, err := readNode(r, format)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to read node #%d: %w", i, err)
		}

		nodes = append(nodes, node)
	}

	var links = make([]Link, 0, graph.NumLinks)
	for i := int32(0); i < graph.NumLinks; i++ {
		var link Link
		if err := binary.Read(r, binary.LittleEndian, &link); err != nil {
			return nil, nil, fmt.Errorf("unable to read CLink: %w", err)
		}

		links = append(links, link)
	}

	return nodes, links, nil
}

func readNode(r io.Reader, format NodeFormat) (Node, error) {
	switch format {
	case NodeFormatValve:
		var node ValveNode
		var err = binary.Read(r, binary.LittleEndian, &node)
		return node, err
	case NodeFormatDecay:
		var node DecayNode
		var err = binary.Read(r, binary.LittleEndian, &node)
		return node, err
	default:
		return nil, fmt.Errorf("unknown node format: %d", format)
	}
}

func assertSizeof(typ any, expected int) {
	var actual = binary.Size(typ)
	if actual != expected {
		panic(fmt.Errorf("invalid size for %T: got %d expected %d", typ, actual, expected))
	}
}

func init() {
	assertSizeof(Graph{}, 8396)
	assertSizeof(ValveNode{}, 88)
	assertSizeof(DecayNode{}, 96)
	assertSizeof(Link{}, 24)
}
