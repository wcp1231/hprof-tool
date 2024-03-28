package hprof

import global "hprof-tool/pkg/util"

// Object array dump.
type HProfObjectArrayRecord struct {
	HProfBasicRecord

	// Object ID.
	ArrayObjectId uint64
	// Stack trace serial number.
	StackTraceSerialNumber uint32
	// Class object ID of the array elements, associated with HProfClassDump.
	ArrayClassObjectId uint64
	// Element object IDs.
	ElementObjectIds []uint64
}

func (m *HProfObjectArrayRecord) Id() uint64 {
	if m != nil {
		return m.ArrayObjectId
	}
	return 0
}

func (m *HProfObjectArrayRecord) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeObjectArrayDump)
}

func (m *HProfObjectArrayRecord) Size() uint32 {
	return uint32(8 + 4 + 8 + 4 + len(m.ElementObjectIds)*global.ID_SIZE)
}

func ReadHProfObjectArrayRecord(pr *HProfReader) (*HProfObjectArrayRecord, error) {
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
	acoid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	vs := []uint64{}
	for i := uint32(0); i < asz; i++ {
		v, err := pr.readID()
		if err != nil {
			return nil, err
		}
		vs = append(vs, v)
	}
	size := int(pr.pos - pos)
	return &HProfObjectArrayRecord{
		HProfBasicRecord:       HProfBasicRecord{pos, size},
		ArrayObjectId:          aoid,
		StackTraceSerialNumber: stsn,
		ArrayClassObjectId:     acoid,
		ElementObjectIds:       vs,
	}, nil
}

func ReadHProfObjectArrayRecordWithPos(pr *HProfReader, pos int64) (*HProfObjectArrayRecord, error) {
	err := pr.Seek(pos)
	if err != nil {
		return nil, err
	}
	return ReadHProfObjectArrayRecord(pr)
}
