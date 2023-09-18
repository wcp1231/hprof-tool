package hprof

// Stack frame record.
type HProfFrameRecord struct {
	HProfBasicRecord

	// Stack frame ID.
	StackFrameId uint64
	// Method name, associated with HProfRecordUTF8.
	MethodNameId uint64
	// Method signature, associated with HProfRecordUTF8.
	MethodSignatureId uint64
	// Source file name, associated with HProfRecordUTF8.
	SourceFileNameId uint64
	// Class serial number, associated with HProfRecordLoadClass.
	ClassSerialNumber uint32
	// Line number if available.
	LineNumber int32
}

func (m *HProfFrameRecord) Id() uint64 {
	if m != nil {
		return m.StackFrameId
	}
	return 0
}

func (m *HProfFrameRecord) Type() HProfRecordType {
	return HProfRecordTypeFrame
}

func ReadHProfFrameRecord(pr *HProfReader) (*HProfFrameRecord, error) {
	pos := pr.pos
	_, err := pr.parseRecordSize()
	if err != nil {
		return nil, err
	}
	sfid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	mnid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	msgnid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	sfnid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	csn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	ln, err := pr.readInt32()
	if err != nil {
		return nil, err
	}
	size := int(pr.pos - pos)
	return &HProfFrameRecord{
		HProfBasicRecord:  HProfBasicRecord{pos, size},
		StackFrameId:      sfid,
		MethodNameId:      mnid,
		MethodSignatureId: msgnid,
		SourceFileNameId:  sfnid,
		ClassSerialNumber: csn,
		LineNumber:        ln,
	}, nil
}
