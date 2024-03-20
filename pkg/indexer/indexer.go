package indexer

import (
	"fmt"
	"hprof-tool/pkg/hprof"
	"hprof-tool/pkg/model"
	"hprof-tool/pkg/storage"
)

const (
	OBJECT_CLASS_NAME      = "java/lang/Object"
	LANG_CLASS_NAME        = "java/lang/Class"
	CLASSLOADER_CLASS_NAME = "java/lang/ClassLoader"
)

var PRIMITIVE_TYPE_ARRAY = []string{"", "", "", "",
	"boolean[]", "char[]", "float[]", "double[]",
	"byte[]", "short[]", "int[]", "long[]"}

type Indexer struct {
	hreader *hprof.HProfReader
	storage storage.Storage

	ctx *HeapContext
}

func NewSqliteIndexer(hreader *hprof.HProfReader, storage storage.Storage) *Indexer {
	return &Indexer{
		hreader: hreader,
		storage: storage,

		ctx: newHeapContext(),
	}
}

func (i *Indexer) GetText(tid uint64) (string, error) {
	pos, text, err := i.storage.GetText(tid)
	if err != nil {
		return "", err
	}
	if pos == -1 {
		return text, nil
	}
	record, err := hprof.ReadHProfUTF8RecordWithPos(i.hreader, pos)
	if err != nil {
		return "", err
	}
	return string(record.Name), nil
}

// ForEachClassesWithName 获取所有的 classId 和 className
func (i *Indexer) ForEachClassesWithName(fn func(cid uint64, cname string) error) error {
	return i.storage.ListClasses(func(cid uint64, pos int64, cla *hprof.HProfClassRecord) error {
		_, nameId, err := i.storage.GetLoadClassByClassId(cid)
		if err != nil {
			return err
		}
		name, err := i.GetText(nameId)
		if err != nil {
			return err
		}
		return fn(cid, name)
	})
}

// ForEachClassRecords 获取所有的 class record
func (i *Indexer) ForEachClassRecords(fn func(record *hprof.HProfClassRecord) error) error {
	return i.storage.ListClasses(func(cid uint64, pos int64, cla *hprof.HProfClassRecord) error {
		if cla != nil {
			return fn(cla)
		}
		cla, err := hprof.ReadHProfClassRecordWithPos(i.hreader, pos)
		if err != nil {
			return nil
		}
		return fn(cla)
	})
}

// ForEachInstanceRecords 获取所有的 instance record
func (i *Indexer) ForEachInstanceRecords(fn func(record *hprof.HProfInstanceRecord) error) error {
	return i.storage.ListInstances(func(oid uint64, pos int64, cid uint64) error {
		instance, err := hprof.ReadHProfInstanceRecordWithPos(i.hreader, pos)
		if err != nil {
			return nil
		}
		return fn(instance)
	})
}

func (i *Indexer) ForEachThreads(fn func(record *hprof.HProfThreadRecord) error) error {
	return i.storage.ListThreads(fn)
}

func (i *Indexer) ForEachThreadTraces(fn func(record *hprof.HProfTraceRecord) error) error {
	return i.storage.ListThreadTraces(fn)
}

func (i *Indexer) ForEachThreadFrames(fn func(record *hprof.HProfFrameRecord) error) error {
	return i.storage.ListThreadFrames(fn)
}

func (i *Indexer) ForEachGCRoots(fn func(record hprof.HProfRecord) error) error {
	return i.storage.ListGCRoots(func(pos int64, typ int) error {
		var record hprof.HProfRecord
		var err error
		switch typ {
		case hprof.GCRootType_NATIVE_STATIC:
			record, err = hprof.ReadHProfRootJNIGlobalWithPos(i.hreader, pos)
			if err != nil {
				return err
			}
			err = fn(record)
		case hprof.GCRootType_NATIVE_LOCAL:
			record, err = hprof.ReadHProfRootJNILocalWithPos(i.hreader, pos)
			if err != nil {
				return err
			}
			err = fn(record)
		case hprof.GCRootType_JAVA_LOCAL:
			record, err = hprof.ReadHProfRootJavaFrameWithPos(i.hreader, pos)
			if err != nil {
				return err
			}
			err = fn(record)
		case hprof.GCRootType_SYSTEM_CLASS:
			record, err = hprof.ReadHProfRootStickyClassWithPos(i.hreader, pos)
			if err != nil {
				return err
			}
			err = fn(record)
		case hprof.GCRootType_THREAD_OBJ:
			record, err = hprof.ReadHProfRootThreadObjWithPos(i.hreader, pos)
			if err != nil {
				return err
			}
			err = fn(record)
		case hprof.GCRootType_BUSY_MONITOR:
			record, err = hprof.ReadHProfRootMonitorUsedWithPos(i.hreader, pos)
			if err != nil {
				return err
			}
			err = fn(record)
		default:
			err = fmt.Errorf("unknown gc root record type: %d", typ)
		}
		return err
	})
}

func (i *Indexer) getClassIdByName(name string, defaultId uint64) uint64 {
	cids, exist := i.ctx.className2Cid[name]
	if !exist || len(cids) == 0 {
		return defaultId
	}
	return cids[0]
}

func (i *Indexer) GetClassNameById(classId uint64, defaultName string) string {
	name, exist := i.ctx.classId2Name[classId]
	if !exist {
		return defaultName
	}
	return name
}

func (i *Indexer) GetClassNameByClassSerialNumber(serialNumber uint64) (string, error) {
	cid, _, err := i.storage.GetLoadClassById(serialNumber)
	if err != nil {
		return "", err
	}
	name, exist := i.ctx.classId2Name[cid]
	if !exist {
		return "UNKNOWN", nil
	}
	return name, nil
}

func (i *Indexer) getClassById(cid uint64) (*hprof.HProfClassRecord, error) {
	pos, class, err := i.storage.GetClass(cid)
	if err != nil {
		return nil, err
	}
	if class != nil {
		return class, nil
	}
	class, err = hprof.ReadHProfClassRecordWithPos(i.hreader, pos)
	return class, err
}

func (i *Indexer) GetThreads() map[uint32]*model.Thread {
	// TODO 修复 Thread 信息的处理
	//var traces = map[uint32]*model.StackTrace{}
	var threads = map[uint32]*model.Thread{}
	for _, v := range i.ctx.serNum2stackTrace {
		trace := &model.StackTrace{
			ThreadSerialNumber: v.ThreadSerialNumber,
		}
		for _, fid := range v.FrameIds {
			trace.Frames = append(trace.Frames, i.ctx.id2frame[fid])
		}
		trace.Locals = i.ctx.thread2locals[v.ThreadSerialNumber]
		//traces[k] = trace
		thread := i.ctx.threadSN2thread[v.ThreadSerialNumber]
		fmt.Printf("Get Thread by ThreadSerialNumber = %d\n", v.ThreadSerialNumber)
		threads[v.ThreadSerialNumber] = &model.Thread{
			ObjectId:          thread.ObjectId,
			NameId:            thread.NameId,
			GroupNameId:       thread.GroupNameId,
			GroupParentNameId: thread.GroupParentNameId,
			StackTrace:        trace,
		}
	}
	return threads
}

func (i *Indexer) GetClassesStatistics(fn func(cid uint64, cname string, count, size int64) error) error {
	err := i.storage.CountInstancesByClass(func(cid uint64, count, size int64) error {
		name := i.GetClassNameById(cid, "unkonwn")
		return fn(cid, name, count, size)
	})
	if err != nil {
		return err
	}
	err = i.storage.CountObjectArrayByClass(func(cid uint64, count, size int64) error {
		name := i.GetClassNameById(cid, "unkonwn")
		return fn(cid, name, count, size)
	})
	if err != nil {
		return err
	}
	return i.storage.CountPrimitiveArrayByType(func(ty uint64, count, size int64) error {
		name := PRIMITIVE_TYPE_ARRAY[ty]
		return fn(ty, name, count, size)
	})
}

func (i *Indexer) GetClassDetail(cid uint64) (*Class, error) {
	class, err := i.getClassById(cid)
	if err != nil {
		return nil, err
	}

	//classerLoader, err := i.GetInstanceDetail(class.ClassLoaderObjectId)
	//if err != nil {
	//	return nil, err
	//}

	var constantPoolEntries []*ConstantPoolEntry
	for _, entry := range class.ConstantPoolEntries {
		constantPoolEntries = append(constantPoolEntries, &ConstantPoolEntry{
			Index: entry.ConstantPoolIndex,
			Type:  hprof.HProfValueType_name[entry.Type],
			Value: entry.Value,
		})

	}

	var staticFields []*StaticField
	for _, field := range class.StaticFields {
		name, err := i.GetText(field.NameId)
		if err != nil {
			return nil, err
		}
		staticFields = append(staticFields, &StaticField{
			Name:  name,
			Type:  hprof.HProfValueType_name[field.Type],
			Value: field.Value,
		})
	}

	var instanceFields []*InstanceField
	for _, field := range class.InstanceFields {
		name, err := i.GetText(field.NameId)
		if err != nil {
			return nil, err
		}
		instanceFields = append(instanceFields, &InstanceField{
			Name: name,
			Type: hprof.HProfValueType_name[field.Type],
		})

	}

	className := i.ctx.classId2Name[cid]
	return &Class{
		Id:                       class.ClassObjectId,
		LoaderObjectId:           class.ClassLoaderObjectId,
		SuperClassId:             class.SuperClassObjectId,
		StackTraceSerialNumber:   class.StackTraceSerialNumber,
		SignersObjectId:          class.SignersObjectId,
		Name:                     className,
		ProtectionDomainObjectId: class.ProtectionDomainObjectId,
		InstanceSize:             class.InstanceSize,
		ConstantPoolEntries:      constantPoolEntries,
		StaticFields:             staticFields,
		InstanceFields:           instanceFields,
	}, nil
}

func (i *Indexer) AppendReference(from, to uint64, typ int) error {
	return i.storage.AppendReference(from, to, typ)
}

func (i *Indexer) GetInstanceDetail(oid uint64) (*Instance, error) {
	instanceRecord, err := i.getInstance(oid)
	if err != nil {
		return nil, err
	}
	class, err := i.getClassById(instanceRecord.ClassObjectId)
	if err != nil {
		return nil, err
	}
	classes, err := i.resolveClassHierarchy(class)
	if err != nil {
		return nil, err
	}
	instanceFields := []*hprof.HProfClass_InstanceField{}
	for _, class := range classes {
		instanceFields = append(instanceFields, class.InstanceFields...)
	}
	fiedValues, err := instanceRecord.ReadValues(instanceFields)
	if err != nil {
		return nil, err
	}
	var fields []*InstanceField
	for idx, field := range instanceFields {
		name, err := i.GetText(field.NameId)
		if err != nil {
			return nil, err
		}
		var refrence *Instance = nil
		if field.Type == hprof.HProfValueType_OBJECT {
			refrence, err = i.GetInstanceDetail(fiedValues[idx].(*hprof.HProfInstanceObjectValue).Value)
		}
		fields = append(fields, &InstanceField{
			Name:      name,
			Type:      hprof.HProfValueType_name[field.Type],
			Value:     fiedValues[idx].ValueString(),
			Reference: refrence,
		})
	}
	className := i.ctx.classId2Name[instanceRecord.ClassObjectId]
	instance := &Instance{
		Id:     instanceRecord.ObjectId,
		Class:  className,
		Fields: fields,
	}
	return instance, nil
}

func (i *Indexer) GetInstancesStatistics(cid uint64, typ int, fn func(id uint64, size int64) error) error {
	if typ == hprof.HProfHDRecordTypeObjectArrayDump {
		return i.storage.ListObjectArrayByClass(cid, func(id uint64, pos, size int64) error {
			return fn(id, size)
		})
	}
	if typ == hprof.HProfHDRecordTypePrimitiveArrayDump {
		return i.storage.ListPrimitiveArrayByClass(cid, func(id uint64, pos, size int64) error {
			return fn(id, size)
		})
	}
	return i.storage.ListInstancesByClass(cid, func(id uint64, pos, size int64) error {
		return fn(id, size)
	})
}

// GetRecordInbounds 列出当前 record 的来源 reference
func (i *Indexer) GetRecordInbounds(id uint64, fn func(record hprof.HProfRecord) error) error {
	return i.storage.ListInboundReferences(id, func(from uint64, typ int) error {
		record, err := i.getRecord(from)
		if err != nil {
			return err
		}
		return fn(record)
	})
}

// GetRecordOutbounds 列出当前 record 的来源 reference
func (i *Indexer) GetRecordOutbounds(id uint64, fn func(record hprof.HProfRecord) error) error {
	return i.storage.ListOutboundReferences(id, func(to uint64, typ int) error {
		record, err := i.getRecord(to)
		if err != nil {
			return err
		}
		return fn(record)
	})
}

func (i *Indexer) getInstance(oid uint64) (*hprof.HProfInstanceRecord, error) {
	pos, err := i.storage.GetInstanceById(oid)
	if err != nil {
		return nil, err
	}
	return hprof.ReadHProfInstanceRecordWithPos(i.hreader, pos)
}

func (i *Indexer) getRecord(id uint64) (hprof.HProfRecord, error) {
	pos, typ, cla, err := i.storage.GetRecordById(id)
	if err != nil {
		return nil, err
	}
	if cla != nil {
		return cla, nil
	}
	switch typ {
	case hprof.HProfHDRecordTypeClassDump:
		return hprof.ReadHProfClassRecordWithPos(i.hreader, pos)
	case hprof.HProfHDRecordTypeInstanceDump:
		return hprof.ReadHProfInstanceRecordWithPos(i.hreader, pos)
	case hprof.HProfHDRecordTypeObjectArrayDump:
		return hprof.ReadHProfObjectArrayRecordWithPos(i.hreader, pos)
	case hprof.HProfHDRecordTypePrimitiveArrayDump:
		return hprof.ReadHProfPrimitiveArrayRecordWithPos(i.hreader, pos)
	default:
		return nil, fmt.Errorf("unknown record type: %d", typ)
	}
}

func (i *Indexer) resolveClassHierarchy(class *hprof.HProfClassRecord) ([]*hprof.HProfClassRecord, error) {
	var classes []*hprof.HProfClassRecord
	classes = append(classes, class)
	c := class
	var err error
	for {
		if c.SuperClassObjectId == 0 {
			break
		}
		c, err = i.getClassById(c.SuperClassObjectId)
		if err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}

	return classes, nil
}
