package hprof

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"hprof-tool/pkg/model"
	"io"
	"os"
	"time"
)

var (
	// ValueSize is a size of the HProf values.
	ValueSize = map[HProfValueType]int{
		HProfValueType_OBJECT:  -1,
		HProfValueType_BOOLEAN: 1,
		HProfValueType_CHAR:    2,
		HProfValueType_FLOAT:   4,
		HProfValueType_DOUBLE:  8,
		HProfValueType_BYTE:    1,
		HProfValueType_SHORT:   2,
		HProfValueType_INT:     4,
		HProfValueType_LONG:    8,
	}
)

// HProfHeader is a HProf file header.
type HProfHeader struct {
	// Magic string.
	Header string
	// The size of object IDs.
	IdentifierSize uint32
	// Dump creation time.
	Timestamp time.Time
}

// HProfReader is a HProf file reader.
type HProfReader struct {
	hFile                  *os.File
	reader                 *bufio.Reader
	pos                    int64
	identifierSize         int
	heapDumpFrameLeftBytes uint32

	Header *HProfHeader
}

// NewParser creates a new HProf parser.
func NewReader(r *os.File) *HProfReader {
	return &HProfReader{
		hFile:  r,
		reader: bufio.NewReader(r),
	}
}

// Seek 移动文件位置，读取某段数据
func (p *HProfReader) Seek(offset int64) error {
	pos, err := p.hFile.Seek(offset, 0)
	if err != nil {
		return err
	}
	p.pos = pos
	p.reader.Reset(p.hFile)
	return nil
}

func (p *HProfReader) IdSize() byte {
	var ret byte = 4
	if p.identifierSize > 4 {
		ret = 8
	}
	return ret
}

// ParseHeader parses the HProf header.
func (p *HProfReader) ParseHeader() error {
	bs, err := p.reader.ReadSlice(0x00)
	if err != nil {
		return err
	}
	p.pos += int64(len(bs))

	is, err := p.readUint32()
	if err != nil {
		return err
	}
	p.identifierSize = int(is)

	tsHigh, err := p.readUint32()
	if err != nil {
		return err
	}
	tsLow, err := p.readUint32()
	if err != nil {
		return err
	}
	var tsMilli int64 = int64(tsHigh)
	tsMilli <<= 32
	tsMilli += int64(tsLow)

	p.Header = &HProfHeader{
		Header:         string(bs),
		IdentifierSize: is,
		Timestamp:      time.Unix(0, 0).Add(time.Duration(tsMilli * int64(time.Millisecond))),
	}
	return nil
}

func (p *HProfReader) ParseRecord() (HProfRecord, error) {
	if p.heapDumpFrameLeftBytes > 0 {
		return p.parseHeapDumpFrame()
	}

	rt, err := p.parseType()
	if err != nil {
		return nil, err
	}

	switch HProfRecordType(rt) {
	case HProfRecordTypeUTF8:
		return ReadHProfUTF8Record(p)
	case model.HProfRecordTypeLoadClass:
		return ReadHProfLoadClassRecord(p)
	case model.HProfRecordTypeFrame:
		return ReadHProfFrameRecord(p)
	case model.HProfRecordTypeTrace:
		return ReadHProfTraceRecord(p)
	case model.HProfRecordTypeStartThread:
		return ReadHProfThreadRecord(p)
	case model.HProfRecordTypeHeapDumpSegment:
		return parseHeapDumpSegment(p)
	case model.HProfRecordTypeHeapDumpEnd:
		return parseHeapDumpSegmentEnd(p)
	case model.HProfRecordTypeHeapDump:
		return parseHeapDump(p)
	default:
		fmt.Printf("unknown record type: 0x%x\n", rt)
		sz, err := p.parseRecordSize()
		if err != nil {
			return nil, err
		}
		_, err = p.readBytes(int(sz))
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("unknown record type: 0x%x", rt)
	}
}

func (p *HProfReader) parseHeapDumpFrame() (HProfRecord, error) {
	rt, err := p.readByte()
	if err != nil {
		return nil, err
	}

	switch model.HProfHDRecordType(rt) {
	case model.HProfHDRecordTypeRootJNIGlobal:
		return ReadHProfRootJNIGlobal(p)
	case model.HProfHDRecordTypeRootJNILocal:
		return ReadHProfRootJNILocal(p)
	case model.HProfHDRecordTypeRootJavaFrame:
		return ReadHProfRootJavaFrame(p)
	case model.HProfHDRecordTypeRootStickyClass:
		return ReadHProfRootStickyClass(p)
	case model.HProfHDRecordTypeRootThreadObj:
		return ReadHProfRootThreadObj(p)
	case model.HProfHDRecordTypeRootMonitorUsed:
		return ReadHProfRootMonitorUsed(p)
	case model.HProfHDRecordTypeClassDump:
		return ReadHProfClassRecord(p)
	case model.HProfHDRecordTypeInstanceDump:
		return ReadHProfInstanceRecord(p)
	case model.HProfHDRecordTypeObjectArrayDump:
		return ReadHProfObjectArrayRecord(p)
	case model.HProfHDRecordTypePrimitiveArrayDump:
		return ReadHProfPrimitiveArrayRecord(p)
	default:
		return nil, fmt.Errorf("unknown heap dump record type: 0x%x", rt)
	}
}

func (p *HProfReader) parseType() (byte, error) {
	p.pos += 1
	return p.reader.ReadByte()
}

func (p *HProfReader) parseRecordSize() (uint32, error) {
	_, err := p.readUint32()
	if err != nil {
		return 0, err
	}

	return p.readUint32()
}

func (p *HProfReader) readByte() (byte, error) {
	b, err := p.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	p.pos += 1
	if p.heapDumpFrameLeftBytes > 0 {
		p.heapDumpFrameLeftBytes--
	}
	return b, nil
}

func (p *HProfReader) readID() (uint64, error) {
	var v uint64
	if p.identifierSize == 8 {
		if err := binary.Read(p.reader, binary.BigEndian, &v); err != nil {
			return 0, err
		}
		p.pos += 8
	} else if p.identifierSize == 4 {
		var v2 uint32
		if err := binary.Read(p.reader, binary.BigEndian, &v2); err != nil {
			return 0, err
		}
		v = uint64(v2)
		p.pos += 4
	} else {
		return 0, fmt.Errorf("odd identifier size: %d", p.identifierSize)
	}
	if p.heapDumpFrameLeftBytes > 0 {
		p.heapDumpFrameLeftBytes -= uint32(p.identifierSize)
	}
	return v, nil
}

func (p *HProfReader) readBytes(n int) ([]byte, error) {
	bs := make([]byte, n)
	rn, err := io.ReadFull(p.reader, bs)
	p.pos += int64(rn)
	if err != nil {
		return nil, err
	}
	if p.heapDumpFrameLeftBytes > 0 {
		p.heapDumpFrameLeftBytes -= uint32(len(bs))
	}
	return bs, nil
}

func (p *HProfReader) readArray(ty HProfValueType, n int) ([]byte, error) {
	sz := ValueSize[ty]
	if sz == -1 {
		sz = p.identifierSize
	}
	if sz == 0 {
		return nil, fmt.Errorf("odd value type: %d", ty)
	}

	bs := make([]byte, int(sz)*n)
	rn, err := io.ReadFull(p.reader, bs)
	p.pos += int64(rn)
	if err != nil {
		return nil, err
	}
	if p.heapDumpFrameLeftBytes > 0 {
		p.heapDumpFrameLeftBytes -= uint32(len(bs))
	}
	return bs, nil
}

func (p *HProfReader) readValue(ty HProfValueType) (uint64, error) {
	sz := ValueSize[ty]
	if sz == -1 {
		sz = p.identifierSize
	}
	if sz == 0 {
		return 0, fmt.Errorf("odd value type: %d", ty)
	}

	bs := make([]byte, 8)
	n, err := io.ReadFull(p.reader, bs[:int(sz)])
	p.pos += int64(n)
	if err != nil {
		return 0, err
	}
	if p.heapDumpFrameLeftBytes > 0 {
		p.heapDumpFrameLeftBytes -= uint32(sz)
	}
	return binary.BigEndian.Uint64(bs), nil
}

func (p *HProfReader) readUint16() (uint16, error) {
	var v uint16
	if err := binary.Read(p.reader, binary.BigEndian, &v); err != nil {
		return 0, err
	}
	if p.heapDumpFrameLeftBytes > 0 {
		p.heapDumpFrameLeftBytes -= 2
	}
	p.pos += 2
	return v, nil
}

func (p *HProfReader) readUint32() (uint32, error) {
	var v uint32
	if err := binary.Read(p.reader, binary.BigEndian, &v); err != nil {
		return 0, err
	}
	if p.heapDumpFrameLeftBytes > 0 {
		p.heapDumpFrameLeftBytes -= 4
	}
	p.pos += 4
	return v, nil
}

func (p *HProfReader) readInt32() (int32, error) {
	var v int32
	if err := binary.Read(p.reader, binary.BigEndian, &v); err != nil {
		return 0, err
	}
	if p.heapDumpFrameLeftBytes > 0 {
		p.heapDumpFrameLeftBytes -= 4
	}
	p.pos += 4
	return v, nil
}

func (p *HProfReader) ValueSize(typ HProfValueType) int {
	return ValueSize[typ]
}
