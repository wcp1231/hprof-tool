package model

const (
	GCRootType_UNKNOWN          = 1 << 0
	GCRootType_SYSTEM_CLASS     = 1 << 1
	GCRootType_NATIVE_LOCAL     = 1 << 2
	GCRootType_NATIVE_STATIC    = 1 << 3
	GCRootType_THREAD_BLOCK     = 1 << 4
	GCRootType_BUSY_MONITOR     = 1 << 5
	GCRootType_JAVA_LOCAL       = 1 << 6
	GCRootType_NATIVE_STACK     = 1 << 7
	GCRootType_THREAD_OBJ       = 1 << 8
	GCRootType_FINALIZABLE      = 1 << 9
	GCRootType_UNFINALIZED      = 1 << 10
	GCRootType_UNREACHABLE      = 1 << 11
	GCRootType_JAVA_STACK_FRAME = 1 << 12
)

type GCRootInfo struct {
	ID       uint64
	ThreadId uint64
	Typ      int

	// 不确定有什么用
	ObjectId  uint64
	ContextId uint64 // 就是 ThreadId
}

func NewGcRootInfo(id uint64, threadId uint64, typ int) *GCRootInfo {
	return &GCRootInfo{
		ID:       id,
		ThreadId: threadId,
		Typ:      typ,
	}
}
