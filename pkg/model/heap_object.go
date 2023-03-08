package model

type HeapObject struct {
	Class        *HProfClassDump
	IdSize       uint64
	UsedHeapSize int64
	References   []uint64
	Instance     HProfRecord
}

func NewHeapObject() *HeapObject {
	return &HeapObject{
		References: []uint64{},
	}
}
