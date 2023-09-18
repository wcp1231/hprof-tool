package hprof

import (
	"io"
)

// UTF-8 byte sequence record.
//
// Even though it says UTF-8, its content might not be a valid UTF-8 sequence.
type HProfUTF8Record struct {
	HProfBasicRecord

	NameId uint64
	Name   []byte
}

func (m *HProfUTF8Record) Id() uint64 {
	if m != nil {
		return m.NameId
	}
	return 0
}

func (m *HProfUTF8Record) Type() HProfRecordType {
	return HProfRecordTypeUTF8
}

func ReadHProfUTF8Record(pr *HProfReader) (*HProfUTF8Record, error) {
	pos := pr.pos
	sz, err := pr.parseRecordSize()
	if err != nil {
		return nil, err
	}
	nameID, err := pr.readID()
	if err != nil {
		return nil, err
	}
	bs := make([]byte, int(sz)-pr.identifierSize)
	rn, err := io.ReadFull(pr.reader, bs)
	if err != nil {
		return nil, err
	}
	pr.pos += int64(rn)
	size := int(pr.pos - pos)
	return &HProfUTF8Record{
		HProfBasicRecord: HProfBasicRecord{pos, size},

		NameId: nameID,
		Name:   bs,
	}, nil
}

func ReadHProfUTF8RecordWithPos(pr *HProfReader, pos int64) (*HProfUTF8Record, error) {
	err := pr.Seek(pos)
	if err != nil {
		return nil, err
	}
	return ReadHProfUTF8Record(pr)
}
