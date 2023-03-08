package main

import (
	"fmt"
	"hprof-tool/pkg/snapshot"
)

func main() {
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
