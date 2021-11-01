package analyzer

import (
	"container/list"
	"fmt"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/flow"
	"gonum.org/v1/gonum/graph/simple"
	Dominator "hprof-tool/pkg/dominator"
	"hprof-tool/pkg/indexer"
	"hprof-tool/pkg/model"
)

type Analyzer struct {
	idxer *indexer.Indexer

	gcRoots map[int64]Dominator.Node
	rootNode Dominator.Node
	graph *simple.DirectedGraph

	referenceField *model.HProfClassDump_InstanceField
	referenceClasses map[uint64]bool
}

func NewAnalyzer(idxer *indexer.Indexer) *Analyzer {
	return &Analyzer{
		idxer: idxer,
		gcRoots: make(map[int64]Dominator.Node),
		graph: simple.NewDirectedGraph(),
		referenceClasses: make(map[uint64]bool),
	}
}

func (a *Analyzer) InitReference() {
	fmt.Println("InitReference")
	a.referenceField = a.computeReferenceField()
	if a.referenceField != nil {
		a.computeReferenceClasses()
	}
}

func (a *Analyzer) computeReferenceClasses() {
	visited := make(map[uint64]int)
	a.idxer.ForEachClass(func(dump *model.HProfClassDump) error {
		a.dfsFindReferenceClass(visited, dump)
		return nil
	})
}

// 遍历 class 递归 superclass 找 reference
func (a *Analyzer) dfsFindReferenceClass(visited map[uint64]int, dump *model.HProfClassDump) {
	cid := dump.GetClassObjectId()
	if _, exist := visited[cid]; exist {
		return
	}
	name, err := a.idxer.ClassName(dump.GetClassObjectId())
	if err != nil {
		panic(err)
	}
	if name == "java/lang/ref/WeakReference" || name == "java/lang/ref/SoftReference" ||
		name == "java/lang/ref/FinalReference" || name == "java/lang/ref/PhantomReference" {
		a.referenceClasses[cid] = true
		visited[cid] = 2
		return
	}

	sid := dump.GetSuperClassObjectId()
	if sid <= 0 {
		visited[cid] = 1
		return
	}
	val, exist := visited[sid]
	if exist {
		if val == 2 {
			a.referenceClasses[cid] = true
		}
		visited[cid] = val
		return
	}
	super, err := a.idxer.Class(sid)
	if err != nil {
		panic(err)
	}
	if super != nil {
		a.dfsFindReferenceClass(visited, super)
	}
}

func (a *Analyzer) computeReferenceField() *model.HProfClassDump_InstanceField {
	referenceClass, err := a.idxer.ClassByName("java/lang/ref/Reference")
	if err != nil {
		panic(err)
	}
	for _, field := range referenceClass.GetInstanceFields() {
		name, err := a.idxer.String(field.GetNameId())
		if err != nil {
			panic(err)
		}
		if name == "referent" {
			return field
		}
	}
	return nil
}

func (a *Analyzer) BuildGCRoots() {
	fmt.Println("BuildGCRoots")
	a.idxer.ForEachRootJNIGlobal(func(record *model.HProfRootJNIGlobal) error {
		node := Dominator.GCNode(record.ObjectId, record)
		a.gcRoots[node.ID()] = node
		return nil
	})
	a.idxer.ForEachRootJNILocal(func(record *model.HProfRootJNILocal) error {
		node := Dominator.GCNode(record.ObjectId, record)
		a.gcRoots[node.ID()] = node
		return nil
	})
	a.idxer.ForEachRootJavaFrame(func(record *model.HProfRootJavaFrame) error {
		node := Dominator.GCNode(record.ObjectId, record)
		a.gcRoots[node.ID()] = node
		return nil
	})
	a.idxer.ForEachRootStickyClass(func(record *model.HProfRootStickyClass) error {
		node := Dominator.GCNode(record.ObjectId, record)
		a.gcRoots[node.ID()] = node
		return nil
	})
	a.idxer.ForEachRootThreadObj(func(record *model.HProfRootThreadObj) error {
		node := Dominator.GCNode(record.ThreadObjectId, record)
		a.gcRoots[node.ID()] = node
		return nil
	})
	a.idxer.ForEachRootMonitorUsed(func(record *model.HProfRootMonitorUsed) error {
		i, e :=a.idxer.Instance(record.ObjectId)
		fmt.Printf("%v %v\n", i, e)
		node := Dominator.GCNode(record.ObjectId, record)
		a.gcRoots[node.ID()] = node
		return nil
	})
}

func (a *Analyzer) BuildGraph() {
	fmt.Println("BuildGraph")
	queue := list.New()
	a.rootNode = Dominator.RootNode()
	a.graph.AddNode(a.rootNode)
	for _, v := range a.gcRoots {
		edge := a.graph.NewEdge(a.rootNode, v)
		a.addEdgeSafe(edge)

		switch o := v.Record.(type) {
		case *model.HProfRootJNIGlobal:
			instance, err := a.idxer.GetById(o.GetObjectId())
			if err != nil {
				panic(err)
			}
			if instance != nil {
				name, _ := a.idxer.ClassName(instance.GetClassObjectId())
				fmt.Printf("GC-Root-Global: %s\n", name)
				queue.PushBack(Dominator.NewNode(o.GetObjectId(), instance))
			}
		case *model.HProfRootJNILocal:
			instance, err := a.idxer.GetById(o.GetObjectId())
			if err != nil {
				panic(err)
			}
			if instance != nil {
				name, _ := a.idxer.ClassName(instance.GetClassObjectId())
				fmt.Printf("GC-Root-Local: %s\n", name)
				queue.PushBack(Dominator.NewNode(o.GetObjectId(), instance))
			}
		case *model.HProfRootJavaFrame:
			instance, err := a.idxer.GetById(o.GetObjectId())
			if err != nil {
				panic(err)
			}
			if instance != nil {
				name, _ := a.idxer.ClassName(instance.GetClassObjectId())
				fmt.Printf("GC-Root-Frame: %s\n", name)
				queue.PushBack(Dominator.NewNode(o.GetObjectId(), instance))
			}
		case *model.HProfRootStickyClass:
			instance, err := a.idxer.GetById(o.GetObjectId())
			if err != nil {
				panic(err)
			}
			if instance != nil {
				name, _ := a.idxer.ClassName(instance.GetClassObjectId())
				fmt.Printf("GC-Root-Sticky: %s\n", name)
				queue.PushBack(Dominator.NewNode(o.GetObjectId(), instance))
			}
		case *model.HProfRootThreadObj:
			instance, err := a.idxer.GetById(o.GetThreadObjectId())
			if err != nil {
				panic(err)
			}
			if instance != nil {
				name, _ := a.idxer.ClassName(instance.GetClassObjectId())
				fmt.Printf("GC-Root-Thread: %s\n", name)
				queue.PushBack(Dominator.NewNode(o.GetThreadObjectId(), instance))
			}
		case *model.HProfRootMonitorUsed:
			instance, err := a.idxer.GetById(o.GetObjectId())
			if err != nil {
				panic(err)
			}
			if instance != nil {
				name, _ := a.idxer.ClassName(instance.GetClassObjectId())
				fmt.Printf("GC-Root-Monitor: %s\n", name)
				queue.PushBack(Dominator.NewNode(o.GetObjectId(), instance))
			}
		}
	}
	a.bfsBuildGraph(queue)
}

func (a *Analyzer) bfsBuildGraph(queue *list.List) {
	processedClasses := make(map[uint64]bool)
	visited := make(map[int64]bool)
	for queue.Len() > 0 {
		node := (queue.Remove(queue.Front())).(Dominator.Node)
		if _, exist := visited[node.ID()]; exist {
			continue
		}
		visited[node.ID()] = true
		record := node.Record
		switch o := record.(type) {
		case *model.HProfObjectArrayDump:
			for _, eid := range o.GetElementObjectIds() {
				if eid <= 0 {
					continue
				}
				elementRecord, err := a.idxer.GetById(eid)
				if err != nil {
					panic(err)
				}
				if elementRecord != nil {
					element := Dominator.NewNode(eid, elementRecord)
					//a.addNodeSafe(element)
					edge := a.graph.NewEdge(node, element)
					a.addEdgeSafe(edge)
					queue.PushBack(element)
				}
			}
		case *model.HProfClassDump:
			fieldValues := o.GetStaticFields()
			for _, field := range fieldValues {
				if field.Type == model.HProfValueType_OBJECT && field.Value > 0 {
					elementRecord, err := a.idxer.GetById(field.Value)
					if err != nil {
						panic(err)
					}
					if elementRecord != nil {
						element := Dominator.NewNode(field.Value, elementRecord)
						//a.addNodeSafe(element)
						edge := a.graph.NewEdge(node, element)
						a.addEdgeSafe(edge)
						queue.PushBack(element)
					}
				}
			}
		case *model.HProfInstanceDump:
			if o.ObjectId == 30724655728 {
				fmt.Println("Constructor")
			}
			fields, err := a.idxer.FindInstanceObjectReference(o)
			if err != nil {
				panic(err)
			}
			for _, field := range fields {
				if a.isSpecialReference(field, o) || field.Value <= 0 {
					continue
				}
				elementRecord, err := a.idxer.GetById(field.Value)
				if err != nil {
					//panic(err)
					continue
				}
				if elementRecord != nil {
					element := Dominator.NewNode(field.Value, elementRecord)
					edge := a.graph.NewEdge(node, element)
					a.addEdgeSafe(edge)
					queue.PushBack(element)
				}
			}
			if _, exist := processedClasses[o.GetClassObjectId()]; !exist {
				class, err := a.idxer.Class(o.GetClassObjectId())
				if err != nil {
					panic(err)
				}
				element := Dominator.NewNode(class.GetClassObjectId(), class)
				a.addEdgeSafe(a.graph.NewEdge(node, element))
				queue.PushBack(element)
				processedClasses[o.GetClassObjectId()] = true
			}
		}
	}
}

func (a *Analyzer) addEdgeSafe(edge graph.Edge) {
	if edge.From().ID() == edge.To().ID() {
		return
	}
	a.graph.SetEdge(edge)
}

func (a *Analyzer) isSpecialReference(field *model.HProfClassDump_InstanceField, dump *model.HProfInstanceDump) bool {
	_, exist := a.referenceClasses[dump.GetClassObjectId()]
	// InstanceField 有个 Value 字段，没法直接比较。。
	isSameField := field.NameId == a.referenceField.NameId && field.Type == a.referenceField.Type
	return isSameField && exist
}

func (a *Analyzer) BuildDominatorTree() []Dominator.Node {
	fmt.Println("BuildDominatorTree")
	//tree := flow.DominatorsSLT(a.rootNode, a.graph)
	tree := flow.Dominators(a.rootNode, a.graph)
	nodes := a.graph.Nodes()
	var result []Dominator.Node
	for nodes.Next() {
		node := (nodes.Node()).(Dominator.Node)
		if _, exist := a.gcRoots[node.ID()]; exist {
			continue
		}
		//if node.IID() != 0x727645868 {
		//	continue
		//}
		a.dfsComputeRetained(&node, tree)
		//if node.IID() == 24696178480 { // org.springframework.beans.factory.support.DefaultListableBeanFactory
		//	fmt.Printf("%+v\n", node)
		//}
		if node.Retained > 0 {
			result = append(result, node)
		}
	}
	return result
}

func (a *Analyzer) dfsComputeRetained(currentNode *Dominator.Node, tree flow.DominatorTree) {
	if currentNode.Retained > 0 {
		return
	}

	var retained = currentNode.Size
	subNodes := tree.DominatedBy(currentNode.ID())
	for _, subNode := range subNodes {
		node := subNode.(Dominator.Node)
		a.dfsComputeRetained(&node, tree)
		retained += node.Retained

		//hasEdge := a.graph.HasEdgeFromTo(currentNode.ID(), node.ID())
		//if hasEdge {
		//	if node.IID() == 24698432048 {
		//		name, _ := a.idxer.ClassNameById(node.IID())
		//		fmt.Printf("%d\t%d\t%s\n", node.Size, node.Retained, name)
		//	}
        //}
	}
	currentNode.Retained = retained
}


//func (a *Analyzer) dfsFrame(hf *model.HProfRootJavaFrame) {
//	trace, err := a.idxer.Trace(uint64(hf.GetThreadSerialNumber()))
//	if err != nil {
//		panic(err)
//	}
//	frameId := trace.GetStackFrameIds()[hf.GetFrameNumberInStackTrace()]
//	frame, err := a.idxer.Frame(frameId)
//	if err != nil {
//		panic(err)
//	}
//	clazz, err := a.idxer.ClassBySerialNumber(frame.GetClassSerialNumber())
//}