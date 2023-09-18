package indexer

import (
	"hprof-tool/pkg/model"
)

// HeapContext 存储分析 HProf 文件时的缓存结果
type HeapContext struct {
	// 用于计算 GCRoot 和线程信息
	thread2Id map[uint32]uint64
	gcRoots   map[uint64][]*model.GCRootInfo
	// map[threadId]map[id][]GcRoot
	threadAddressToLocals map[uint64]map[uint64][]*model.GCRootInfo

	// 线程相关
	thread2locals     map[uint32][]*model.LocalFrame
	id2frame          map[uint64]*model.StackFrame
	serNum2stackTrace map[uint32]*stackTrace
	threadSN2thread   map[uint32]*thread

	// map[className]classId
	className2Cid map[string][]uint64
	classId2Name  map[uint64]string

	// references
	classReferences    map[uint64][]uint64
	instanceReferences map[uint64][]uint64
}

func newHeapContext() *HeapContext {
	return &HeapContext{
		thread2Id:             map[uint32]uint64{},
		gcRoots:               map[uint64][]*model.GCRootInfo{},
		threadAddressToLocals: map[uint64]map[uint64][]*model.GCRootInfo{},

		thread2locals:     map[uint32][]*model.LocalFrame{},
		id2frame:          map[uint64]*model.StackFrame{},
		serNum2stackTrace: map[uint32]*stackTrace{},
		threadSN2thread:   map[uint32]*thread{},

		className2Cid: map[string][]uint64{},
		classId2Name:  map[uint64]string{},

		classReferences:    map[uint64][]uint64{},
		instanceReferences: map[uint64][]uint64{},
	}
}
