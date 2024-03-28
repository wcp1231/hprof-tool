package hprof

// Primitive array dump.
type HProfPrimitiveArrayRecord struct {
	HProfBasicRecord

	// Object ID.
	ArrayObjectId uint64
	// Stack trace serial number.
	StackTraceSerialNumber uint32
	// Type of the elements.
	ElementType HProfValueType
	// Element values.
	//
	// Values need to be parsed based on the element_type. If the array is an int
	// array with three elements, this field has 12 bytes.
	Values []byte
}

func (m *HProfPrimitiveArrayRecord) Id() uint64 {
	if m != nil {
		return m.ArrayObjectId
	}
	return 0
}

func (m *HProfPrimitiveArrayRecord) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypePrimitiveArrayDump)
}

func ReadHProfPrimitiveArrayRecord(pr *HProfReader) (*HProfPrimitiveArrayRecord, error) {
	pos := pr.pos
	aoid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	stsn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	asz, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	ty, err := pr.readByte()
	if err != nil {
		return nil, err
	}
	bs, err := pr.readArray(HProfValueType(ty), int(asz))
	if err != nil {
		return nil, err
	}
	size := int(pr.pos - pos)
	return &HProfPrimitiveArrayRecord{
		HProfBasicRecord:       HProfBasicRecord{pos, size},
		ArrayObjectId:          aoid,
		StackTraceSerialNumber: stsn,
		ElementType:            HProfValueType(ty),
		Values:                 bs,
	}, nil
}

func ReadHProfPrimitiveArrayRecordWithPos(pr *HProfReader, pos int64) (*HProfPrimitiveArrayRecord, error) {
	err := pr.Seek(pos)
	if err != nil {
		return nil, err
	}
	return ReadHProfPrimitiveArrayRecord(pr)
}
