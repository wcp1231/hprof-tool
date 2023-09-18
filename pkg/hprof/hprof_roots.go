package hprof

// Root object pointer of JNI globals.
type HProfRootJNIGlobal struct {
	HProfBasicRecord

	// Object ID.
	ObjectId uint64
	// JNI global ref ID. (No idea)
	JniGlobalRefId uint64
}

func (m *HProfRootJNIGlobal) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeRootJNIGlobal)
}

func ReadHProfRootJNIGlobal(pr *HProfReader) (*HProfRootJNIGlobal, error) {
	pos := pr.pos
	oid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	rid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	size := int(pr.pos - pos)
	return &HProfRootJNIGlobal{
		HProfBasicRecord: HProfBasicRecord{pos, size},
		ObjectId:         oid,
		JniGlobalRefId:   rid,
	}, nil
}

func ReadHProfRootJNIGlobalWithPos(pr *HProfReader, pos int64) (*HProfRootJNIGlobal, error) {
	err := pr.Seek(pos)
	if err != nil {
		return nil, err
	}
	return ReadHProfRootJNIGlobal(pr)
}

// Root object pointer of JNI locals.
type HProfRootJNILocal struct {
	HProfBasicRecord

	// Object ID.
	ObjectId uint64 `json:"object_id,omitempty"`
	// Thread serial number.
	ThreadSerialNumber uint32 `json:"thread_serial_number,omitempty"`
	// Frame number in the trace.
	FrameNumberInStackTrace uint32 `json:"frame_number_in_stack_trace,omitempty"`
}

func (m *HProfRootJNILocal) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeRootJNILocal)
}

func ReadHProfRootJNILocal(pr *HProfReader) (*HProfRootJNILocal, error) {
	pos := pr.pos
	oid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	tsn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	fn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	size := int(pr.pos - pos)
	return &HProfRootJNILocal{
		HProfBasicRecord:        HProfBasicRecord{pos, size},
		ObjectId:                oid,
		ThreadSerialNumber:      tsn,
		FrameNumberInStackTrace: fn,
	}, nil
}

func ReadHProfRootJNILocalWithPos(pr *HProfReader, pos int64) (*HProfRootJNILocal, error) {
	err := pr.Seek(pos)
	if err != nil {
		return nil, err
	}
	return ReadHProfRootJNILocal(pr)
}

// Root object pointer on JVM stack (e.g. local variables).
type HProfRootJavaFrame struct {
	HProfBasicRecord

	// Object ID.
	ObjectId uint64 `json:"object_id,omitempty"`
	// Thread serial number.
	ThreadSerialNumber uint32 `json:"thread_serial_number,omitempty"`
	// Frame number in the trace.
	FrameNumberInStackTrace uint32 `json:"frame_number_in_stack_trace,omitempty"`
}

func (m *HProfRootJavaFrame) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeRootJavaFrame)
}

func ReadHProfRootJavaFrame(pr *HProfReader) (*HProfRootJavaFrame, error) {
	pos := pr.pos
	oid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	tsn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	fn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	size := int(pr.pos - pos)
	return &HProfRootJavaFrame{
		HProfBasicRecord:        HProfBasicRecord{pos, size},
		ObjectId:                oid,
		ThreadSerialNumber:      tsn,
		FrameNumberInStackTrace: fn,
	}, nil
}

func ReadHProfRootJavaFrameWithPos(pr *HProfReader, pos int64) (*HProfRootJavaFrame, error) {
	err := pr.Seek(pos)
	if err != nil {
		return nil, err
	}
	return ReadHProfRootJavaFrame(pr)
}

// System classes (No idea).
type HProfRootStickyClass struct {
	HProfBasicRecord

	// Object ID.
	ObjectId uint64 `json:"object_id,omitempty"`
}

func (m *HProfRootStickyClass) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeRootStickyClass)
}

func ReadHProfRootStickyClass(pr *HProfReader) (*HProfRootStickyClass, error) {
	pos := pr.pos
	oid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	size := int(pr.pos - pos)
	return &HProfRootStickyClass{
		HProfBasicRecord: HProfBasicRecord{pos, size},
		ObjectId:         oid,
	}, nil
}

func ReadHProfRootStickyClassWithPos(pr *HProfReader, pos int64) (*HProfRootStickyClass, error) {
	err := pr.Seek(pos)
	if err != nil {
		return nil, err
	}
	return ReadHProfRootStickyClass(pr)
}

// Thread object.
type HProfRootThreadObj struct {
	HProfBasicRecord

	// Object ID.
	ThreadObjectId uint64 `json:"thread_object_id,omitempty"`
	// Thread sequence number. (It seems this is same as thread serial number.)
	ThreadSequenceNumber uint32 `json:"thread_sequence_number,omitempty"`
	// Stack trace serial number.
	StackTraceSequenceNumber uint32 `json:"stack_trace_sequence_number,omitempty"`
}

func (m *HProfRootThreadObj) Type() HProfRecordType {
	return HProfRecordType(HProfHDRecordTypeRootThreadObj)
}

func ReadHProfRootThreadObj(pr *HProfReader) (*HProfRootThreadObj, error) {
	pos := pr.pos
	toid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	tsn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	stsn, err := pr.readUint32()
	if err != nil {
		return nil, err
	}
	size := int(pr.pos - pos)
	return &HProfRootThreadObj{
		HProfBasicRecord:         HProfBasicRecord{pos, size},
		ThreadObjectId:           toid,
		ThreadSequenceNumber:     tsn,
		StackTraceSequenceNumber: stsn,
	}, nil
}

func ReadHProfRootThreadObjWithPos(pr *HProfReader, pos int64) (*HProfRootThreadObj, error) {
	err := pr.Seek(pos)
	if err != nil {
		return nil, err
	}
	return ReadHProfRootThreadObj(pr)
}

// Busy monitor.
type HProfRootMonitorUsed struct {
	HProfBasicRecord
	// Object ID.
	ObjectId uint64 `json:"object_id,omitempty"`
}

func (m *HProfRootMonitorUsed) Type() HProfRecordType {
	return HProfHDRecordTypeRootMonitorUsed
}

func ReadHProfRootMonitorUsed(pr *HProfReader) (*HProfRootMonitorUsed, error) {
	oid, err := pr.readID()
	if err != nil {
		return nil, err
	}
	return &HProfRootMonitorUsed{
		ObjectId: oid,
	}, nil
}

func ReadHProfRootMonitorUsedWithPos(pr *HProfReader, pos int64) (*HProfRootMonitorUsed, error) {
	err := pr.Seek(pos)
	if err != nil {
		return nil, err
	}
	return ReadHProfRootMonitorUsed(pr)
}
