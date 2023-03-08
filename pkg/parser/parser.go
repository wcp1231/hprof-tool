package parser

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"hprof-tool/pkg/model"
	"io"
	"math"
	"os"
	"time"
)

var (
	// ValueSize is a size of the HProf values.
	ValueSize = map[model.HProfValueType]int{
		model.HProfValueType_OBJECT:  -1,
		model.HProfValueType_BOOLEAN: 1,
		model.HProfValueType_CHAR:    2,
		model.HProfValueType_FLOAT:   4,
		model.HProfValueType_DOUBLE:  8,
		model.HProfValueType_BYTE:    1,
		model.HProfValueType_SHORT:   2,
		model.HProfValueType_INT:     4,
		model.HProfValueType_LONG:    8,
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

// HProfParser is a HProf file parser.
type HProfParser struct {
	hFile                  *os.File
	reader                 *bufio.Reader
	pos                    int64
	identifierSize         int
	heapDumpFrameLeftBytes uint32
}

// NewParser creates a new HProf parser.
func NewParser(r *os.File) *HProfParser {
	return &HProfParser{
		hFile:  r,
		reader: bufio.NewReader(r),
	}
}

// Seek 移动文件位置，读取某段数据
func (p *HProfParser) Seek(offset int64) error {
	pos, err := p.hFile.Seek(offset, 0)
	if err != nil {
		return err
	}
	p.pos = pos
	p.reader.Reset(p.hFile)
	return nil
}

func (p *HProfParser) IdSize() byte {
	var ret byte = 4
	if p.identifierSize > 4 {
		ret = 8
	}
	return ret
}

// ParseHeader parses the HProf header.
func (p *HProfParser) ParseHeader() (*HProfHeader, error) {
	bs, err := p.reader.ReadSlice(0x00)
	if err != nil {
		return nil, err
	}
	p.pos += int64(len(bs))

	is, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	p.identifierSize = int(is)

	tsHigh, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	tsLow, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	var tsMilli int64 = int64(tsHigh)
	tsMilli <<= 32
	tsMilli += int64(tsLow)

	return &HProfHeader{
		Header:         string(bs),
		IdentifierSize: is,
		Timestamp:      time.Unix(0, 0).Add(time.Duration(tsMilli * int64(time.Millisecond))),
	}, nil
}

// ParseRecord returns the next HProf record.
//
// HProf file consists of sequence of records. Heapdump records and heapdump
// segement records contains subrecords inside. This method parses out those
// recordss and subrecords and returns one record for each. The returned value
// is one of the followings:
//
// *   `*hprofdata.HProfRecordUTF8`
// *   `*hprofdata.HProfRecordLoadClass`
// *   `*hprofdata.HProfRecordFrame`
// *   `*hprofdata.HProfRecordTrace`
// *   `*hprofdata.HProfRecordHeapDumpBoundary`
// *   `*hprofdata.HProfClassDump`
// *   `*hprofdata.HProfInstanceDump`
// *   `*hprofdata.HProfObjectArrayDump`
// *   `*hprofdata.HProfPrimitiveArrayDump`
// *   `*hprofdata.HProfRootJNIGlobal`
// *   `*hprofdata.HProfRootJNILocal`
// *   `*hprofdata.HProfRootJavaFrame`
// *   `*hprofdata.HProfRootStickyClass`
// *   `*hprofdata.HProfRootThreadObj`
//
// It returns io.EOF at the end of the file.
func (p *HProfParser) ParseRecord() (model.HProfRecord, error) {
	if p.heapDumpFrameLeftBytes > 0 {
		return p.parseHeapDumpFrame()
	}

	rt, err := p.parseType()
	if err != nil {
		return nil, err
	}

	switch model.HProfRecordType(rt) {
	case model.HProfRecordTypeUTF8:
		return p.parseUtf8Record()
	case model.HProfRecordTypeLoadClass:
		return p.parseLoadedClassRecord()
	case model.HProfRecordTypeFrame:
		return p.parseFrameRecord()
	case model.HProfRecordTypeTrace:
		return p.parseTraceRecord()
	case model.HProfRecordTypeHeapDumpSegment:
		return p.parseHeapDumpSegment()
	case model.HProfRecordTypeHeapDumpEnd:
		return p.parseHeapDumpSegmentEnd()
	case model.HProfRecordTypeHeapDump:
		return p.parseHeapDump()
	default:
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

func (p *HProfParser) parseHeapDumpFrame() (model.HProfRecord, error) {
	rt, err := p.readByte()
	if err != nil {
		return nil, err
	}

	switch model.HProfHDRecordType(rt) {
	case model.HProfHDRecordTypeRootJNIGlobal:
		return p.parseRootJNIGlobalRecord()
	case model.HProfHDRecordTypeRootJNILocal:
		return p.parseRootJNILocalRecord()
	case model.HProfHDRecordTypeRootJavaFrame:
		return p.parseRootJavaFrameRecord()
	case model.HProfHDRecordTypeRootStickyClass:
		return p.parseRootStickyClassRecord()
	case model.HProfHDRecordTypeRootThreadObj:
		return p.parseRootThreadObjRecord()
	case model.HProfHDRecordTypeRootMonitorUsed:
		return p.ParseRootMonitorUsedRecord()
	case model.HProfHDRecordTypeClassDump:
		return p.ParseClassDumpRecord()
	case model.HProfHDRecordTypeInstanceDump:
		return p.ParseInstanceRecord()
	case model.HProfHDRecordTypeObjectArrayDump:
		return p.ParseObjectArrayRecord()
	case model.HProfHDRecordTypePrimitiveArrayDump:
		return p.ParsePrimitiveArrayDumpRecord()
	default:
		return nil, fmt.Errorf("unknown heap dump record type: 0x%x", rt)
	}
}

func (p *HProfParser) parseType() (byte, error) {
	p.pos += 1
	return p.reader.ReadByte()
}

func (p *HProfParser) parseRecordSize() (uint32, error) {
	_, err := p.readUint32()
	if err != nil {
		return 0, err
	}

	return p.readUint32()
}

func (p *HProfParser) parseUtf8Record() (*model.HProfRecordUTF8, error) {
	pos := p.pos
	sz, err := p.parseRecordSize()
	if err != nil {
		return nil, err
	}
	nameID, err := p.readID()
	if err != nil {
		return nil, err
	}
	bs := make([]byte, int(sz)-p.identifierSize)
	rn, err := io.ReadFull(p.reader, bs)
	if err != nil {
		return nil, err
	}
	p.pos += int64(rn)

	return &model.HProfRecordUTF8{
		NameId: nameID,
		Name:   bs,

		POS: pos,
	}, nil
}

func (p *HProfParser) parseLoadedClassRecord() (*model.HProfRecordLoadClass, error) {
	pos := p.pos
	_, err := p.parseRecordSize()
	if err != nil {
		return nil, err
	}

	csn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	oid, err := p.readID()
	if err != nil {
		return nil, err
	}
	tsn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	cnid, err := p.readID()
	if err != nil {
		return nil, err
	}
	return &model.HProfRecordLoadClass{
		ClassSerialNumber:      csn,
		ClassObjectId:          oid,
		StackTraceSerialNumber: tsn,
		ClassNameId:            cnid,

		POS: pos,
	}, nil
}

func (p *HProfParser) parseFrameRecord() (*model.HProfRecordFrame, error) {
	pos := p.pos
	_, err := p.parseRecordSize()
	if err != nil {
		return nil, err
	}

	sfid, err := p.readID()
	if err != nil {
		return nil, err
	}
	mnid, err := p.readID()
	if err != nil {
		return nil, err
	}
	msgnid, err := p.readID()
	if err != nil {
		return nil, err
	}
	sfnid, err := p.readID()
	if err != nil {
		return nil, err
	}
	csn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	ln, err := p.readInt32()
	if err != nil {
		return nil, err
	}
	return &model.HProfRecordFrame{
		StackFrameId:      sfid,
		MethodNameId:      mnid,
		MethodSignatureId: msgnid,
		SourceFileNameId:  sfnid,
		ClassSerialNumber: csn,
		LineNumber:        ln,

		POS: pos,
	}, nil
}

func (p *HProfParser) parseTraceRecord() (*model.HProfRecordTrace, error) {
	pos := p.pos
	_, err := p.parseRecordSize()
	if err != nil {
		return nil, err
	}

	stsn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	tsn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	nr, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	sfids := []uint64{}
	for i := uint32(0); i < nr; i++ {
		sfid, err := p.readID()
		if err != nil {
			return nil, err
		}
		sfids = append(sfids, sfid)
	}
	return &model.HProfRecordTrace{
		StackTraceSerialNumber: stsn,
		ThreadSerialNumber:     tsn,
		StackFrameIds:          sfids,

		POS: pos,
	}, nil
}

func (p *HProfParser) parseHeapDumpSegment() (*model.HProfRecordHeapDumpBoundary, error) {
	sz, err := p.parseRecordSize()
	if err != nil {
		return nil, err
	}

	if sz == 0 {
		// Truncated. Set to the max int.
		sz = math.MaxUint32
	}
	p.heapDumpFrameLeftBytes = sz
	return &model.HProfRecordHeapDumpBoundary{}, nil
}

func (p *HProfParser) parseHeapDumpSegmentEnd() (*model.HProfRecordHeapDumpBoundary, error) {
	_, err := p.parseRecordSize()
	if err != nil {
		return nil, err
	}
	return &model.HProfRecordHeapDumpBoundary{}, nil
}

func (p *HProfParser) parseHeapDump() (*model.HProfRecordHeapDumpBoundary, error) {
	sz, err := p.parseRecordSize()
	if err != nil {
		return nil, err
	}
	p.heapDumpFrameLeftBytes = sz
	return &model.HProfRecordHeapDumpBoundary{}, nil
}

func (p *HProfParser) parseRootJNIGlobalRecord() (*model.HProfRootJNIGlobal, error) {
	pos := p.pos
	oid, err := p.readID()
	if err != nil {
		return nil, err
	}
	rid, err := p.readID()
	if err != nil {
		return nil, err
	}
	return &model.HProfRootJNIGlobal{
		ObjectId:       oid,
		JniGlobalRefId: rid,

		POS: pos,
	}, nil
}

func (p *HProfParser) parseRootJNILocalRecord() (*model.HProfRootJNILocal, error) {
	pos := p.pos
	oid, err := p.readID()
	if err != nil {
		return nil, err
	}
	tsn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	fn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	return &model.HProfRootJNILocal{
		ObjectId:                oid,
		ThreadSerialNumber:      tsn,
		FrameNumberInStackTrace: fn,

		POS: pos,
	}, nil
}

func (p *HProfParser) parseRootJavaFrameRecord() (*model.HProfRootJavaFrame, error) {
	pos := p.pos
	oid, err := p.readID()
	if err != nil {
		return nil, err
	}
	tsn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	fn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	return &model.HProfRootJavaFrame{
		ObjectId:                oid,
		ThreadSerialNumber:      tsn,
		FrameNumberInStackTrace: fn,

		POS: pos,
	}, nil
}

func (p *HProfParser) parseRootStickyClassRecord() (*model.HProfRootStickyClass, error) {
	pos := p.pos
	oid, err := p.readID()
	if err != nil {
		return nil, err
	}
	return &model.HProfRootStickyClass{
		ObjectId: oid,

		POS: pos,
	}, nil
}

func (p *HProfParser) parseRootThreadObjRecord() (*model.HProfRootThreadObj, error) {
	pos := p.pos
	toid, err := p.readID()
	if err != nil {
		return nil, err
	}
	tsn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	stsn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	return &model.HProfRootThreadObj{
		ThreadObjectId:           toid,
		ThreadSequenceNumber:     tsn,
		StackTraceSequenceNumber: stsn,

		POS: pos,
	}, nil
}

func (p *HProfParser) ParseRootMonitorUsedRecord() (*model.HProfRootMonitorUsed, error) {
	oid, err := p.readID()
	if err != nil {
		return nil, err
	}
	return &model.HProfRootMonitorUsed{
		ObjectId: oid,
	}, nil
}

func (p *HProfParser) ParseClassDumpRecord() (*model.HProfClassDump, error) {
	pos := p.pos
	coid, err := p.readID()
	if err != nil {
		return nil, err
	}
	stsn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	scoid, err := p.readID()
	if err != nil {
		return nil, err
	}
	cloid, err := p.readID()
	if err != nil {
		return nil, err
	}
	sgnoid, err := p.readID()
	if err != nil {
		return nil, err
	}
	pdoid, err := p.readID()
	if err != nil {
		return nil, err
	}

	_, err = p.readID()
	if err != nil {
		return nil, err
	}
	_, err = p.readID()
	if err != nil {
		return nil, err
	}
	insz, err := p.readUint32()
	if err != nil {
		return nil, err
	}

	cpsz, err := p.readUint16()
	if err != nil {
		return nil, err
	}
	cps := []*model.HProfClassDump_ConstantPoolEntry{}
	for i := uint16(0); i < cpsz; i++ {
		cpix, err := p.readUint16()
		if err != nil {
			return nil, err
		}
		ty, err := p.readByte()
		if err != nil {
			return nil, err
		}
		v, err := p.readValue(model.HProfValueType(ty))
		if err != nil {
			return nil, err
		}
		cps = append(cps, &model.HProfClassDump_ConstantPoolEntry{
			ConstantPoolIndex: uint32(cpix),
			Type:              model.HProfValueType(ty),
			Value:             v,
		})
	}

	sfsz, err := p.readUint16()
	if err != nil {
		return nil, err
	}
	sfs := []*model.HProfClassDump_StaticField{}
	for i := uint16(0); i < sfsz; i++ {
		sfnid, err := p.readID()
		if err != nil {
			return nil, err
		}
		ty, err := p.readByte()
		if err != nil {
			return nil, err
		}
		v, err := p.readValue(model.HProfValueType(ty))
		if err != nil {
			return nil, err
		}
		sfs = append(sfs, &model.HProfClassDump_StaticField{
			NameId: sfnid,
			Type:   model.HProfValueType(ty),
			Value:  v,
		})
	}

	ifsz, err := p.readUint16()
	if err != nil {
		return nil, err
	}
	ifs := []*model.HProfClassDump_InstanceField{}
	for i := uint16(0); i < ifsz; i++ {
		ifnid, err := p.readID()
		if err != nil {
			return nil, err
		}
		ty, err := p.readByte()
		if err != nil {
			return nil, err
		}
		ifs = append(ifs, &model.HProfClassDump_InstanceField{
			NameId: ifnid,
			Type:   model.HProfValueType(ty),
		})
	}

	return &model.HProfClassDump{
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

		POS: pos,
	}, nil
}

func (p *HProfParser) ParseInstanceRecord() (*model.HProfInstanceDump, error) {
	pos := p.pos
	oid, err := p.readID()
	if err != nil {
		return nil, err
	}
	stsn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	coid, err := p.readID()
	if err != nil {
		return nil, err
	}
	fsz, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	bs, err := p.readBytes(int(fsz))
	if err != nil {
		return nil, err
	}
	return &model.HProfInstanceDump{
		ObjectId:               oid,
		StackTraceSerialNumber: stsn,
		ClassObjectId:          coid,
		Values:                 bs,

		POS: pos,
	}, nil
}

func (p *HProfParser) ParseObjectArrayRecord() (*model.HProfObjectArrayDump, error) {
	pos := p.pos
	aoid, err := p.readID()
	if err != nil {
		return nil, err
	}
	stsn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	asz, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	acoid, err := p.readID()
	if err != nil {
		return nil, err
	}
	vs := []uint64{}
	for i := uint32(0); i < asz; i++ {
		v, err := p.readID()
		if err != nil {
			return nil, err
		}
		vs = append(vs, v)
	}
	return &model.HProfObjectArrayDump{
		ArrayObjectId:          aoid,
		StackTraceSerialNumber: stsn,
		ArrayClassObjectId:     acoid,
		ElementObjectIds:       vs,

		POS: pos,
	}, nil
}

func (p *HProfParser) ParsePrimitiveArrayDumpRecord() (*model.HProfPrimitiveArrayDump, error) {
	pos := p.pos
	aoid, err := p.readID()
	if err != nil {
		return nil, err
	}
	stsn, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	asz, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	ty, err := p.readByte()
	if err != nil {
		return nil, err
	}
	bs, err := p.readArray(model.HProfValueType(ty), int(asz))
	if err != nil {
		return nil, err
	}
	return &model.HProfPrimitiveArrayDump{
		ArrayObjectId:          aoid,
		StackTraceSerialNumber: stsn,
		ElementType:            model.HProfValueType(ty),
		Values:                 bs,

		POS: pos,
	}, nil
}

func (p *HProfParser) readByte() (byte, error) {
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

func (p *HProfParser) readID() (uint64, error) {
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

func (p *HProfParser) readBytes(n int) ([]byte, error) {
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

func (p *HProfParser) readArray(ty model.HProfValueType, n int) ([]byte, error) {
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

func (p *HProfParser) readValue(ty model.HProfValueType) (uint64, error) {
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

func (p *HProfParser) readUint16() (uint16, error) {
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

func (p *HProfParser) readUint32() (uint32, error) {
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

func (p *HProfParser) readInt32() (int32, error) {
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

func (p *HProfParser) Pos() int64 {
	return p.pos
}

func (p *HProfParser) ValueSize(typ model.HProfValueType) int {
	return ValueSize[typ]
}

// ReadID 临时用来读 instance value
func (p *HProfParser) ReadIDFromReader(idSize int, reader io.Reader) (uint64, error) {
	var v uint64
	if idSize == 8 {
		if err := binary.Read(reader, binary.BigEndian, &v); err != nil {
			return 0, err
		}
	} else if idSize == 4 {
		var v2 uint32
		if err := binary.Read(reader, binary.BigEndian, &v2); err != nil {
			return 0, err
		}
		v = uint64(v2)
	} else {
		return 0, fmt.Errorf("odd identifier size: %d", p.identifierSize)
	}
	return v, nil
}

// ReadValueFromReader 临时用来读 instance value
func (p *HProfParser) ReadValueFromReader(ty model.HProfValueType, reader io.Reader) (uint64, error) {
	sz := ValueSize[ty]
	if sz == -1 {
		sz = p.identifierSize
	}
	if sz == 0 {
		return 0, fmt.Errorf("odd value type: %d", ty)
	}
	bs := make([]byte, 8)
	_, err := io.ReadFull(reader, bs[:int(sz)])
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(bs), nil
}
