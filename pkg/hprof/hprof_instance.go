package hprof

import (
	"bytes"
)

// Instance dump.
type HProfInstanceRecord struct {
	HProfBasicRecord

	// Object ID.
	ObjectId uint64
	// Stack trace serial number.
	StackTraceSerialNumber uint32
	// Class object ID, associated with HProfClassDump.
	ClassObjectId uint64
	// Instance field values.
	//
	// The instance field values are serialized in the order of the instance field
	// definition of HProfClassDump. If the class has three int fields, this
	// values starts from three 4-byte integers. Then, it continues to the super
	// class's instance fields.
	Values []byte
}

func (m *HProfInstanceRecord) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeInstanceDump)
}

func (m *HProfInstanceRecord) Size() uint32 {
	//return uint32(8 + 4 + 8 + 4 + len(m.Values))
	return uint32(len(m.Values) + 16) // TODO 空 instance 大小是 16 还是 0 ？
	// TODO 会调整大小，变成 8 的整数倍，得确认具体逻辑
}

func (m *HProfInstanceRecord) ReadValues(fields []*HProfClass_InstanceField) ([]HProfInstanceFieldValue, error) {
	reader := bytes.NewReader(m.Values)
	result := []HProfInstanceFieldValue{}
	for _, field := range fields {
		var fv HProfInstanceFieldValue
		var err error
		switch field.Type {
		case HProfValueType_BOOLEAN:
			fv, err = readBooleanValue(reader)
		case HProfValueType_BYTE:
			fv, err = readByteValue(reader)
		case HProfValueType_CHAR:
			fv, err = readCharValue(reader)
		case HProfValueType_DOUBLE:
			fv, err = readDoubleValue(reader)
		case HProfValueType_FLOAT:
			fv, err = readFloatValue(reader)
		case HProfValueType_INT:
			fv, err = readIntValue(reader)
		case HProfValueType_LONG:
			fv, err = readLongValue(reader)
		case HProfValueType_OBJECT:
			fv, err = readObjectValue(reader)
		case HProfValueType_SHORT:
			fv, err = readShortValue(reader)
		}
		if err != nil {
			return nil, err
		}
		result = append(result, fv)
	}
	return result, nil
}

func ReadHProfInstanceRecord(pr *HProfReader) (*HProfInstanceRecord, error) {
	pos := pr.pos
	oid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	stsn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	coid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	fsz, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	bs, err := pr.readBytes(int(fsz))
	if err != nil {
		return nil, err
	}
	size := int(pr.pos - pos)
	return &HProfInstanceRecord{
		HProfBasicRecord:       HProfBasicRecord{pos, size},
		ObjectId:               oid,
		StackTraceSerialNumber: stsn,
		ClassObjectId:          coid,
		Values:                 bs,
	}, nil
}

func ReadHProfInstanceRecordWithPos(pr *HProfReader, pos int64) (*HProfInstanceRecord, error) {
	err := pr.Seek(pos)
	if err != nil {
		return nil, err
	}
	return ReadHProfInstanceRecord(pr)
}
