package hprof

// Stack trace record.
type HProfTraceRecord struct {
	HProfBasicRecord

	// Stack trace serial number.
	StackTraceSerialNumber uint32
	// Thread serial number.
	ThreadSerialNumber uint32
	// Stack frame IDs, associated with HProfRecordFrame.
	StackFrameIds []uint64
}

func (m *HProfTraceRecord) Id() uint64 {
	if m != nil {
		return uint64(m.StackTraceSerialNumber)
	}
	return 0
}

func (m *HProfTraceRecord) Type() HProfRecordType {
	return HProfRecordTypeTrace
}

func ReadHProfTraceRecord(pr *HProfReader) (*HProfTraceRecord, error) {
	pos := pr.pos
	_, err := pr.parseRecordSize()
	if err != nil {
		return nil, err
	}

	stsn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	tsn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	nr, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	sfids := []uint64{}
	for i := uint32(0); i < nr; i++ {
		sfid, err := pr.readID()
		if err != nil {
			return nil, err
		}
		sfids = append(sfids, sfid)
	}
	size := int(pr.pos - pos)
	return &HProfTraceRecord{
		HProfBasicRecord:       HProfBasicRecord{pos, size},
		StackTraceSerialNumber: stsn,
		ThreadSerialNumber:     tsn,
		StackFrameIds:          sfids,
	}, nil
}
