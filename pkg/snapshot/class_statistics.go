package snapshot

type ClassStatistics struct {
	Id            uint64
	Name          string
	InstanceCount int64
	InstanceSize  int64
}

type InstanceStatistics struct {
	Id   uint64
	Size int64
}
