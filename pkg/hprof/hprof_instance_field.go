package hprof

import (
	"bytes"
	"encoding/binary"
)

type HProfInstanceFieldValue interface {
	ValueType() HProfValueType
}

// Instance fields.
type HProfInstanceBasicValue struct {
	// Instance field name, associated with HProfRecordUTF8.
	NameId uint64
	// Type of the instance field.
	Type HProfValueType
}

func (b *HProfInstanceBasicValue) ValueType() HProfValueType {
	return b.Type
}

type HProfInstanceArrayValue struct {
	HProfInstanceBasicValue
}

type HProfInstanceBooleanValue struct {
	HProfInstanceBasicValue
	Value bool
}

func readBooleanValue(reader *bytes.Reader) (*HProfInstanceBooleanValue, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	return &HProfInstanceBooleanValue{
		HProfInstanceBasicValue: HProfInstanceBasicValue{
			Type: HProfValueType_BOOLEAN,
		},
		Value: b != 0,
	}, nil
}

type HProfInstanceByteValue struct {
	HProfInstanceBasicValue
	Value byte
}

func readByteValue(reader *bytes.Reader) (*HProfInstanceByteValue, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	return &HProfInstanceByteValue{
		HProfInstanceBasicValue: HProfInstanceBasicValue{
			Type: HProfValueType_BYTE,
		},
		Value: b,
	}, nil
}

type HProfInstanceCharValue struct {
	HProfInstanceBasicValue
	Value uint16
}

func readCharValue(reader *bytes.Reader) (*HProfInstanceCharValue, error) {
	var v uint16
	if err := binary.Read(reader, binary.BigEndian, &v); err != nil {
		return nil, err
	}
	return &HProfInstanceCharValue{
		HProfInstanceBasicValue: HProfInstanceBasicValue{
			Type: HProfValueType_BYTE,
		},
		Value: v,
	}, nil
}

type HProfInstanceDoubleValue struct {
	HProfInstanceBasicValue
	Value float64
}

func readDoubleValue(reader *bytes.Reader) (*HProfInstanceDoubleValue, error) {
	var v float64
	if err := binary.Read(reader, binary.BigEndian, &v); err != nil {
		return nil, err
	}
	return &HProfInstanceDoubleValue{
		HProfInstanceBasicValue: HProfInstanceBasicValue{
			Type: HProfValueType_BYTE,
		},
		Value: v,
	}, nil
}

type HProfInstanceFloatValue struct {
	HProfInstanceBasicValue
	Value float32
}

func readFloatValue(reader *bytes.Reader) (*HProfInstanceFloatValue, error) {
	var v float32
	if err := binary.Read(reader, binary.BigEndian, &v); err != nil {
		return nil, err
	}
	return &HProfInstanceFloatValue{
		HProfInstanceBasicValue: HProfInstanceBasicValue{
			Type: HProfValueType_BYTE,
		},
		Value: v,
	}, nil
}

type HProfInstanceIntValue struct {
	HProfInstanceBasicValue
	Value int32
}

func readIntValue(reader *bytes.Reader) (*HProfInstanceIntValue, error) {
	var v int32
	if err := binary.Read(reader, binary.BigEndian, &v); err != nil {
		return nil, err
	}
	return &HProfInstanceIntValue{
		HProfInstanceBasicValue: HProfInstanceBasicValue{
			Type: HProfValueType_BYTE,
		},
		Value: v,
	}, nil
}

type HProfInstanceLongValue struct {
	HProfInstanceBasicValue
	Value int64
}

func readLongValue(reader *bytes.Reader) (*HProfInstanceLongValue, error) {
	var v int64
	if err := binary.Read(reader, binary.BigEndian, &v); err != nil {
		return nil, err
	}
	return &HProfInstanceLongValue{
		HProfInstanceBasicValue: HProfInstanceBasicValue{
			Type: HProfValueType_BYTE,
		},
		Value: v,
	}, nil
}

type HProfInstanceObjectValue struct {
	HProfInstanceBasicValue
	Value uint64
}

func readObjectValue(reader *bytes.Reader) (*HProfInstanceObjectValue, error) {
	var v uint64
	if err := binary.Read(reader, binary.BigEndian, &v); err != nil {
		return nil, err
	}
	return &HProfInstanceObjectValue{
		HProfInstanceBasicValue: HProfInstanceBasicValue{
			Type: HProfValueType_BYTE,
		},
		Value: v,
	}, nil
}

type HProfInstanceShortValue struct {
	HProfInstanceBasicValue
	Value int16
}

func readShortValue(reader *bytes.Reader) (*HProfInstanceShortValue, error) {
	var v int16
	if err := binary.Read(reader, binary.BigEndian, &v); err != nil {
		return nil, err
	}
	return &HProfInstanceShortValue{
		HProfInstanceBasicValue: HProfInstanceBasicValue{
			Type: HProfValueType_BYTE,
		},
		Value: v,
	}, nil
}

type HProfInstanceUnknownValue struct {
	HProfInstanceBasicValue
}
