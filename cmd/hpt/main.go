package main

import (
	"fmt"
	"hprof-tool/pkg/analyzer"
	"hprof-tool/pkg/indexer"
	"sort"
)

// An Item is something we manage in a priority queue.
type Item struct {
	Value    uint64 // The value of the item; arbitrary.
	Priority uint32    // The priority of the item in the queue.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].Priority > pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func main() {
	idxer, err := indexer.OpenOrCreateIndex(
		"./test-dump-file/heap_dump_test.hprof",
		"./test-dump-file/heap_dump_test.index")
	if err != nil {
		panic(err)
	}

	/*
	class, err := idxer.ClassByName("java/lang/ApplicationShutdownHooks")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v %d\n", class, class.InstanceSize)
	idxer.ForEachInstance(func(instanceDump *model.HProfInstanceDump) error {
		if instanceDump.ClassObjectId == class.ClassObjectId {
			fmt.Printf("%+v %d\n", instanceDump, instanceDump.Size())
		}
		return nil
	})*/

	/*
	var iid uint64 = 0x727645868
	inst, _ :=idxer.Instance(iid)
	class, _ := idxer.Class(inst.GetClassObjectId())
	fmt.Printf("%d %+v", inst.Size(), class)
	 */


	fmt.Println("Start analyze")
	anlz := analyzer.NewAnalyzer(idxer)
	anlz.InitReference()
	anlz.BuildGCRoots()
	anlz.BuildGraph()
	nodes := anlz.BuildDominatorTree()
	var pq PriorityQueue
	for _, node := range nodes {
		pq = append(pq, &Item{
			Value: node.IID(),
			Priority: node.Retained,
		})
		//if node.IID() == iid {
		//	name, _ := idxer.ClassNameById(node.IID())
		//	fmt.Printf("%s %v\n", name, node.Retained)
		//	return
		//}
	}
	sort.Sort(pq)
	for i, v := range pq {
		name, err := idxer.ClassNameById(v.Value)
		if err != nil {
			name = fmt.Sprintf("%d", v.Value)
		}
		fmt.Printf("%s\t\t%d\n", name, v.Priority)
		if i > 10 {
			break
		}
	}
}
