package indexer

import (
	"fmt"
	"hprof-tool/pkg/hprof"
	"hprof-tool/pkg/model"
)

type GCRootsProcessor struct {
	i *Indexer
}

func newGCRootProcessor(i *Indexer) *GCRootsProcessor {
	return &GCRootsProcessor{i}
}

func (p *GCRootsProcessor) process() error {
	println("GCRootsProcessor start")
	return p.i.ForEachGCRoots(func(r hprof.HProfRecord) error {
		switch r.(type) {
		case *hprof.HProfRootJNIGlobal:
			p.onRootJNIGlobalRecord(r.(*hprof.HProfRootJNIGlobal))
		case *hprof.HProfRootJNILocal:
			p.onRootJNILocalRecord(r.(*hprof.HProfRootJNILocal))
		case *hprof.HProfRootJavaFrame:
			p.onRootJavaFrameRecord(r.(*hprof.HProfRootJavaFrame))
		case *hprof.HProfRootStickyClass:
			p.onRootStickyClassRecord(r.(*hprof.HProfRootStickyClass))
		case *hprof.HProfRootThreadObj:
			p.onRootThreadObjRecord(r.(*hprof.HProfRootThreadObj))
		case *hprof.HProfRootMonitorUsed:
			p.onRootMonitorUsedRecord(r.(*hprof.HProfRootMonitorUsed))
		default:
			return fmt.Errorf("unknown gc root type: %#v", r)
		}
		return nil
	})
}

func (p *GCRootsProcessor) onRootJNIGlobalRecord(r *hprof.HProfRootJNIGlobal) {
	p.addGcRoot(r.ObjectId, 0, model.GCRootType_NATIVE_STATIC)
}

func (p *GCRootsProcessor) onRootJNILocalRecord(r *hprof.HProfRootJNILocal) {
	p.addGcRootWithThread(r.ObjectId, r.ThreadSerialNumber, model.GCRootType_NATIVE_LOCAL, int32(r.FrameNumberInStackTrace))
}

func (p *GCRootsProcessor) onRootJavaFrameRecord(r *hprof.HProfRootJavaFrame) {
	p.addGcRootWithThread(r.ObjectId, r.ThreadSerialNumber, model.GCRootType_JAVA_LOCAL, int32(r.FrameNumberInStackTrace))
}

func (p *GCRootsProcessor) onRootStickyClassRecord(r *hprof.HProfRootStickyClass) {
	p.addGcRoot(r.ObjectId, 0, model.GCRootType_SYSTEM_CLASS)
}

func (p *GCRootsProcessor) onRootThreadObjRecord(r *hprof.HProfRootThreadObj) {
	p.i.ctx.thread2Id[r.ThreadSequenceNumber] = r.ThreadObjectId
	p.addGcRoot(r.ThreadObjectId, 0, model.GCRootType_THREAD_OBJ)
}

func (p *GCRootsProcessor) onRootMonitorUsedRecord(r *hprof.HProfRootMonitorUsed) {
	p.addGcRoot(r.ObjectId, 0, model.GCRootType_BUSY_MONITOR)
}

func (p *GCRootsProcessor) addGcRootWithThread(id uint64, threadSerialNumber uint32, typ int, lineNumber int32) {
	threadId, exist := p.i.ctx.thread2Id[threadSerialNumber]
	if exist {
		p.addGcRoot(id, threadId, typ)
	} else {
		p.addGcRoot(id, 0, typ)
	}
	// 记录线程信息
	if lineNumber >= 0 {
		p.i.ctx.thread2locals[threadSerialNumber] = append(p.i.ctx.thread2locals[threadSerialNumber],
			model.NewLocalFrame(id, lineNumber))
	}
}

func (p *GCRootsProcessor) addGcRoot(id uint64, threadId uint64, typ int) {
	if threadId != 0 {
		localAddressToRootInfo, exist := p.i.ctx.threadAddressToLocals[threadId]
		if !exist {
			localAddressToRootInfo = map[uint64][]*model.GCRootInfo{}
		}
		localAddressToRootInfo[id] = append(localAddressToRootInfo[id], model.NewGcRootInfo(id, threadId, typ))
		p.i.ctx.threadAddressToLocals[threadId] = localAddressToRootInfo
	}
	gcRootInfo := model.NewGcRootInfo(id, threadId, typ)
	p.i.ctx.gcRoots[id] = append(p.i.ctx.gcRoots[id], gcRootInfo)
}
