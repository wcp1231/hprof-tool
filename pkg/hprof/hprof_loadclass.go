package hprof

// Load class record.
type HProfLoadClassRecord struct {
	HProfBasicRecord

	// Class serial number.
	ClassSerialNumber uint32
	// Class object ID, associated with HProfClassDump.
	ClassObjectId uint64
	// Stack trace serial number. Mostly unused unless the class is dynamically
	// created and loaded with a custom class loader?
	StackTraceSerialNumber uint32
	// Class name, associated with HProfRecordUTF8.
	ClassNameId uint64
}

func (m *HProfLoadClassRecord) Id() uint64 {
	if m != nil {
		return uint64(m.ClassSerialNumber)
	}
	return 0
}

func (m *HProfLoadClassRecord) Type() HProfRecordType {
	return HProfRecordTypeLoadClass
}

func ReadHProfLoadClassRecord(pr *HProfReader) (*HProfLoadClassRecord, error) {
	pos := pr.pos
	_, err := pr.parseRecordSize()
	if err != nil {
		return nil, err
	}
	csn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	oid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	tsn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	cnid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	size := int(pr.pos - pos)
	return &HProfLoadClassRecord{
		HProfBasicRecord:       HProfBasicRecord{pos, size},
		ClassSerialNumber:      csn,
		ClassObjectId:          oid,
		StackTraceSerialNumber: tsn,
		ClassNameId:            cnid,
	}, nil
}
