package main

import (
	"fmt"
	"hprof-tool/pkg/model"
	"hprof-tool/pkg/snapshot"
	"sort"
)

func main() {
	s, err := snapshot.NewSnapshot2("./test-dump-file/heap_dump_test.hprof")
	if err != nil {
		panic(err)
	}
	err = s.EnsureCreateIndex()
	if err != nil {
		panic(err)
	}
	println("EnsureCreateIndex done")

	//threads := s.GetThreads()
	//err = printThreads(s, threads)
	//if err != nil {
	//	panic(err)
	//}

	classes, err := s.ListClassesStatistics()
	if err != nil {
		panic(err)
	}
	printClasses(classes)
}

func printThreads(s *snapshot.Snapshot2, threads map[uint32]*model.Thread) error {
	for id, thread := range threads {
		name, err := s.GetText(thread.NameId)
		if err != nil {
			return err
		}
		groupName, err := s.GetText(thread.GroupNameId)
		if err != nil {
			return err
		}
		groupParentName, err := s.GetText(thread.GroupParentNameId)
		if err != nil {
			return err
		}
		fmt.Printf("Thread %s (%s, %s) with ThreadSerialNumber = %d\n", name, groupName, groupParentName, id)
		for _, frame := range thread.StackTrace.Frames {
			cname, err := s.GetClassNameByClassSerialNumber(uint64(frame.ClassSerialNumber))
			if err != nil {
				return err
			}
			sourceFile, err := s.GetText(frame.SourceFileId)
			if err != nil {
				return err
			}
			method, err := s.GetText(frame.MethodId)
			if err != nil {
				return err
			}
			fmt.Printf("  %s.%s(%s:%d)\n", cname, method, sourceFile, frame.Line)
		}
		println()
	}
	return nil
}

func printClasses(classes []snapshot.ClassStatistics) {
	sort.Slice(classes, func(i, j int) bool {
		return classes[i].InstanceCount > classes[j].InstanceCount
	})
	println("Classes:")
	for _, c := range classes {
		fmt.Printf("%s(%d), %d, %d\n", c.Name, c.Id, c.InstanceCount, c.InstanceSize)
	}
}

func backup_main() {
	s, err := snapshot.NewSnapshot("./test-dump-file/heap_dump_test.hprof")
	if err != nil {
		panic(err)
	}
	err = s.Read()
	if err != nil {
		panic(err)
	}

	classes, err := s.ClassCount()
	instances, err := s.InstanceCount()
	fmt.Printf("Classes: %d, Instances: %d, %v\n", classes, instances, err)
}
