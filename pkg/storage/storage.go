package storage

import "hprof-tool/pkg/model"

const (
	TIMESTAMP_KEY = "timestamp"
	THREADS_KEY   = "threads"
)

type Storage interface {
	Init() error
	SaveLoadClass(id uint32, classId uint64, nameId uint64) error
	AddLoadClass(classId uint64, nameId uint64) error
	SaveInstance(typ model.HProfRecordType, oid, cid int64, value interface{}) error
	UpdateClassCidToLandClass(classId uint64) error
	UpdatePrimitiveArrayClassId(elementType, classId uint64) error
	PutKV(key string, value interface{}) error
	SaveText(id uint64, pos int64) error
	AddText(txt string) (uint64, error)

	FindLoadClassByName(nameId uint64) (*model.HProfRecordLoadClass, error)
	FindClass(oid uint64) (*model.HProfClassDump, error)
	ListClasses(fn func(dump *model.HProfClassDump) error) error
	ListInstances(fn func(dump model.HProfObjectRecord) error) error
	ListHeapObject(fn func(dump *model.HeapObject) error) error

	GetInboundById(id int) []int
	SetInboundById(id int, ids []int)
	GetOutboundById(id int) []int
	SetOutboundById(id uint64, ids []uint64) error

	CountInstances() (int, error)
}
