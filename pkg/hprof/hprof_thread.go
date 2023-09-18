package hprof

// Stack frame record.
type HProfThreadRecord struct {
	HProfBasicRecord

	// Thread serial number.
	ThreadSerialNumber uint32
	// Thread Object ID.
	ThreadObjectId uint64
	// Stack trace serial number
	StackTraceSerialNumber uint32
	// thread name ID
	ThreadNameId uint64
	// thread group name ID
	ThreadGroupNameId uint64
	// thread group parent name ID
	ThreadGroupParentNameId uint64
}

func (m *HProfThreadRecord) Id() uint64 {
	if m != nil {
		return uint64(m.ThreadSerialNumber)
	}
	return 0
}

func (m *HProfThreadRecord) Type() HProfRecordType {
	return HProfRecordTypeStartThread
}

func ReadHProfThreadRecord(pr *HProfReader) (*HProfThreadRecord, error) {
	pos := pr.pos
	_, err := pr.parseRecordSize()
	if err != nil {
		return nil, err
	}

	tsn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	tid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	stsn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	threadNameId, err := pr.readID()
	if err != nil {
		return nil, err
	}
	threadGroupNameId, err := pr.readID()
	if err != nil {
		return nil, err
	}
	threadGroupParentNameId, err := pr.readID()
	if err != nil {
		return nil, err
	}
	size := int(pr.pos - pos)
	return &HProfThreadRecord{
		HProfBasicRecord:        HProfBasicRecord{pos, size},
		ThreadSerialNumber:      tsn,
		ThreadObjectId:          tid,
		StackTraceSerialNumber:  stsn,
		ThreadNameId:            threadNameId,
		ThreadGroupNameId:       threadGroupNameId,
		ThreadGroupParentNameId: threadGroupParentNameId,
	}, nil
}
