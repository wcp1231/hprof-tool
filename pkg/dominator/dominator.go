package Dominator

import (
	"fmt"
	"gonum.org/v1/gonum/graph"
	"hprof-tool/pkg/model"
)

type Node struct {
	id int64
	iid uint64
	Name string
	Size uint32
	Retained uint32
	Record interface{}
}

// ID returns the ID number of the node.
func (n Node) ID() int64 {
	return n.id
}

func (n Node) IID() uint64 {
	return n.iid
}

func RootNode() Node {
	return Node{0, 0, "RootNode", 0, 0, nil}
}

func GCNode(objectId uint64, record interface{}) Node {
	return Node{
		id: int64(objectId),
		iid: objectId,
		Name:     "Root",
		Size:     0,
		Retained: 0,
		Record:   record,
	}
}

func NewNode(oid uint64, record model.HProfDumpWithSize) Node {
	if oid == 30724655728 {// || oid == 24697311080 || oid == 30726897728 || oid == 24698751120 {
		fmt.Println("java/lang/reflect/Constructor")
	}
	return Node{
		id: int64(oid),
		iid: oid,
		Name:     "Instance",
		Size:     record.Size(),
		Retained: 0,
		Record:   record,
	}
}

func NewTestNode(id int64, name string, size uint32) Node {
	return Node{
		id: id,
		Name: name,
		Size: size,
	}
}

// Edge is a simple graph edge.
type Edge struct {
	F, T graph.Node
}

// From returns the from-node of the edge.
func (e Edge) From() graph.Node { return e.F }

// To returns the to-node of the edge.
func (e Edge) To() graph.Node { return e.T }

// ReversedLine returns a new Edge with the F and T fields
// swapped.
func (e Edge) ReversedEdge() graph.Edge { return Edge{F: e.T, T: e.F} }
