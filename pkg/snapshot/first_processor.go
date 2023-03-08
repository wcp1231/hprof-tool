package snapshot

import (
	"bytes"
	"fmt"
	"hprof-tool/pkg/model"
	"hprof-tool/pkg/parser"
	"hprof-tool/pkg/storage"
	"io"
	"math"
)

const (
	OBJECT_CLASS_NAME      = "java/lang/Object"
	LANG_CLASS_NAME        = "java/lang/Class"
	CLASSLOADER_CLASS_NAME = "java/lang/ClassLoader"
)

var PRIMITIVE_TYPE_ARRAY = []string{"", "", "", "",
	"boolean[]", "char[]", "float[]", "double[]",
	"byte[]", "short[]", "int[]", "long[]"}

type stackTrace struct {
	ThreadSerialNumber uint32
	FrameIds           []uint64
}

type FirstProcessor struct {
	snapshot *Snapshot
	storage  storage.Storage
	parser   *parser.HProfParser

	header *parser.HProfHeader

	// 用于计算 GCRoot 和线程信息
	thread2Id map[uint32]uint64

	gcRoots map[uint64][]*model.GCRootInfo
	// map[threadId]map[id][]GcRoot
	threadAddressToLocals map[uint64]map[uint64][]*model.GCRootInfo

	// 线程相关
	thread2locals     map[uint32][]*model.LocalFrame
	id2frame          map[uint64]*model.StackFrame
	serNum2stackTrace map[uint32]*stackTrace

	// ClassId 和 size
	requiredClasses         map[uint64]int
	requiredArrayClasses    map[uint64]bool
	requiredPrimitiveArrays map[model.HProfValueType]bool
	maxObjectId             uint64
	objectAlign             uint64

	// 系统类相关
	idx2Name      map[uint64]string
	javaClassName map[string]uint64

	//
	classHierarchyCache map[uint64][]*model.HProfClassDump

	//
	outbound map[uint64][]uint64
}

func NewFirstProcessor(snapshot *Snapshot, parser *parser.HProfParser) *FirstProcessor {
	return &FirstProcessor{
		snapshot: snapshot,
		storage:  snapshot.storage,
		parser:   parser,

		thread2Id:               map[uint32]uint64{},
		gcRoots:                 map[uint64][]*model.GCRootInfo{},
		threadAddressToLocals:   map[uint64]map[uint64][]*model.GCRootInfo{},
		thread2locals:           map[uint32][]*model.LocalFrame{},
		id2frame:                map[uint64]*model.StackFrame{},
		serNum2stackTrace:       map[uint32]*stackTrace{},
		requiredClasses:         map[uint64]int{},
		requiredArrayClasses:    map[uint64]bool{},
		requiredPrimitiveArrays: map[model.HProfValueType]bool{},
		idx2Name:                map[uint64]string{},
		javaClassName:           map[string]uint64{},
		classHierarchyCache:     map[uint64][]*model.HProfClassDump{},
		outbound:                map[uint64][]uint64{},
	}
}

// Process 初次处理 HProf 文件
// 1 解析所有 record 并索引
// 2 补充 instance 和一般 class 需要的系统类信息
// 3 处理线程信息和 GC Roots 信息
// 4 记录类的所有 出连接 outbound
func (p *FirstProcessor) Process() error {
	err := p.parseWholeFile()
	if err != nil {
		return err
	}

	err = p.saveThreads()
	if err != nil {
		return err
	}

	// TODO 处理 GC Roots 相关信息
	p.GetGCRoots()
	p.GetThreadAddressToLocals()

	err = p.calculateAlignment()
	if err != nil {
		return err
	}
	err = p.createRequiredFakeClasses()
	if err != nil {
		return err
	}
	// TODO addTypesAndDummyStatics

	// TODO 计算每个 Class 自身占用 Heap 大小，以及它的 instance 的大小

	err = p.logReferencesForClass()
	if err != nil {
		return err
	}

	count, err := p.storage.CountInstances()
	if err != nil {
		return err
	}
	fmt.Printf("Instances count: %d\n", count)

	// pass2parser
	err = p.readObjects()
	if err != nil {
		return err
	}

	return nil
}

func (p *FirstProcessor) logReferencesForClass() error {
	err := p.storage.ListClasses(func(class *model.HProfClassDump) error {
		// TODO 这里要干啥？
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (p *FirstProcessor) saveThreads() error {
	threads := p.getThreads()
	return p.storage.PutKV(storage.THREADS_KEY, threads)
}

// calculateAlignment 计算 objectAlign
// 用于创建虚拟类，以及计算对象大小
func (p *FirstProcessor) calculateAlignment() error {
	const minAlign = 8
	const maxAlign = 256
	max := func(a, b uint64) uint64 {
		if a > b {
			return a
		}
		return b
	}
	min := func(a, b uint64) uint64 {
		if a > b {
			return b
		}
		return a
	}
	var prev uint64 = 0
	var align uint64 = 0
	err := p.storage.ListInstances(func(instance model.HProfObjectRecord) error {
		next := instance.Id()
		if next == 0 {
			return nil
		}
		var diff = next - prev
		prev = next
		if next == diff {
			return nil
		}
		if align == 0 {
			align = diff
			return nil
		}
		mx := max(align, diff)
		mn := min(align, diff)
		d := mx % mn
		for d != 0 {
			mx = mn
			mn = d
			d = mx % mn
		}
		align = mn
		if align <= minAlign {
			return io.EOF
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return err
	}
	p.objectAlign = max(min(align, maxAlign), minAlign)
	return nil
}

// createRequiredFakeClasses 创建缺失的系统类
func (p *FirstProcessor) createRequiredFakeClasses() error {
	// 创建 java 系统类
	objectClass, err := p.storage.FindLoadClassByName(p.javaClassName[OBJECT_CLASS_NAME])
	if err != nil {
		return err
	}
	// 创建两个系统类
	nameId, err := p.getOrCreateText(LANG_CLASS_NAME)
	if err != nil {
		return err
	}
	newClass, err := p.storage.FindLoadClassByName(nameId)
	if err != nil {
		return err
	}
	nextObjectAddress := p.maxObjectId
	// java.lang.Class 的 ID
	var langClassId uint64
	if newClass == nil {
		nextObjectAddress = +p.objectAlign
		err = p.addFakeClass(nextObjectAddress, objectClass.ClassObjectId, nameId, nil)
		if err != nil {
			return err
		}
		langClassId = nextObjectAddress
	} else {
		langClassId = newClass.ClassObjectId
	}
	nameId, err = p.getOrCreateText(CLASSLOADER_CLASS_NAME)
	if err != nil {
		return err
	}
	newClass, err = p.storage.FindLoadClassByName(nameId)
	if err != nil {
		return err
	}
	if newClass == nil {
		nextObjectAddress += p.objectAlign
		err = p.addFakeClass(nextObjectAddress, objectClass.ClassObjectId, nameId, nil)
		if err != nil {
			return err
		}
	}

	clsid := 0
	for classId := range p.requiredArrayClasses {
		arrayClass, err := p.storage.FindClass(classId)
		if err != nil {
			return err
		}
		if arrayClass == nil {
			nameId, err = p.getOrCreateText(fmt.Sprintf("unknown-class-%d[]", clsid))
			clsid += 1
			if err != nil {
				return err
			}
			err = p.addFakeClass(classId, objectClass.ClassObjectId, nameId, nil)
			if err != nil {
				return err
			}
		}
	}

	for primitiveType := range p.requiredPrimitiveArrays {
		name := PRIMITIVE_TYPE_ARRAY[primitiveType]
		nameId, err = p.getOrCreateText(name)
		if err != nil {
			return err
		}
		loadClass, err := p.storage.FindLoadClassByName(nameId)
		if err != nil {
			return err
		}
		if loadClass == nil {
			nextObjectAddress += p.objectAlign
			err = p.addFakeClass(nextObjectAddress, objectClass.ClassObjectId, nameId, nil)
			if err != nil {
				return err
			}
			// 更新 instance 里所有这个类型的 cid
			err = p.storage.UpdatePrimitiveArrayClassId(uint64(primitiveType), nextObjectAddress)
			if err != nil {
				return err
			}
		}
	}

	for cid, size := range p.requiredClasses {
		requiredClass, err := p.storage.FindClass(cid)
		if err != nil {
			return err
		}
		if requiredClass == nil {
			if size >= math.MaxInt32 {
				size = 0
			}
			//fieldCount := size/4 + bits.OnesCount(uint(size%4))
			var fields []*model.HProfClassDump_InstanceField
			i := 0
			for ; i < size/4; i++ {
				nameId, err = p.getOrCreateText(fmt.Sprintf("unknown-field-%d", i))
				if err != nil {
					return err
				}
				fields = append(fields, &model.HProfClassDump_InstanceField{
					NameId: nameId,
					Type:   model.HProfValueType_INT,
				})
			}
			if size&2 != 0 {
				nameId, err = p.getOrCreateText(fmt.Sprintf("unknown-field-%d", i))
				i += 1
				if err != nil {
					return err
				}
				fields = append(fields, &model.HProfClassDump_InstanceField{
					NameId: nameId,
					Type:   model.HProfValueType_SHORT,
				})
			}
			if size&1 != 0 {
				nameId, err = p.getOrCreateText(fmt.Sprintf("unknown-field-%d", i))
				i += 1
				if err != nil {
					return err
				}
				fields = append(fields, &model.HProfClassDump_InstanceField{
					NameId: nameId,
					Type:   model.HProfValueType_BYTE,
				})
			}
			className := fmt.Sprintf("unknown-class-%d", clsid)
			clsid += 1
			nameId, err = p.getOrCreateText(className)
			if err != nil {
				return err
			}
			err = p.addFakeClass(cid, objectClass.ClassObjectId, nameId, fields)
		}
	}

	// 更新所有 ClassId == -1 的类的 cid
	err = p.storage.UpdateClassCidToLandClass(langClassId)
	if err != nil {
		return err
	}
	return nil
}

func (p *FirstProcessor) addFakeClass(objectId, superClassObjectId uint64, nameId uint64, fields []*model.HProfClassDump_InstanceField) error {
	newClass := &model.HProfClassDump{
		ClassObjectId:       objectId,
		SuperClassObjectId:  superClassObjectId,
		ClassLoaderObjectId: 0,
		InstanceFields:      fields,
	}
	err := p.storage.SaveInstance(newClass.Type(), int64(objectId), -1, newClass)
	if err != nil {
		return err
	}
	return p.storage.AddLoadClass(objectId, nameId)
}

func (p *FirstProcessor) getOrCreateText(text string) (uint64, error) {
	textId, exist := p.javaClassName[text]
	if exist {
		return textId, nil
	}
	return p.storage.AddText(text)
}

// GetGCRoots 返回 GC Roots
func (p *FirstProcessor) GetGCRoots() map[uint64][]*model.GCRootInfo {
	return p.gcRoots
}

// GetThreadAddressToLocals 返回 GC 相关内容
func (p *FirstProcessor) GetThreadAddressToLocals() map[uint64]map[uint64][]*model.GCRootInfo {
	return p.threadAddressToLocals
}

// getThreads 返回线程信息
func (p *FirstProcessor) getThreads() map[uint32]*model.StackTrace {
	var traces = map[uint32]*model.StackTrace{}
	for k, v := range p.serNum2stackTrace {
		trace := &model.StackTrace{
			ThreadSerialNumber: v.ThreadSerialNumber,
		}
		for _, fid := range v.FrameIds {
			trace.Frames = append(trace.Frames, p.id2frame[fid])
		}
		trace.Locals = p.thread2locals[v.ThreadSerialNumber]
		traces[k] = trace
	}
	return traces
}

func (p *FirstProcessor) parseWholeFile() error {
	err := p.parseHeader()
	if err != nil {
		return err
	}

	//var prev int64
	for {
		r, err := p.parser.ParseRecord()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// TODO 上报进度
		//if pos, err := s.hFile.Seek(0, 1); err == nil && pos-prev > (1<<30) {
		//	log.Printf("currently %d GiB", pos/(1<<30))
		//	prev = pos
		//}

		switch r.(type) {
		case *model.HProfRecordUTF8:
			err = p.onUTF8Record(r.(*model.HProfRecordUTF8))
		case *model.HProfRecordLoadClass:
			err = p.onLoadClassRecord(r.(*model.HProfRecordLoadClass))
		case *model.HProfRecordFrame:
			err = p.onFrameRecord(r.(*model.HProfRecordFrame))
		case *model.HProfRecordTrace:
			err = p.onTraceRecord(r.(*model.HProfRecordTrace))
		case *model.HProfRecordHeapDumpBoundary:
			err = nil
		case *model.HProfClassDump:
			err = p.onClassRecord(r.(*model.HProfClassDump))
		case *model.HProfInstanceDump:
			err = p.onInstanceRecord(r.(*model.HProfInstanceDump))
		case *model.HProfObjectArrayDump:
			err = p.onObjectArrayRecord(r.(*model.HProfObjectArrayDump))
		case *model.HProfPrimitiveArrayDump:
			err = p.onPrimitiveRecord(r.(*model.HProfPrimitiveArrayDump))
		case *model.HProfRootJNIGlobal:
			err = p.onRootJNIGlobalRecord(r.(*model.HProfRootJNIGlobal))
		case *model.HProfRootJNILocal:
			err = p.onRootJNILocalRecord(r.(*model.HProfRootJNILocal))
		case *model.HProfRootJavaFrame:
			err = p.onRootJavaFrameRecord(r.(*model.HProfRootJavaFrame))
		case *model.HProfRootStickyClass:
			err = p.onRootStickyClassRecord(r.(*model.HProfRootStickyClass))
		case *model.HProfRootThreadObj:
			err = p.onRootThreadObjRecord(r.(*model.HProfRootThreadObj))
		case *model.HProfRootMonitorUsed:
			err = p.onRootMonitorUsedRecord(r.(*model.HProfRootMonitorUsed))
		default:
			err = fmt.Errorf("unknown record type: %#v", r)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (p *FirstProcessor) parseHeader() error {
	header, err := p.parser.ParseHeader()
	if err != nil {
		return err
	}
	p.header = header
	return p.storage.PutKV(storage.TIMESTAMP_KEY, header.Timestamp.Unix())
}

func (p *FirstProcessor) onUTF8Record(record *model.HProfRecordUTF8) error {
	text := string(record.Name)
	if text == OBJECT_CLASS_NAME || text == LANG_CLASS_NAME || text == CLASSLOADER_CLASS_NAME {
		p.idx2Name[record.NameId] = text
		p.javaClassName[text] = record.NameId
	}
	return p.storage.SaveText(record.NameId, record.Pos())
}

func (p *FirstProcessor) onLoadClassRecord(record *model.HProfRecordLoadClass) error {
	return p.storage.SaveLoadClass(record.ClassSerialNumber, record.ClassObjectId, record.ClassNameId)
}

func (p *FirstProcessor) onFrameRecord(record *model.HProfRecordFrame) error {
	p.id2frame[record.StackFrameId] = &model.StackFrame{
		FrameId:           record.StackFrameId,
		MethodId:          record.MethodNameId,
		SignatureId:       record.MethodSignatureId,
		SourceFileId:      record.SourceFileNameId,
		ClassSerialNumber: record.ClassSerialNumber,
		Line:              record.LineNumber,
	}
	return nil
}

func (p *FirstProcessor) onTraceRecord(record *model.HProfRecordTrace) error {
	p.serNum2stackTrace[record.StackTraceSerialNumber] = &stackTrace{
		ThreadSerialNumber: record.ThreadSerialNumber,
		FrameIds:           record.StackFrameIds,
	}
	return nil
}

func (p *FirstProcessor) onClassRecord(record *model.HProfClassDump) error {
	if record.SuperClassObjectId != 0 {
		ownFieldsSize := 0
		for _, field := range record.InstanceFields {
			if model.HProfValueType_OBJECT == field.Type {
				ownFieldsSize += int(p.parser.IdSize())
			} else {
				ownFieldsSize += p.parser.ValueSize(field.Type)
			}
		}
		supersize := int(record.InstanceSize) - ownFieldsSize
		if supersize < 0 {
			supersize = 0
		}
		p.reportRequiredClassAndSize(record.SuperClassObjectId, supersize, false)
	}

	// 用于 createRequiredFakeClasses
	if record.ClassObjectId > p.maxObjectId {
		p.maxObjectId = record.ClassObjectId
	}

	// ClassObjectId 写入 DB
	return p.storage.SaveInstance(record.Type(), int64(record.ClassObjectId), -1, record)
}

func (p *FirstProcessor) onInstanceRecord(record *model.HProfInstanceDump) error {
	p.reportRequiredClassAndSize(record.ClassObjectId, len(record.Values), true)

	// 用于 createRequiredFakeClasses
	if record.ObjectId > p.maxObjectId {
		p.maxObjectId = record.ObjectId
	}

	// InstanceId 写入 DB
	return p.storage.SaveInstance(record.Type(), int64(record.ObjectId), int64(record.ClassObjectId), record)
}

func (p *FirstProcessor) onObjectArrayRecord(record *model.HProfObjectArrayDump) error {
	p.requiredArrayClasses[record.ArrayClassObjectId] = true

	// 用于 createRequiredFakeClasses
	if record.ArrayObjectId > p.maxObjectId {
		p.maxObjectId = record.ArrayObjectId
	}

	// InstanceId 写入 DB
	return p.storage.SaveInstance(record.Type(), int64(record.ArrayObjectId), int64(record.ArrayClassObjectId), record)
}

func (p *FirstProcessor) onPrimitiveRecord(record *model.HProfPrimitiveArrayDump) error {
	p.requiredPrimitiveArrays[record.ElementType] = true
	// InstanceId 写入 DB

	// 用于 createRequiredFakeClasses
	if record.ArrayObjectId > p.maxObjectId {
		p.maxObjectId = record.ArrayObjectId
	}

	// 这里用 ElementType 作为 classId，后续再替换
	return p.storage.SaveInstance(record.Type(), int64(record.ArrayObjectId), int64(record.ElementType), record)
}

func (p *FirstProcessor) onRootJNIGlobalRecord(record *model.HProfRootJNIGlobal) error {
	p.addGcRoot(record.ObjectId, 0, model.GCRootType_NATIVE_STATIC)
	return nil
}

func (p *FirstProcessor) onRootJNILocalRecord(record *model.HProfRootJNILocal) error {
	p.addGcRootWithThread(record.ObjectId, record.ThreadSerialNumber, model.GCRootType_NATIVE_LOCAL,
		int32(record.FrameNumberInStackTrace))
	return nil
}

func (p *FirstProcessor) onRootJavaFrameRecord(record *model.HProfRootJavaFrame) error {
	p.addGcRootWithThread(record.ObjectId, record.ThreadSerialNumber, model.GCRootType_JAVA_LOCAL,
		int32(record.FrameNumberInStackTrace))
	return nil
}

func (p *FirstProcessor) onRootStickyClassRecord(record *model.HProfRootStickyClass) error {
	p.addGcRoot(record.ObjectId, 0, model.GCRootType_SYSTEM_CLASS)
	return nil
}

func (p *FirstProcessor) onRootThreadObjRecord(record *model.HProfRootThreadObj) error {
	p.thread2Id[record.ThreadSequenceNumber] = record.ThreadObjectId
	p.addGcRoot(record.ThreadObjectId, 0, model.GCRootType_THREAD_OBJ)
	return nil
}

func (p *FirstProcessor) onRootMonitorUsedRecord(record *model.HProfRootMonitorUsed) error {
	p.addGcRoot(record.ObjectId, 0, model.GCRootType_BUSY_MONITOR)
	return nil
}

func (p *FirstProcessor) addGcRootWithThread(id uint64, threadSerialNumber uint32, typ int, lineNumber int32) {
	threadId, exist := p.thread2Id[threadSerialNumber]
	if exist {
		p.addGcRoot(id, threadId, typ)
	} else {
		p.addGcRoot(id, 0, typ)
	}
	// 记录线程信息
	if lineNumber >= 0 {
		p.thread2locals[threadSerialNumber] = append(p.thread2locals[threadSerialNumber],
			model.NewLocalFrame(id, lineNumber))
	}
}

func (p *FirstProcessor) addGcRoot(id uint64, threadId uint64, typ int) {
	if threadId != 0 {
		localAddressToRootInfo, exist := p.threadAddressToLocals[threadId]
		if !exist {
			localAddressToRootInfo = map[uint64][]*model.GCRootInfo{}
		}
		localAddressToRootInfo[id] = append(localAddressToRootInfo[id], model.NewGcRootInfo(id, threadId, typ))
		p.threadAddressToLocals[threadId] = localAddressToRootInfo
	}
	gcRootInfo := model.NewGcRootInfo(id, threadId, typ)
	p.gcRoots[id] = append(p.gcRoots[id], gcRootInfo)
}

func (p *FirstProcessor) reportRequiredClassAndSize(cid uint64, size int, sizeKnown bool) {
	if _, exist := p.requiredClasses[cid]; !exist {
		p.requiredClasses[cid] = size
	}
	if sizeKnown {
		p.requiredClasses[cid] = size
	}
}

// readObjects 遍历所有 instances
// 这里先从 SQLite 里读取
func (p *FirstProcessor) readObjects() error {
	idx := 0
	err := p.storage.ListHeapObject(func(obj *model.HeapObject) error {
		idx += 1
		err := p.prepareHeapObject(obj)
		if err != nil {
			return err
		}

		// TODO Discarded object
		// 为啥会有找不到的情况？

		// check if some thread to local variables references have to be added
		localVars, exist := p.threadAddressToLocals[obj.Instance.Id()]
		if exist {
			for k, _ := range localVars {
				obj.References = append(obj.References, k)
			}
		}

		// 更新 outbound
		p.outbound[obj.Instance.Id()] = obj.References
		return nil
	})

	for k, v := range p.outbound {
		err = p.storage.SetOutboundById(k, v)
		if err != nil {
			return err
		}
	}
	return err
}

func (p *FirstProcessor) prepareHeapObject(obj *model.HeapObject) error {
	// TODO 更新 usedHeapSize
	obj.References = append(obj.References, obj.Class.ClassObjectId)

	if array, ok := obj.Instance.(*model.HProfObjectArrayDump); ok {
		obj.References = append(obj.References, array.ElementObjectIds...)
	}

	if inst, ok := obj.Instance.(*model.HProfInstanceDump); ok {
		hierarchy, err := p.resolveClassHierarchy(inst.ClassObjectId)
		reader := bytes.NewReader(inst.Values)
		// TODO 准备读取 value
		thisClazz := hierarchy[0]
		obj.Class = thisClazz
		// TODO inst 有可能是一个 Class？
		objcl, err := p.storage.FindClass(inst.ObjectId)
		if err != nil {
			return err
		}
		if objcl != nil {
			obj.References = append(obj.References, objcl.GetReferences()...)
		} else {
			// TODO usedHeapSize
			obj.References = append(obj.References, thisClazz.ClassObjectId)
		}

		for _, class := range hierarchy {
			for _, field := range class.InstanceFields {
				typ := field.Type
				// TODO Find match for pseudo-statics

				if typ == model.HProfValueType_OBJECT {
					// TODO 更新 pseudo-statics field 的值
					refId, err := p.parser.ReadIDFromReader(int(p.header.IdentifierSize), reader)
					if err != nil {
						return err
					}
					if refId != 0 {
						obj.References = append(obj.References, refId)
					}
				} else {
					_, err := p.parser.ReadValueFromReader(typ, reader)
					if err != nil {
						return err
					}
					// TODO 更新 pseudo-statics field 的值
				}
			}
		}
	}

	return nil
}

func (p *FirstProcessor) resolveClassHierarchy(cid uint64) ([]*model.HProfClassDump, error) {
	cached, exist := p.classHierarchyCache[cid]
	if exist {
		return cached, nil
	}

	cached = []*model.HProfClassDump{}
	class, err := p.storage.FindClass(cid)
	if err != nil {
		return nil, err
	}
	cached = append(cached, class)
	for class.SuperClassObjectId != 0 {
		class, err = p.storage.FindClass(class.SuperClassObjectId)
		if err != nil {
			return nil, err
		}
		cached = append(cached, class)
	}
	p.classHierarchyCache[cid] = cached
	return cached, nil
}
