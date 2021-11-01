package indexer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hprof-tool/pkg/model"
	"hprof-tool/pkg/parser"
	global "hprof-tool/pkg/util"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	keyIdSize                = "idsize"
	keyPrefixString          = "string-"
	keyPrefixLoadedClass     = "loadedclass-"
	keyPrefixLoadedClassIdx  = "loadedclassidx-"
	keyPrefixLoadedClassName = "loadedclassname-"
	keyPrefixFrame           = "frame-"
	keyPrefixTrace           = "trace-"
	keyPrefixClass           = "class-"
	keyPrefixId2Key          = "id2key-"
	keyPrefixInstance        = "instance-"
	keyPrefixObjectArray     = "objectarray-"
	keyPrefixPrimitiveArray  = "primitivearray-"
	keyPrefixRootJNIGlobal   = "rootjniglobal-"
	keyPrefixRootJNILocal    = "rootjnilocal-"
	keyPrefixRootJavaFrame   = "rootjavaframe-"
	keyPrefixRootStickyClass = "rootstickyclass-"
	keyPrefixRootThreadObj   = "rootthreadobj-"
	keyPrefixRootMonitorUsed = "rootmonitorused-"
)

// OpenOrCreateIndex opens or creates a DB based on the HProf file.
func OpenOrCreateIndex(heapFilePath, indexFilePath string) (*Indexer, error) {
	if _, err := os.Stat(indexFilePath); os.IsNotExist(err) {
		if err := createIndex(heapFilePath, indexFilePath); err != nil {
			return nil, err
		}
	}

	db, err := leveldb.OpenFile(indexFilePath, nil)
	if err != nil {
		return nil, err
	}
	val, err := db.Get([]byte(keyIdSize), nil)
	if err != nil {
		return nil, err
	}

	global.ID_SIZE = int(val[0])
	return &Indexer{db: db}, nil
}

func createIndex(heapFilePath, indexFilePath string) error {
	f, err := os.Open(heapFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	db, err := leveldb.OpenFile(indexFilePath, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	p := parser.NewParser(f)
	_, err = p.ParseHeader()
	if err != nil {
		return err
	}

	cs := &counters{}
	var prev int64
	stringMap := make(map[uint64][]byte)
	batch := new(leveldb.Batch)
	batch.Put([]byte(keyIdSize), append([]byte(""),  p.IdSize()))
	for {
		r, err := p.ParseRecord()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if pos, err := f.Seek(0, 1); err == nil && pos-prev > (1<<30) {
			log.Printf("currently %d GiB", pos/(1<<30))
			prev = pos
		}

		if err := addRecordToDB(batch, cs, r, stringMap); err != nil {
			return err
		}
		if batch.Len() > 100000 {
			if err := db.Write(batch, nil); err != nil {
				return err
			}
			batch.Reset()
		}
	}
	if batch.Len() > 0 {
		if err := db.Write(batch, nil); err != nil {
			return err
		}
	}
	return nil
}

type counters struct {
	countJNIGlobal   uint64
	countJNILocal    uint64
	countJavaFrame   uint64
	countStickyClass uint64
	countThreadObj   uint64
	countMonitorUsed uint64
}

func addRecordToDB(batch *leveldb.Batch, cs *counters, record interface{}, stringMap map[uint64][]byte) error {
	switch record.(type) {
	case *model.HProfRecordUTF8:
		return addRecordUTF8(batch, cs, record, stringMap)
	case *model.HProfRecordLoadClass:
		return addRecordLoadClass(batch, cs, record, stringMap)
	case *model.HProfRecordFrame:
		return addRecordFrame(batch, cs, record)
	case *model.HProfRecordTrace:
		return addRecordTrace(batch, cs, record)
	case *model.HProfRecordHeapDumpBoundary:
		return nil
	case *model.HProfClassDump:
		return addRecordClassDump(batch, cs, record)
	case *model.HProfInstanceDump:
		return addRecordInstanceDump(batch, cs, record)
	case *model.HProfObjectArrayDump:
		return addRecordObjectArrayDump(batch, cs, record)
	case *model.HProfPrimitiveArrayDump:
		return addRecordPrimitiveArrayDump(batch, cs, record)
	case *model.HProfRootJNIGlobal:
		return addRecordRootJNIGlobal(batch, cs, record)
	case *model.HProfRootJNILocal:
		return addRecordRootJNILocal(batch, cs, record)
	case *model.HProfRootJavaFrame:
		return addRecordRootJavaFrame(batch, cs, record)
	case *model.HProfRootStickyClass:
		return addRecordRootStickyClass(batch, cs, record)
	case *model.HProfRootThreadObj:
		return addRecordRootThreadObj(batch, cs, record)
	case *model.HProfRootMonitorUsed:
		return addRecordRootMonitorUsed(batch, cs, record)
	default:
		return fmt.Errorf("unknown record type: %#v", record)
	}
}

func addRecordUTF8(batch *leveldb.Batch, cs *counters, record interface{}, stringMap map[uint64][]byte) error {
	utf8 := record.(*model.HProfRecordUTF8)
	key := utf8.GetNameId()
	bs, err := json.Marshal(record)
	if err != nil {
		return err
	}
	batch.Put(createKey(keyPrefixString, key), bs)
	stringMap[key] = utf8.Name
	return nil
}

func addRecordLoadClass(batch *leveldb.Batch, cs *counters, record interface{}, stringMap map[uint64][]byte) error {
	bs, err := json.Marshal(record)
	if err != nil {
		return err
	}
	loadedClass := record.(*model.HProfRecordLoadClass)
	name := stringMap[loadedClass.GetClassNameId()]
	key := uint64(loadedClass.GetClassSerialNumber())
	loadedClassKey := createKey(keyPrefixLoadedClass, key)
	batch.Put(loadedClassKey, bs)
	batch.Put(createKey(keyPrefixLoadedClassIdx, loadedClass.GetClassObjectId()), loadedClassKey)
	nameKey := append([]byte(keyPrefixLoadedClassName), name...)
	batch.Put(nameKey, loadedClassKey)
	return nil
}

func addRecordFrame(batch *leveldb.Batch, cs *counters, record interface{}) error {
	key := (record.(*model.HProfRecordFrame)).GetStackFrameId()
	bs, err := json.Marshal(record)
	if err != nil {
		return err
	}
	batch.Put(createKey(keyPrefixFrame, key), bs)
	return nil
}

func addRecordTrace(batch *leveldb.Batch, cs *counters, record interface{}) error {
	key := uint64((record.(*model.HProfRecordTrace)).GetStackTraceSerialNumber())
	//key := uint64((record.(*model.HProfRecordTrace)).GetThreadSerialNumber())
	bs, err := json.Marshal(record)
	if err != nil {
		return err
	}
	batch.Put(createKey(keyPrefixTrace, key), bs)
	return nil
}

func addRecordClassDump(batch *leveldb.Batch, cs *counters, record interface{}) error {
	key := (record.(*model.HProfClassDump)).GetClassObjectId()
	instanceKey := createKey(keyPrefixClass, key)
	bs, err := json.Marshal(record)
	if err != nil {
		return err
	}
	batch.Put(instanceKey, bs)
	instanceKey = append(instanceKey, 'c')
	batch.Put(createKey(keyPrefixId2Key, key), instanceKey)
	return nil
}

func addRecordInstanceDump(batch *leveldb.Batch, cs *counters, record interface{}) error {
	key := (record.(*model.HProfInstanceDump)).GetObjectId()
	instanceKey := createKey(keyPrefixInstance, key)
	bs, err := json.Marshal(record)
	if err != nil {
		return err
	}
	batch.Put(instanceKey, bs)
	instanceKey = append(instanceKey, 'i')
	batch.Put(createKey(keyPrefixId2Key, key), instanceKey)
	return nil
}

func addRecordObjectArrayDump(batch *leveldb.Batch, cs *counters, record interface{}) error {
	key := (record.(*model.HProfObjectArrayDump)).GetArrayObjectId()
	instanceKey := createKey(keyPrefixObjectArray, key)
	bs, err := json.Marshal(record)
	if err != nil {
		return err
	}
	batch.Put(instanceKey, bs)
	instanceKey = append(instanceKey, 'o')
	batch.Put(createKey(keyPrefixId2Key, key), instanceKey)
	return nil
}

func addRecordPrimitiveArrayDump(batch *leveldb.Batch, cs *counters, record interface{}) error {
	key := (record.(*model.HProfPrimitiveArrayDump)).GetArrayObjectId()
	instanceKey := createKey(keyPrefixPrimitiveArray, key)
	bs, err := json.Marshal(record)
	if err != nil {
		return err
	}
	batch.Put(instanceKey, bs)
	instanceKey = append(instanceKey, 'p')
	batch.Put(createKey(keyPrefixId2Key, key), instanceKey)
	return nil
}

func addRecordRootJNIGlobal(batch *leveldb.Batch, cs *counters, record interface{}) error {
	key := cs.countJNIGlobal
	cs.countJNIGlobal++
	bs, err := json.Marshal(record)
	if err != nil {
		return err
	}
	batch.Put(createKey(keyPrefixRootJNIGlobal, key), bs)
	return nil
}

func addRecordRootJNILocal(batch *leveldb.Batch, cs *counters, record interface{}) error {
	key := cs.countJNILocal
	cs.countJNILocal++
	bs, err := json.Marshal(record)
	if err != nil {
		return err
	}
	batch.Put(createKey(keyPrefixRootJNILocal, key), bs)
	return nil
}

func addRecordRootJavaFrame(batch *leveldb.Batch, cs *counters, record interface{}) error {
	key := cs.countJavaFrame
	cs.countJavaFrame++
	bs, err := json.Marshal(record)
	if err != nil {
		return err
	}
	batch.Put(createKey(keyPrefixRootJavaFrame, key), bs)
	return nil
}

func addRecordRootStickyClass(batch *leveldb.Batch, cs *counters, record interface{}) error {
	key := cs.countStickyClass
	cs.countStickyClass++
	bs, err := json.Marshal(record)
	if err != nil {
		return err
	}
	batch.Put(createKey(keyPrefixRootStickyClass, key), bs)
	return nil
}

func addRecordRootThreadObj(batch *leveldb.Batch, cs *counters, record interface{}) error {
	key := cs.countThreadObj
	cs.countThreadObj++
	bs, err := json.Marshal(record)
	if err != nil {
		return err
	}
	batch.Put(createKey(keyPrefixRootThreadObj, key), bs)
	return nil
}

func addRecordRootMonitorUsed(batch *leveldb.Batch, cs *counters, record interface{}) error {
	key := cs.countMonitorUsed
	cs.countMonitorUsed++
	bs, err := json.Marshal(record)
	if err != nil {
		return err
	}
	batch.Put(createKey(keyPrefixRootMonitorUsed, key), bs)
	return nil
}

func createKey(prefix string, id uint64) []byte {
	return []byte(prefix + strconv.FormatUint(id, 16))
}

// Index is indexed HProf data.
type Indexer struct {
	db *leveldb.DB
}

func (idx *Indexer) loadProto(prefix string, id uint64, m interface{}) error {
	return idx.loadProtoByKey(createKey(prefix, id), m)
}

func (idx *Indexer) loadProtoByKey(key []byte, m interface{}) error {
	bs, err := idx.db.Get(key, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, m)
}

func (idx *Indexer) loadIndex(key []byte) ([]byte, error) {
	return idx.db.Get(key, nil)
}

func (idx *Indexer) GetById(oid uint64) (model.HProfDumpWithSize, error) {
	key, err := idx.db.Get(createKey(keyPrefixId2Key, oid), nil)
	if err != nil {
		return nil, err
	}
	klen := len(key) - 1
	switch key[klen] {
	case 'c':
		var d model.HProfClassDump
		err := idx.loadProtoByKey(key[:klen], &d)
		if err != nil {
			return nil, err
		}
		return &d, nil
	case 'i':
		var d model.HProfInstanceDump
		err := idx.loadProtoByKey(key[:klen], &d)
		if err != nil {
			return nil, err
		}
		return &d, nil
	case 'o':
		var d model.HProfObjectArrayDump
		err := idx.loadProtoByKey(key[:klen], &d)
		if err != nil {
			return nil, err
		}
		return &d, nil
	case 'p':
		var d model.HProfPrimitiveArrayDump
		err := idx.loadProtoByKey(key[:klen], &d)
		if err != nil {
			return nil, err
		}
		return &d, nil
	default:
		return nil, fmt.Errorf("unknown instance id & key: %d %s", oid, key)
	}
}

// String returns a name based on a name ID.
func (idx *Indexer) String(nameID uint64) (string, error) {
	var d model.HProfRecordUTF8
	if err := idx.loadProto(keyPrefixString, nameID, &d); err != nil {
		return "", err
	}
	return string(d.GetName()), nil
}

// LoadedClass returns a HProfRecordLoadClass based on a class serial number.
func (idx *Indexer) LoadedClass(classSerialNumber uint32) (*model.HProfRecordLoadClass, error) {
	var d model.HProfRecordLoadClass
	if err := idx.loadProto(keyPrefixLoadedClass, uint64(classSerialNumber), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// LoadedClassByOId returns a HProfRecordLoadClass based on a class serial number.
func (idx *Indexer) LoadedClassByOID(classObjectId uint64) (*model.HProfRecordLoadClass, error) {
	indexKey := createKey(keyPrefixLoadedClassIdx, classObjectId)
	loadedClassKey, err := idx.loadIndex(indexKey)
	if err != nil {
		return nil, err
	}
	var d model.HProfRecordLoadClass
	if err := idx.loadProtoByKey(loadedClassKey, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// LoadedClassByOId returns a HProfRecordLoadClass based on a class serial number.
func (idx *Indexer) LoadedClassByName(name string) (*model.HProfRecordLoadClass, error) {
	indexKey := []byte(keyPrefixLoadedClassName + name)
	loadedClassKey, err := idx.loadIndex(indexKey)
	if err != nil {
		return nil, err
	}
	var d model.HProfRecordLoadClass
	if err := idx.loadProtoByKey(loadedClassKey, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// Frame returns a HProfRecordFrame based on a stack frame ID.
func (idx *Indexer) Frame(stackFrameID uint64) (*model.HProfRecordFrame, error) {
	var d model.HProfRecordFrame
	if err := idx.loadProto(keyPrefixFrame, stackFrameID, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// Trace returns a HProfRecordTrace based on a stack trace serial number.
func (idx *Indexer) Trace(stackTraceSerialNumber uint64) (*model.HProfRecordTrace, error) {
	var d model.HProfRecordTrace
	if err := idx.loadProto(keyPrefixTrace, stackTraceSerialNumber, &d); err != nil {
		return nil, err
	}
	return &d, nil
}
//func (idx *Indexer) Trace(threadSerialNumber uint64) (*model.HProfRecordTrace, error) {
//	var d model.HProfRecordTrace
//	if err := idx.loadProto(keyPrefixTrace, threadSerialNumber, &d); err != nil {
//		return nil, err
//	}
//	return &d, nil
//}

// Class returns a HProfClassDump based on a class object ID.
func (idx *Indexer) Class(classObjectID uint64) (*model.HProfClassDump, error) {
	var d model.HProfClassDump
	if err := idx.loadProto(keyPrefixClass, classObjectID, &d); err != nil {
		return nil, err
	}
	return &d, nil
}
func (idx *Indexer) ClassBySerialNumber(classSerialNumber uint32) (*model.HProfClassDump, error) {
	loaded, err := idx.LoadedClass(classSerialNumber)
	if err != nil {
		return nil, err
	}
	return idx.Class(loaded.ClassObjectId)
}

// LoadedClassByOId returns a HProfRecordLoadClass based on a class serial number.
func (idx *Indexer) ClassByName(name string) (*model.HProfClassDump, error) {
	loaded, err := idx.LoadedClassByName(name)
	if err != nil {
		return nil, err
	}
	return idx.Class(loaded.ClassObjectId)
}

func (idx *Indexer) ClassName(classObjectID uint64) (string, error) {
	loaded, err := idx.LoadedClassByOID(classObjectID)
	if err != nil {
		return "", err
	}
	return idx.String(loaded.GetClassNameId())
}

func (idx *Indexer) ClassNameById(id uint64) (string, error) {
	dump, err := idx.GetById(id)
	if err != nil {
		return "", err
	}
	classId := dump.GetClassObjectId()
	return idx.ClassName(classId)
}

// ForEachInstance iterates through all HProfInstanceDump objects.
func (idx *Indexer) ForEachClass(fn func(dump *model.HProfClassDump) error) error {
	iter := idx.db.NewIterator(util.BytesPrefix([]byte(keyPrefixClass)), nil)
	defer iter.Release()
	for iter.Next() {
		var d model.HProfClassDump
		if err := json.Unmarshal(iter.Value(), &d); err != nil {
			return err
		}
		if err := fn(&d); err != nil {
			return err
		}
	}
	return iter.Error()
}

// Instance returns a HProfInstanceDump based on an object ID.
func (idx *Indexer) Instance(objectID uint64) (*model.HProfInstanceDump, error) {
	var d model.HProfInstanceDump
	if err := idx.loadProto(keyPrefixInstance, objectID, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// ForEachInstance iterates through all HProfInstanceDump objects.
func (idx *Indexer) ForEachInstance(fn func(*model.HProfInstanceDump) error) error {
	iter := idx.db.NewIterator(util.BytesPrefix([]byte(keyPrefixInstance)), nil)
	defer iter.Release()
	for iter.Next() {
		var d model.HProfInstanceDump
		if err := json.Unmarshal(iter.Value(), &d); err != nil {
			return err
		}
		if err := fn(&d); err != nil {
			return err
		}
	}
	return iter.Error()
}

// ObjectArray returns a HProfObjectArrayDump based on an array object ID.
func (idx *Indexer) ObjectArray(arrayObjectID uint64) (*model.HProfObjectArrayDump, error) {
	var d model.HProfObjectArrayDump
	if err := idx.loadProto(keyPrefixObjectArray, arrayObjectID, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// ForEachObjectArray iterates through all HProfObjectArrayDump objects.
func (idx *Indexer) ForEachObjectArray(fn func(*model.HProfObjectArrayDump) error) error {
	iter := idx.db.NewIterator(util.BytesPrefix([]byte(keyPrefixObjectArray)), nil)
	defer iter.Release()
	for iter.Next() {
		var d model.HProfObjectArrayDump
		if err := json.Unmarshal(iter.Value(), &d); err != nil {
			return err
		}
		if err := fn(&d); err != nil {
			return err
		}
	}
	return iter.Error()
}

// PrimitiveArray returns a HProfPrimitiveArrayDump based on an array object ID.
func (idx *Indexer) PrimitiveArray(arrayObjectID uint64) (*model.HProfPrimitiveArrayDump, error) {
	var d model.HProfPrimitiveArrayDump
	if err := idx.loadProto(keyPrefixPrimitiveArray, arrayObjectID, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// ForEachPrimitiveArray iterates through all HProfPrimitiveArrayDump objects.
func (idx *Indexer) ForEachPrimitiveArray(fn func(*model.HProfPrimitiveArrayDump) error) error {
	iter := idx.db.NewIterator(util.BytesPrefix([]byte(keyPrefixPrimitiveArray)), nil)
	defer iter.Release()
	for iter.Next() {
		var d model.HProfPrimitiveArrayDump
		if err := json.Unmarshal(iter.Value(), &d); err != nil {
			return err
		}
		if err := fn(&d); err != nil {
			return err
		}
	}
	return iter.Error()
}

// ForEachRootJNIGlobal iterates through all HProfRootJNIGlobal objects.
func (idx *Indexer) ForEachRootJNIGlobal(fn func(*model.HProfRootJNIGlobal) error) error {
	iter := idx.db.NewIterator(util.BytesPrefix([]byte(keyPrefixRootJNIGlobal)), nil)
	defer iter.Release()
	for iter.Next() {
		var d model.HProfRootJNIGlobal
		if err := json.Unmarshal(iter.Value(), &d); err != nil {
			return err
		}
		if err := fn(&d); err != nil {
			return err
		}
	}
	return iter.Error()
}

// ForEachRootJNILocal iterates through all HProfRootJNILocal objects.
func (idx *Indexer) ForEachRootJNILocal(fn func(*model.HProfRootJNILocal) error) error {
	iter := idx.db.NewIterator(util.BytesPrefix([]byte(keyPrefixRootJNILocal)), nil)
	defer iter.Release()
	for iter.Next() {
		var d model.HProfRootJNILocal
		if err := json.Unmarshal(iter.Value(), &d); err != nil {
			return err
		}
		if err := fn(&d); err != nil {
			return err
		}
	}
	return iter.Error()
}

// ForEachRootJavaFrame iterates through all HProfRootJavaFrame objects.
func (idx *Indexer) ForEachRootJavaFrame(fn func(*model.HProfRootJavaFrame) error) error {
	iter := idx.db.NewIterator(util.BytesPrefix([]byte(keyPrefixRootJavaFrame)), nil)
	defer iter.Release()
	for iter.Next() {
		var d model.HProfRootJavaFrame
		if err := json.Unmarshal(iter.Value(), &d); err != nil {
			return err
		}
		if err := fn(&d); err != nil {
			return err
		}
	}
	return iter.Error()
}

// ForEachRootStickyClass iterates through all HProfRootStickyClass objects.
func (idx *Indexer) ForEachRootStickyClass(fn func(*model.HProfRootStickyClass) error) error {
	iter := idx.db.NewIterator(util.BytesPrefix([]byte(keyPrefixRootStickyClass)), nil)
	defer iter.Release()
	for iter.Next() {
		var d model.HProfRootStickyClass
		if err := json.Unmarshal(iter.Value(), &d); err != nil {
			return err
		}
		if err := fn(&d); err != nil {
			return err
		}
	}
	return iter.Error()
}

// ForEachRootThreadObj iterates through all HProfRootThreadObj objects.
func (idx *Indexer) ForEachRootThreadObj(fn func(*model.HProfRootThreadObj) error) error {
	iter := idx.db.NewIterator(util.BytesPrefix([]byte(keyPrefixRootThreadObj)), nil)
	defer iter.Release()
	for iter.Next() {
		var d model.HProfRootThreadObj
		if err := json.Unmarshal(iter.Value(), &d); err != nil {
			return err
		}
		if err := fn(&d); err != nil {
			return err
		}
	}
	return iter.Error()
}

// ForEachRootMonitorUsed iterates through all HProfRootMonitorUsed objects.
func (idx *Indexer) ForEachRootMonitorUsed(fn func(*model.HProfRootMonitorUsed) error) error {
	iter := idx.db.NewIterator(util.BytesPrefix([]byte(keyPrefixRootMonitorUsed)), nil)
	defer iter.Release()
	for iter.Next() {
		var d model.HProfRootMonitorUsed
		if err := json.Unmarshal(iter.Value(), &d); err != nil {
			return err
		}
		if err := fn(&d); err != nil {
			return err
		}
	}
	return iter.Error()
}


func (idx *Indexer) FindInstanceObjectReference(instance *model.HProfInstanceDump) ([]*model.HProfClassDump_InstanceField, error) {
	class, err := idx.Class(instance.GetClassObjectId())
	if err != nil {
		return nil, err
	}
	fields, err := idx.findAllInstanceFields(class)
	if err != nil {
		return nil, err
	}
	var objectFields []*model.HProfClassDump_InstanceField
	offset := 0
	for _, field := range fields {
		if field.Type == model.HProfValueType_OBJECT {
			bs := instance.Values[offset:offset + global.ID_SIZE]
			field.Value = binary.BigEndian.Uint64(bs)
			objectFields = append(objectFields, field)
			offset += global.ID_SIZE
		} else {
			offset += parser.ValueSize[field.Type]
		}
	}

	return objectFields, nil
}

func (idx *Indexer) findAllInstanceFields(class *model.HProfClassDump) ([]*model.HProfClassDump_InstanceField, error) {
	var fields []*model.HProfClassDump_InstanceField
	for _, field := range class.InstanceFields {
		fields = append(fields, &model.HProfClassDump_InstanceField{
			NameId: field.NameId,
			Type: field.Type,
		})
	}
	sid := class.GetSuperClassObjectId()
	if sid <= 0 {
		return fields, nil
	}
	super, err := idx.Class(sid)
	if err != nil {
		return nil, err
	}
	if super != nil {
		superFields, err := idx.findAllInstanceFields(super)
		if err != nil {
			return nil, err
		}
		fields = append(fields, superFields...)
	}
	return fields, nil
}