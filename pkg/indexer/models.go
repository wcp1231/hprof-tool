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
