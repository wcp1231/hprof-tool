package model

import (
	global "hprof-tool/pkg/util"
)

type HProfRecord interface {
	Id() uint64
	Type() HProfRecordType
	Pos() int64
}

type HProfObjectRecord interface {
	HProfRecord
	ClassId() uint64
}

// UTF-8 byte sequence record.
//
// Even though it says UTF-8, its content might not be a valid UTF-8 sequence.
type HProfRecordUTF8 struct {
	NameId uint64 `json:"name_id,omitempty"`
	Name   []byte `json:"name,omitempty"`

	POS int64
}

func (m *HProfRecordUTF8) Id() uint64 {
	if m != nil {
		return m.NameId
	}
	return 0
}

func (m *HProfRecordUTF8) Type() HProfRecordType {
	return HProfRecordTypeUTF8
}

func (m *HProfRecordUTF8) Pos() int64 {
	if m != nil {
		return m.POS
	}
	return 0
}

func (m *HProfRecordUTF8) GetName() []byte {
	if m != nil {
		return m.Name
	}
	return nil
}

// Load class record.
type HProfRecordLoadClass struct {
	// Class serial number.
	ClassSerialNumber uint32 `json:"class_serial_number,omitempty"`
	// Class object ID, associated with HProfClassDump.
	ClassObjectId uint64 `json:"class_object_id,omitempty"`
	// Stack trace serial number. Mostly unused unless the class is dynamically
	// created and loaded with a custom class loader?
	StackTraceSerialNumber uint32 `json:"stack_trace_serial_number,omitempty"`
	// Class name, associated with HProfRecordUTF8.
	ClassNameId uint64 `json:"class_name_id,omitempty"`

	POS int64
}

func (m *HProfRecordLoadClass) Id() uint64 {
	if m != nil {
		return uint64(m.ClassSerialNumber)
	}
	return 0
}

func (m *HProfRecordLoadClass) Type() HProfRecordType {
	return HProfRecordTypeLoadClass
}

func (m *HProfRecordLoadClass) Pos() int64 {
	if m != nil {
		return m.POS
	}
	return 0
}

func (m *HProfRecordLoadClass) GetClassSerialNumber() uint32 {
	if m != nil {
		return m.ClassSerialNumber
	}
	return 0
}

func (m *HProfRecordLoadClass) GetClassObjectId() uint64 {
	if m != nil {
		return m.ClassObjectId
	}
	return 0
}

func (m *HProfRecordLoadClass) GetStackTraceSerialNumber() uint32 {
	if m != nil {
		return m.StackTraceSerialNumber
	}
	return 0
}

func (m *HProfRecordLoadClass) GetClassNameId() uint64 {
	if m != nil {
		return m.ClassNameId
	}
	return 0
}

// Stack frame record.
type HProfRecordFrame struct {
	// Stack frame ID.
	StackFrameId uint64 `json:"stack_frame_id,omitempty"`
	// Method name, associated with HProfRecordUTF8.
	MethodNameId uint64 `json:"method_name_id,omitempty"`
	// Method signature, associated with HProfRecordUTF8.
	MethodSignatureId uint64 `json:"method_signature_id,omitempty"`
	// Source file name, associated with HProfRecordUTF8.
	SourceFileNameId uint64 `json:"source_file_name_id,omitempty"`
	// Class serial number, associated with HProfRecordLoadClass.
	ClassSerialNumber uint32 `json:"class_serial_number,omitempty"`
	// Line number if available.
	LineNumber int32 `json:"line_number,omitempty"`

	POS int64
}

func (m *HProfRecordFrame) Id() uint64 {
	if m != nil {
		return m.StackFrameId
	}
	return 0
}

func (m *HProfRecordFrame) Type() HProfRecordType {
	return HProfRecordTypeFrame
}

func (m *HProfRecordFrame) Pos() int64 {
	if m != nil {
		return m.POS
	}
	return 0
}

func (m *HProfRecordFrame) GetStackFrameId() uint64 {
	if m != nil {
		return m.StackFrameId
	}
	return 0
}

func (m *HProfRecordFrame) GetMethodNameId() uint64 {
	if m != nil {
		return m.MethodNameId
	}
	return 0
}

func (m *HProfRecordFrame) GetMethodSignatureId() uint64 {
	if m != nil {
		return m.MethodSignatureId
	}
	return 0
}

func (m *HProfRecordFrame) GetSourceFileNameId() uint64 {
	if m != nil {
		return m.SourceFileNameId
	}
	return 0
}

func (m *HProfRecordFrame) GetClassSerialNumber() uint32 {
	if m != nil {
		return m.ClassSerialNumber
	}
	return 0
}

func (m *HProfRecordFrame) GetLineNumber() int32 {
	if m != nil {
		return m.LineNumber
	}
	return 0
}

// Stack trace record.
type HProfRecordTrace struct {
	// Stack trace serial number.
	StackTraceSerialNumber uint32 `json:"stack_trace_serial_number,omitempty"`
	// Thread serial number.
	ThreadSerialNumber uint32 `json:"thread_serial_number,omitempty"`
	// Stack frame IDs, associated with HProfRecordFrame.
	StackFrameIds []uint64 `json:"stack_frame_ids,omitempty"`

	POS int64
}

func (m *HProfRecordTrace) Id() uint64 {
	if m != nil {
		return uint64(m.StackTraceSerialNumber)
	}
	return 0
}

func (m *HProfRecordTrace) Type() HProfRecordType {
	return HProfRecordTypeTrace
}

func (m *HProfRecordTrace) Pos() int64 {
	if m != nil {
		return m.POS
	}
	return 0
}

func (m *HProfRecordTrace) GetStackTraceSerialNumber() uint32 {
	if m != nil {
		return m.StackTraceSerialNumber
	}
	return 0
}

func (m *HProfRecordTrace) GetThreadSerialNumber() uint32 {
	if m != nil {
		return m.ThreadSerialNumber
	}
	return 0
}

func (m *HProfRecordTrace) GetStackFrameIds() []uint64 {
	if m != nil {
		return m.StackFrameIds
	}
	return nil
}

type HProfRecordHeapDumpBoundary struct{}

func (m *HProfRecordHeapDumpBoundary) Id() uint64 {
	return 0
}

func (m *HProfRecordHeapDumpBoundary) Type() HProfRecordType {
	return HProfRecordTypeHeapDump
}

func (m *HProfRecordHeapDumpBoundary) Pos() int64 {
	return 0
}

// Class data dump.
type HProfClassDump struct {
	// Class object ID.
	ClassObjectId uint64 `json:"class_object_id,omitempty"`
	// Stack trace serial number.
	StackTraceSerialNumber uint32 `json:"stack_trace_serial_number,omitempty"`
	// Super class object ID, associated with another HProfClassDump.
	SuperClassObjectId uint64 `json:"super_class_object_id,omitempty"`
	// Class loader object ID, associated with HProfInstanceDump.
	ClassLoaderObjectId uint64 `json:"class_loader_object_id,omitempty"`
	// Signer of the class. (Looks like ClassLoaders can have signatures...)
	SignersObjectId uint64 `json:"signers_object_id,omitempty"`
	// Protection domain object ID. (No idea)
	ProtectionDomainObjectId uint64 `json:"protection_domain_object_id,omitempty"`
	// Instance size.
	InstanceSize        uint32                              `json:"instance_size,omitempty"`
	ConstantPoolEntries []*HProfClassDump_ConstantPoolEntry `json:"constant_pool_entries,omitempty"`
	StaticFields        []*HProfClassDump_StaticField       `json:"static_fields,omitempty"`
	InstanceFields      []*HProfClassDump_InstanceField     `json:"instance_fields,omitempty"`

	POS int64
}

func (m *HProfClassDump) Id() uint64 {
	if m != nil {
		return m.ClassObjectId
	}
	return 0
}

func (m *HProfClassDump) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeClassDump)
}

func (m *HProfClassDump) Pos() int64 {
	if m != nil {
		return m.POS
	}
	return 0
}

// GetReferences 计算连接
func (m *HProfClassDump) GetReferences() []uint64 {
	var references []uint64
	references = append(references, m.ClassObjectId)
	if m.SuperClassObjectId > 0 {
		references = append(references, m.SuperClassObjectId)
	}
	references = append(references, m.ClassLoaderObjectId)
	for _, field := range m.StaticFields {
		if field.Type == HProfValueType_OBJECT {
			references = append(references, field.Value)
		}
	}
	return references
}

func (m *HProfClassDump) GetClassObjectId() uint64 {
	if m != nil {
		return m.ClassObjectId
	}
	return 0
}

func (m *HProfClassDump) GetStackTraceSerialNumber() uint32 {
	if m != nil {
		return m.StackTraceSerialNumber
	}
	return 0
}

func (m *HProfClassDump) GetSuperClassObjectId() uint64 {
	if m != nil {
		return m.SuperClassObjectId
	}
	return 0
}

func (m *HProfClassDump) GetClassLoaderObjectId() uint64 {
	if m != nil {
		return m.ClassLoaderObjectId
	}
	return 0
}

func (m *HProfClassDump) GetSignersObjectId() uint64 {
	if m != nil {
		return m.SignersObjectId
	}
	return 0
}

func (m *HProfClassDump) GetProtectionDomainObjectId() uint64 {
	if m != nil {
		return m.ProtectionDomainObjectId
	}
	return 0
}

func (m *HProfClassDump) GetInstanceSize() uint32 {
	if m != nil {
		return m.InstanceSize
	}
	return 0
}

func (m *HProfClassDump) GetConstantPoolEntries() []*HProfClassDump_ConstantPoolEntry {
	if m != nil {
		return m.ConstantPoolEntries
	}
	return nil
}

func (m *HProfClassDump) GetStaticFields() []*HProfClassDump_StaticField {
	if m != nil {
		return m.StaticFields
	}
	return nil
}

func (m *HProfClassDump) GetInstanceFields() []*HProfClassDump_InstanceField {
	if m != nil {
		return m.InstanceFields
	}
	return nil
}

func (m *HProfClassDump) Size() uint32 {
	return 0
}

// Constant pool entry (appears to be unused according to heapDumper.cpp).
type HProfClassDump_ConstantPoolEntry struct {
	ConstantPoolIndex uint32         `json:"constant_pool_index,omitempty"`
	Type              HProfValueType `json:"type,omitempty"`
	Value             uint64         `json:"value,omitempty"`
}

func (m *HProfClassDump_ConstantPoolEntry) GetConstantPoolIndex() uint32 {
	if m != nil {
		return m.ConstantPoolIndex
	}
	return 0
}

func (m *HProfClassDump_ConstantPoolEntry) GetType() HProfValueType {
	if m != nil {
		return m.Type
	}
	return HProfValueType_UNKNOWN_HPROF_VALUE_TYPE
}

func (m *HProfClassDump_ConstantPoolEntry) GetValue() uint64 {
	if m != nil {
		return m.Value
	}
	return 0
}

// Static fields.
type HProfClassDump_StaticField struct {
	// Static field name, associated with HProfRecordUTF8.
	NameId uint64 `json:"name_id,omitempty"`
	// Type of the static field.
	Type HProfValueType `json:"type,omitempty"`
	// Value of the static field. Must be interpreted based on the type.
	Value uint64 `json:"value,omitempty"`
}

func (m *HProfClassDump_StaticField) GetNameId() uint64 {
	if m != nil {
		return m.NameId
	}
	return 0
}

func (m *HProfClassDump_StaticField) GetType() HProfValueType {
	if m != nil {
		return m.Type
	}
	return HProfValueType_UNKNOWN_HPROF_VALUE_TYPE
}

func (m *HProfClassDump_StaticField) GetValue() uint64 {
	if m != nil {
		return m.Value
	}
	return 0
}

// Instance fields.
type HProfClassDump_InstanceField struct {
	// Instance field name, associated with HProfRecordUTF8.
	NameId uint64 `json:"name_id,omitempty"`
	// Type of the instance field.
	Type HProfValueType `json:"type,omitempty"`

	// 只用来记录对象类型的 ID
	Value uint64
}

func (m *HProfClassDump_InstanceField) GetNameId() uint64 {
	if m != nil {
		return m.NameId
	}
	return 0
}

func (m *HProfClassDump_InstanceField) GetType() HProfValueType {
	if m != nil {
		return m.Type
	}
	return HProfValueType_UNKNOWN_HPROF_VALUE_TYPE
}

// Instance dump.
type HProfInstanceDump struct {
	// Object ID.
	ObjectId uint64 `json:"object_id,omitempty"`
	// Stack trace serial number.
	StackTraceSerialNumber uint32 `json:"stack_trace_serial_number,omitempty"`
	// Class object ID, associated with HProfClassDump.
	ClassObjectId uint64 `json:"class_object_id,omitempty"`
	// Instance field values.
	//
	// The instance field values are serialized in the order of the instance field
	// definition of HProfClassDump. If the class has three int fields, this
	// values starts from three 4-byte integers. Then, it continues to the super
	// class's instance fields.
	Values []byte `json:"values,omitempty"`

	POS int64
}

func (m *HProfInstanceDump) Id() uint64 {
	if m != nil {
		return m.ObjectId
	}
	return 0
}

func (m *HProfInstanceDump) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeInstanceDump)
}

func (m *HProfInstanceDump) Pos() int64 {
	if m != nil {
		return m.POS
	}
	return 0
}

func (m *HProfInstanceDump) ClassId() uint64 {
	return m.ClassObjectId
}

func (m *HProfInstanceDump) GetObjectId() uint64 {
	if m != nil {
		return m.ObjectId
	}
	return 0
}

func (m *HProfInstanceDump) GetStackTraceSerialNumber() uint32 {
	if m != nil {
		return m.StackTraceSerialNumber
	}
	return 0
}

func (m *HProfInstanceDump) GetClassObjectId() uint64 {
	if m != nil {
		return m.ClassObjectId
	}
	return 0
}

func (m *HProfInstanceDump) GetValues() []byte {
	if m != nil {
		return m.Values
	}
	return nil
}

func (m *HProfInstanceDump) Size() uint32 {
	//return uint32(8 + 4 + 8 + 4 + len(m.Values))
	return uint32(len(m.Values) + 16) // TODO 空 instance 大小是 16 还是 0 ？
}

// Object array dump.
type HProfObjectArrayDump struct {
	// Object ID.
	ArrayObjectId uint64 `json:"array_object_id,omitempty"`
	// Stack trace serial number.
	StackTraceSerialNumber uint32 `json:"stack_trace_serial_number,omitempty"`
	// Class object ID of the array elements, associated with HProfClassDump.
	ArrayClassObjectId uint64 `json:"array_class_object_id,omitempty"`
	// Element object IDs.
	ElementObjectIds []uint64 `json:"element_object_ids,omitempty"`

	POS int64
}

func (m *HProfObjectArrayDump) Id() uint64 {
	if m != nil {
		return m.ArrayObjectId
	}
	return 0
}

func (m *HProfObjectArrayDump) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeObjectArrayDump)
}

func (m *HProfObjectArrayDump) Pos() int64 {
	if m != nil {
		return m.POS
	}
	return 0
}

func (m *HProfObjectArrayDump) ClassId() uint64 {
	return m.ArrayClassObjectId
}

func (m *HProfObjectArrayDump) GetArrayObjectId() uint64 {
	if m != nil {
		return m.ArrayObjectId
	}
	return 0
}

func (m *HProfObjectArrayDump) GetStackTraceSerialNumber() uint32 {
	if m != nil {
		return m.StackTraceSerialNumber
	}
	return 0
}

func (m *HProfObjectArrayDump) GetArrayClassObjectId() uint64 {
	if m != nil {
		return m.ArrayClassObjectId
	}
	return 0
}

func (m *HProfObjectArrayDump) GetElementObjectIds() []uint64 {
	if m != nil {
		return m.ElementObjectIds
	}
	return nil
}

func (m *HProfObjectArrayDump) GetClassObjectId() uint64 {
	return m.GetArrayClassObjectId()
}
func (m *HProfObjectArrayDump) Size() uint32 {
	return uint32(8 + 4 + 8 + 4 + len(m.ElementObjectIds)*global.ID_SIZE)
}

// Primitive array dump.
type HProfPrimitiveArrayDump struct {
	// Object ID.
	ArrayObjectId uint64 `json:"array_object_id,omitempty"`
	// Stack trace serial number.
	StackTraceSerialNumber uint32 `json:"stack_trace_serial_number,omitempty"`
	// Type of the elements.
	ElementType HProfValueType `json:"element_type,omitempty"`
	// Element values.
	//
	// Values need to be parsed based on the element_type. If the array is an int
	// array with three elements, this field has 12 bytes.
	Values []byte `json:"values,omitempty"`

	POS int64
}

func (m *HProfPrimitiveArrayDump) Id() uint64 {
	if m != nil {
		return m.ArrayObjectId
	}
	return 0
}

func (m *HProfPrimitiveArrayDump) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypePrimitiveArrayDump)
}

func (m *HProfPrimitiveArrayDump) Pos() int64 {
	if m != nil {
		return m.POS
	}
	return 0
}

func (m *HProfPrimitiveArrayDump) ClassId() uint64 {
	return uint64(m.ElementType)
}

func (m *HProfPrimitiveArrayDump) GetArrayObjectId() uint64 {
	if m != nil {
		return m.ArrayObjectId
	}
	return 0
}

func (m *HProfPrimitiveArrayDump) GetStackTraceSerialNumber() uint32 {
	if m != nil {
		return m.StackTraceSerialNumber
	}
	return 0
}

func (m *HProfPrimitiveArrayDump) GetElementType() HProfValueType {
	if m != nil {
		return m.ElementType
	}
	return HProfValueType_UNKNOWN_HPROF_VALUE_TYPE
}

func (m *HProfPrimitiveArrayDump) GetValues() []byte {
	if m != nil {
		return m.Values
	}
	return nil
}

func (m *HProfPrimitiveArrayDump) GetClassObjectId() uint64 {
	return m.GetArrayObjectId()
}
func (m *HProfPrimitiveArrayDump) Size() uint32 {
	return uint32(8 + 4 + 8 + 4 + len(m.Values))
}

// Root object pointer of JNI globals.
type HProfRootJNIGlobal struct {
	// Object ID.
	ObjectId uint64 `json:"object_id,omitempty"`
	// JNI global ref ID. (No idea)
	JniGlobalRefId uint64 `json:"jni_global_ref_id,omitempty"`

	POS int64
}

func (m *HProfRootJNIGlobal) Id() uint64 {
	if m != nil {
		return m.ObjectId
	}
	return 0
}

func (m *HProfRootJNIGlobal) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeRootJNIGlobal)
}

func (m *HProfRootJNIGlobal) Pos() int64 {
	if m != nil {
		return m.POS
	}
	return 0
}

func (m *HProfRootJNIGlobal) GetObjectId() uint64 {
	if m != nil {
		return m.ObjectId
	}
	return 0
}

func (m *HProfRootJNIGlobal) GetJniGlobalRefId() uint64 {
	if m != nil {
		return m.JniGlobalRefId
	}
	return 0
}

// Root object pointer of JNI locals.
type HProfRootJNILocal struct {
	// Object ID.
	ObjectId uint64 `json:"object_id,omitempty"`
	// Thread serial number.
	ThreadSerialNumber uint32 `json:"thread_serial_number,omitempty"`
	// Frame number in the trace.
	FrameNumberInStackTrace uint32 `json:"frame_number_in_stack_trace,omitempty"`

	POS int64
}

func (m *HProfRootJNILocal) Id() uint64 {
	if m != nil {
		return m.ObjectId
	}
	return 0
}

func (m *HProfRootJNILocal) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeRootJNILocal)
}

func (m *HProfRootJNILocal) Pos() int64 {
	if m != nil {
		return m.POS
	}
	return 0
}

func (m *HProfRootJNILocal) GetObjectId() uint64 {
	if m != nil {
		return m.ObjectId
	}
	return 0
}

func (m *HProfRootJNILocal) GetThreadSerialNumber() uint32 {
	if m != nil {
		return m.ThreadSerialNumber
	}
	return 0
}

func (m *HProfRootJNILocal) GetFrameNumberInStackTrace() uint32 {
	if m != nil {
		return m.FrameNumberInStackTrace
	}
	return 0
}

// Root object pointer on JVM stack (e.g. local variables).
type HProfRootJavaFrame struct {
	// Object ID.
	ObjectId uint64 `json:"object_id,omitempty"`
	// Thread serial number.
	ThreadSerialNumber uint32 `json:"thread_serial_number,omitempty"`
	// Frame number in the trace.
	FrameNumberInStackTrace uint32 `json:"frame_number_in_stack_trace,omitempty"`

	POS int64
}

func (m *HProfRootJavaFrame) Id() uint64 {
	if m != nil {
		return m.ObjectId
	}
	return 0
}

func (m *HProfRootJavaFrame) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeRootJavaFrame)
}

func (m *HProfRootJavaFrame) Pos() int64 {
	if m != nil {
		return m.POS
	}
	return 0
}

func (m *HProfRootJavaFrame) GetObjectId() uint64 {
	if m != nil {
		return m.ObjectId
	}
	return 0
}

func (m *HProfRootJavaFrame) GetThreadSerialNumber() uint32 {
	if m != nil {
		return m.ThreadSerialNumber
	}
	return 0
}

func (m *HProfRootJavaFrame) GetFrameNumberInStackTrace() uint32 {
	if m != nil {
		return m.FrameNumberInStackTrace
	}
	return 0
}

// System classes (No idea).
type HProfRootStickyClass struct {
	// Object ID.
	ObjectId uint64 `json:"object_id,omitempty"`

	POS int64
}

func (m *HProfRootStickyClass) Id() uint64 {
	if m != nil {
		return m.ObjectId
	}
	return 0
}

func (m *HProfRootStickyClass) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeRootStickyClass)
}

func (m *HProfRootStickyClass) Pos() int64 {
	if m != nil {
		return m.POS
	}
	return 0
}

func (m *HProfRootStickyClass) GetObjectId() uint64 {
	if m != nil {
		return m.ObjectId
	}
	return 0
}

// Thread object.
type HProfRootThreadObj struct {
	// Object ID.
	ThreadObjectId uint64 `json:"thread_object_id,omitempty"`
	// Thread sequence number. (It seems this is same as thread serial number.)
	ThreadSequenceNumber uint32 `json:"thread_sequence_number,omitempty"`
	// Stack trace serial number.
	StackTraceSequenceNumber uint32 `json:"stack_trace_sequence_number,omitempty"`

	POS int64
}

func (m *HProfRootThreadObj) Id() uint64 {
	if m != nil {
		return m.ThreadObjectId
	}
	return 0
}

func (m *HProfRootThreadObj) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeRootThreadObj)
}

func (m *HProfRootThreadObj) Pos() int64 {
	if m != nil {
		return m.POS
	}
	return 0
}

func (m *HProfRootThreadObj) GetThreadObjectId() uint64 {
	if m != nil {
		return m.ThreadObjectId
	}
	return 0
}

func (m *HProfRootThreadObj) GetThreadSequenceNumber() uint32 {
	if m != nil {
		return m.ThreadSequenceNumber
	}
	return 0
}

func (m *HProfRootThreadObj) GetStackTraceSequenceNumber() uint32 {
	if m != nil {
		return m.StackTraceSequenceNumber
	}
	return 0
}

// Busy monitor.
type HProfRootMonitorUsed struct {
	// Object ID.
	ObjectId uint64 `json:"object_id,omitempty"`
}

func (m *HProfRootMonitorUsed) Id() uint64 {
	return 0
}

func (m *HProfRootMonitorUsed) Type() HProfRecordType {
	return HProfHDRecordTypeRootMonitorUsed
}

func (m *HProfRootMonitorUsed) Pos() int64 {
	return 0
}

func (m *HProfRootMonitorUsed) GetObjectId() uint64 {
	if m != nil {
		return m.ObjectId
	}
	return 0
}

type HProfDumpWithSize interface {
	GetClassObjectId() uint64
	Size() uint32
}
