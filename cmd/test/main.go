package main

import (
	"fmt"
	"gonum.org/v1/gonum/graph/flow"
	"gonum.org/v1/gonum/graph/simple"
	Dominator "hprof-tool/pkg/dominator"
)

func main() {
	g := simple.NewDirectedGraph()
	root := Dominator.NewTestNode(1, "Root", 1)
	obj2 := Dominator.NewTestNode(2, "Obj2", 13)
	obj3 := Dominator.NewTestNode(3, "Obj3", 11)
	obj4 := Dominator.NewTestNode(4, "Obj4", 7)
	obj5 := Dominator.NewTestNode(5, "Obj5", 5)
	obj6 := Dominator.NewTestNode(6, "Obj6", 3)
	obj7 := Dominator.NewTestNode(7, "Obj7", 1)
	obj8 := Dominator.NewTestNode(8, "Obj8", 10)
	g.SetEdge(g.NewEdge(root, obj2))
	g.SetEdge(g.NewEdge(root, obj3))
	g.SetEdge(g.NewEdge(obj2, obj4))
	g.SetEdge(g.NewEdge(obj2, obj5))
	g.SetEdge(g.NewEdge(obj3, obj6))
	g.SetEdge(g.NewEdge(obj3, obj7))

	g.SetEdge(g.NewEdge(obj6, obj5))
	g.SetEdge(g.NewEdge(obj5, obj8))

	tree := flow.DominatorsSLT(root, g)
	fmt.Printf("%+v\n", tree)
}
