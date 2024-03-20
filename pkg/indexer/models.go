package indexer

type Instance struct {
	Id     uint64           `json:"id"`
	Class  string           `json:"class"`
	Fields []*InstanceField `json:"fields"`
}

type InstanceField struct {
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Value     string    `json:"value"`
	Reference *Instance `json:"reference"`
}

type Class struct {
	Id                       uint64               `json:"id"`
	LoaderObjectId           uint64               `json:"loader_object_id"`
	SuperClassId             uint64               `json:"super_class_id"`
	StackTraceSerialNumber   uint32               `json:"stack_trace_serial_number"`
	SignersObjectId          uint64               `json:"signers_object_id"`
	Name                     string               `json:"name"`
	ProtectionDomainObjectId uint64               `json:"protection_domain_object_id"`
	InstanceSize             uint32               `json:"instance_size"`
	ConstantPoolEntries      []*ConstantPoolEntry `json:"constant_pools"`
	StaticFields             []*StaticField       `json:"static_fields"`
	InstanceFields           []*InstanceField     `json:"instance_fields"`
}

type ConstantPoolEntry struct {
	Index uint32 `json:"idx"`
	Type  string `json:"type"`
	Value uint64 `json:"value"`
}

type StaticField struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value uint64 `json:"value"`
}
