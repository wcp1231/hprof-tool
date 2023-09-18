package model

// LocalFrame 应该就是 native 线程信息
type LocalFrame struct {
	ObjectId   uint64
	LineNumber int32
}

func NewLocalFrame(id uint64, line int32) *LocalFrame {
	return &LocalFrame{
		ObjectId:   id,
		LineNumber: line,
	}
}

type StackFrame struct {
	FrameId           uint64
	MethodId          uint64
	SignatureId       uint64
	SourceFileId      uint64
	ClassSerialNumber uint32
	Line              int32
}

type StackTrace struct {
	ThreadSerialNumber uint32
	Frames             []*StackFrame
	Locals             []*LocalFrame
}

type Thread struct {
	ObjectId          uint64
	NameId            uint64
	GroupNameId       uint64
	GroupParentNameId uint64
	StackTrace        *StackTrace
}
