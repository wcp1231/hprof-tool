package hprof

type HProfRecord interface {
	PosAndSize() (int64, int)
}

// HProfBasicRecord 共有字段
// Pos 字段位置
// Size 字段长度
type HProfBasicRecord struct {
	Pos  int64
	Size int
}

func (h *HProfBasicRecord) PosAndSize() (int64, int) {
	return h.Pos, h.Size
}

// HProfRecordType is a HProf record type.
type HProfRecordType byte

// HProfHDRecordType is a HProf heap dump subrecord type.
type HProfHDRecordType HProfRecordType

// HProf's record types.
const (
	HProfRecordTypeUTF8            HProfRecordType = 0x01
	HProfRecordTypeLoadClass                       = 0x02
	HProfRecordTypeUnloadClass                     = 0x03
	HProfRecordTypeFrame                           = 0x04
	HProfRecordTypeTrace                           = 0x05
	HProfRecordTypeAllocSites                      = 0x06
	HProfRecordTypeHeapSummary                     = 0x07
	HProfRecordTypeStartThread                     = 0x0a
	HProfRecordTypeEndThread                       = 0x0b
	HProfRecordTypeHeapDump                        = 0x0c
	HProfRecordTypeHeapDumpSegment                 = 0x1c
	HProfRecordTypeHeapDumpEnd                     = 0x2c
	HProfRecordTypeCPUSamples                      = 0x0d
	HProfRecordTypeControlSettings                 = 0x0e

	HProfHDRecordTypeRootUnknown     HProfHDRecordType = 0xff
	HProfHDRecordTypeRootJNIGlobal                     = 0x01
	HProfHDRecordTypeRootJNILocal                      = 0x02
	HProfHDRecordTypeRootJavaFrame                     = 0x03
	HProfHDRecordTypeRootNativeStack                   = 0x04
	HProfHDRecordTypeRootStickyClass                   = 0x05
	HProfHDRecordTypeRootThreadBlock                   = 0x06
	HProfHDRecordTypeRootMonitorUsed                   = 0x07
	HProfHDRecordTypeRootThreadObj                     = 0x08

	HProfHDRecordTypeClassDump          = 0x20
	HProfHDRecordTypeInstanceDump       = 0x21
	HProfHDRecordTypeObjectArrayDump    = 0x22
	HProfHDRecordTypePrimitiveArrayDump = 0x23
)

type HProfValueType int32

const (
	HProfValueType_UNKNOWN_HPROF_VALUE_TYPE HProfValueType = 0
	// Object. The value of this type is an object_id of HProfInstanceDump,
	// array_object_id of HProfObjectArrayDump or HProfPrimitiveArrayDump,
	// or class_object_id of HProfClassDump.
	//
	// The value is basically a pointer and its size is defined in the hprof
	// header, which is typically 4 bytes for 32-bit JVM hprof dumps or 8 bytes
	// for 64-bit JVM hprof dumps.
	HProfValueType_OBJECT HProfValueType = 2
	// Boolean. Takes 0 or 1. One byte.
	HProfValueType_BOOLEAN HProfValueType = 4
	// Character. Two bytes.
	HProfValueType_CHAR HProfValueType = 5
	// Float. 4 bytes
	HProfValueType_FLOAT HProfValueType = 6
	// Double. 8 bytes.
	HProfValueType_DOUBLE HProfValueType = 7
	// Byte. One byte.
	HProfValueType_BYTE HProfValueType = 8
	// Short. Two bytes.
	HProfValueType_SHORT HProfValueType = 9
	// Integer. 4 bytes.
	HProfValueType_INT HProfValueType = 10
	// Long. 8 bytes.
	HProfValueType_LONG HProfValueType = 11
)

var HProfValueType_name = map[HProfValueType]string{
	0:  "UNKNOWN_HPROF_VALUE_TYPE",
	2:  "OBJECT",
	4:  "BOOLEAN",
	5:  "CHAR",
	6:  "FLOAT",
	7:  "DOUBLE",
	8:  "BYTE",
	9:  "SHORT",
	10: "INT",
	11: "LONG",
}

var HProfValueType_value = map[string]int32{
	"UNKNOWN_HPROF_VALUE_TYPE": 0,
	"OBJECT":                   2,
	"BOOLEAN":                  4,
	"CHAR":                     5,
	"FLOAT":                    6,
	"DOUBLE":                   7,
	"BYTE":                     8,
	"SHORT":                    9,
	"INT":                      10,
	"LONG":                     11,
}
