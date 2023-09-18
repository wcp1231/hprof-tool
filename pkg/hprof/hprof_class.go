package hprof

// Constant pool entry (appears to be unused according to heapDumper.cpp).
type HProfClass_ConstantPoolEntry struct {
	ConstantPoolIndex uint32
	Type              HProfValueType
	Value             uint64
}

func (m *HProfClass_ConstantPoolEntry) GetType() HProfValueType {
	if m != nil {
		return m.Type
	}
	return HProfValueType_UNKNOWN_HPROF_VALUE_TYPE
}

// Static fields.
type HProfClass_StaticField struct {
	// Static field name, associated with HProfRecordUTF8.
	NameId uint64
	// Type of the static field.
	Type HProfValueType
	// Value of the static field. Must be interpreted based on the type.
	Value uint64
}

func (m *HProfClass_StaticField) GetType() HProfValueType {
	if m != nil {
		return m.Type
	}
	return HProfValueType_UNKNOWN_HPROF_VALUE_TYPE
}

// Instance fields.
type HProfClass_InstanceField struct {
	// Instance field name, associated with HProfRecordUTF8.
	NameId uint64
	// Type of the instance field.
	Type HProfValueType
}

func (m *HProfClass_InstanceField) GetType() HProfValueType {
	if m != nil {
		return m.Type
	}
	return HProfValueType_UNKNOWN_HPROF_VALUE_TYPE
}

// Class data dump.
type HProfClassRecord struct {
	HProfBasicRecord

	// Class object ID.
	ClassObjectId uint64
	// Stack trace serial number.
	StackTraceSerialNumber uint32
	// Super class object ID, associated with another HProfClassDump.
	SuperClassObjectId uint64
	// Class loader object ID, associated with HProfInstanceDump.
	ClassLoaderObjectId uint64
	// Signer of the class. (Looks like ClassLoaders can have signatures...)
	SignersObjectId uint64
	// Protection domain object ID. (No idea)
	ProtectionDomainObjectId uint64
	// Instance size.
	InstanceSize        uint32
	ConstantPoolEntries []*HProfClass_ConstantPoolEntry
	StaticFields        []*HProfClass_StaticField
	InstanceFields      []*HProfClass_InstanceField
}

func (m *HProfClassRecord) Id() uint64 {
	if m != nil {
		return m.ClassObjectId
	}
	return 0
}

func (m *HProfClassRecord) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeClassDump)
}

func ReadHProfClassRecord(pr *HProfReader) (*HProfClassRecord, error) {
	pos := pr.pos
	coid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	stsn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	scoid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	cloid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	sgnoid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	pdoid, err := pr.readID()
	if err != nil {
		return nil, err
	}

	_, err = pr.readID()
	if err != nil {
		return nil, err
	}
	_, err = pr.readID()
	if err != nil {
		return nil, err
	}
	insz, err := pr.readUint32()
	if err != nil {
		return nil, err
	}

	cpsz, err := pr.readUint16()
	if err != nil {
		return nil, err
	}
	cps := []*HProfClass_ConstantPoolEntry{}
	for i := uint16(0); i < cpsz; i++ {
		cpix, err := pr.readUint16()
		if err != nil {
			return nil, err
		}
		ty, err := pr.readByte()
		if err != nil {
			return nil, err
		}
		v, err := pr.readValue(HProfValueType(ty))
		if err != nil {
			return nil, err
		}
		cps = append(cps, &HProfClass_ConstantPoolEntry{
			ConstantPoolIndex: uint32(cpix),
			Type:              HProfValueType(ty),
			Value:             v,
		})
	}

	sfsz, err := pr.readUint16()
	if err != nil {
		return nil, err
	}
	sfs := []*HProfClass_StaticField{}
	for i := uint16(0); i < sfsz; i++ {
		sfnid, err := pr.readID()
		if err != nil {
			return nil, err
		}
		ty, err := pr.readByte()
		if err != nil {
			return nil, err
		}
		v, err := pr.readValue(HProfValueType(ty))
		if err != nil {
			return nil, err
		}
		sfs = append(sfs, &HProfClass_StaticField{
			NameId: sfnid,
			Type:   HProfValueType(ty),
			Value:  v,
		})
	}

	ifsz, err := pr.readUint16()
	if err != nil {
		return nil, err
	}
	ifs := []*HProfClass_InstanceField{}
	for i := uint16(0); i < ifsz; i++ {
		ifnid, err := pr.readID()
		if err != nil {
			return nil, err
		}
		ty, err := pr.readByte()
		if err != nil {
			return nil, err
		}
		ifs = append(ifs, &HProfClass_InstanceField{
			NameId: ifnid,
			Type:   HProfValueType(ty),
		})
	}

	size := int(pr.pos - pos)
	return &HProfClassRecord{
		HProfBasicRecord:         HProfBasicRecord{pos, size},
		ClassObjectId:            coid,
		StackTraceSerialNumber:   stsn,
		SuperClassObjectId:       scoid,
		ClassLoaderObjectId:      cloid,
		SignersObjectId:          sgnoid,
		ProtectionDomainObjectId: pdoid,
		InstanceSize:             insz,
		ConstantPoolEntries:      cps,
		StaticFields:             sfs,
		InstanceFields:           ifs,
	}, nil
}

func ReadHProfClassRecordWithPos(pr *HProfReader, pos int64) (*HProfClassRecord, error) {
	err := pr.Seek(pos)
	if err != nil {
		return nil, err
	}
	return ReadHProfClassRecord(pr)
}
