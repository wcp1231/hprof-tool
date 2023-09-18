package indexer

type thread struct {
	ObjectId          uint64
	NameId            uint64
	GroupNameId       uint64
	GroupParentNameId uint64
}
type stackTrace struct {
	ThreadSerialNumber uint32
	FrameIds           []uint64
}
