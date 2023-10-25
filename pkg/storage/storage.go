package storage

import (
	"hprof-tool/pkg/hprof"
)

const (
	TIMESTAMP_KEY = "timestamp"
	THREADS_KEY   = "threads"
)

type Storage interface {
	Init() error
	Close() error

	PutKV(key string, value interface{}) error
	SaveText(id uint64, pos int64) error
	AddText(txt string) (uint64, error)
	GetText(id uint64) (int64, string, error)
	UpdateTextAndPos(id uint32, pos int64, txt string) error

	SaveLoadClass(id uint32, classId uint64, nameId uint64) error
	AddLoadClass(classId uint64, nameId uint64) error
	GetLoadClassById(id uint64) (uint64, uint64, error)
	GetLoadClassByClassId(cid uint64) (uint64, uint64, error)

	SaveClass(pos, cid int64, instanceSize int) error
	AddClass(fakeClass *hprof.HProfClassRecord) (uint64, error)
	GetClass(cid uint64) (int64, *hprof.HProfClassRecord, error)
	ListClasses(fn func(id uint64, pos int64, cla *hprof.HProfClassRecord) error) error

	SaveInstance(pos, oid, cid int64, size int) error
	GetInstanceById(id uint64) (int64, error)
	ListInstances(fn func(id uint64, pos int64, cid uint64) error) error
	ListInstancesByClass(cid uint64, fn func(id uint64, pos, size int64) error) error
	CountInstancesByClass(fn func(cid uint64, count, size int64) error) error

	SaveObjectArray(pos, oid, cid int64, size int) error
	ListObjectArrayByClass(cid uint64, fn func(id uint64, pos, size int64) error) error
	CountObjectArrayByClass(fn func(cid uint64, count, size int64) error) error

	SavePrimitiveArray(pos, oid, typ int64, size int) error
	ListPrimitiveArrayByClass(typ uint64, fn func(id uint64, pos, size int64) error) error
	CountPrimitiveArrayByType(fn func(cid uint64, count, size int64) error) error

	SaveGCRoot(typ int, pos int64) error
	ListGCRoots(fn func(pos int64, typ int) error) error

	SaveThread(r *hprof.HProfThreadRecord) error
	ListThreads(fn func(r *hprof.HProfThreadRecord) error) error
	SaveThreadTrace(r *hprof.HProfTraceRecord) error
	ListThreadTraces(fn func(r *hprof.HProfTraceRecord) error) error
	SaveThreadFrame(r *hprof.HProfFrameRecord) error
	ListThreadFrames(fn func(r *hprof.HProfFrameRecord) error) error

	AppendReference(from, to uint64, typ int) error
	ListInboundReferences(rid uint64, fn func(from uint64, typ int) error) error
	ListOutboundReferences(rid uint64, fn func(to uint64, typ int) error) error

	GetRecordById(id uint64) (int64, int, hprof.HProfRecord, error)
}
