package indexer

import (
	"fmt"
	"hprof-tool/pkg/hprof"
	"io"
)

func (i *Indexer) CreateIndex() error {
	err := i.hreader.ParseHeader()
	if err != nil {
		return err
	}

	for {
		r, err := i.hreader.ParseRecord()
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
		case *hprof.HProfUTF8Record:
			err = i.onUTF8Record(r.(*hprof.HProfUTF8Record))
		case *hprof.HProfLoadClassRecord:
			err = i.onLoadClassRecord(r.(*hprof.HProfLoadClassRecord))
		case *hprof.HProfFrameRecord:
			err = i.onFrameRecord(r.(*hprof.HProfFrameRecord))
		case *hprof.HProfTraceRecord:
			err = i.onTraceRecord(r.(*hprof.HProfTraceRecord))
		case *hprof.HProfThreadRecord:
			err = i.onThreadRecord(r.(*hprof.HProfThreadRecord))
		case *hprof.HProfClassRecord:
			err = i.onClassRecord(r.(*hprof.HProfClassRecord))
		case *hprof.HProfInstanceRecord:
			err = i.onInstanceRecord(r.(*hprof.HProfInstanceRecord))
		case *hprof.HProfObjectArrayRecord:
			err = i.onObjectArrayRecord(r.(*hprof.HProfObjectArrayRecord))
		case *hprof.HProfPrimitiveArrayRecord:
			err = i.onPrimitiveArrayRecord(r.(*hprof.HProfPrimitiveArrayRecord))
		case *hprof.HProfRootJNIGlobal:
			err = i.onRootJNIGlobalRecord(r.(*hprof.HProfRootJNIGlobal))
		case *hprof.HProfRootJNILocal:
			err = i.onRootJNILocalRecord(r.(*hprof.HProfRootJNILocal))
		case *hprof.HProfRootJavaFrame:
			err = i.onRootJavaFrameRecord(r.(*hprof.HProfRootJavaFrame))
		case *hprof.HProfRootStickyClass:
			err = i.onRootStickyClassRecord(r.(*hprof.HProfRootStickyClass))
		case *hprof.HProfRootThreadObj:
			err = i.onRootThreadObjRecord(r.(*hprof.HProfRootThreadObj))
		case *hprof.HProfRootMonitorUsed:
			err = i.onRootMonitorUsedRecord(r.(*hprof.HProfRootMonitorUsed))
		case *hprof.HProfRecordHeapDumpBoundary:
			err = nil
		default:
			err = fmt.Errorf("unknown record type: %#v", r)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Indexer) onUTF8Record(record *hprof.HProfUTF8Record) error {
	// text := string(record.Name)
	// if text == OBJECT_CLASS_NAME || text == LANG_CLASS_NAME || text == CLASSLOADER_CLASS_NAME {
	// 	_, err := i.storage.AddText(text)
	// 	return err
	// }
	pos, _ := record.PosAndSize()
	return i.storage.SaveText(record.NameId, pos)
}

func (i *Indexer) onLoadClassRecord(record *hprof.HProfLoadClassRecord) error {
	return i.storage.SaveLoadClass(record.ClassSerialNumber, record.ClassObjectId, record.ClassNameId)
}

func (i *Indexer) onFrameRecord(record *hprof.HProfFrameRecord) error {
	return i.storage.SaveThreadFrame(record)
}

func (i *Indexer) onTraceRecord(record *hprof.HProfTraceRecord) error {
	return i.storage.SaveThreadTrace(record)
}

func (i *Indexer) onThreadRecord(record *hprof.HProfThreadRecord) error {
	return i.storage.SaveThread(record)
}

func (i *Indexer) onClassRecord(record *hprof.HProfClassRecord) error {
	// ClassObjectId 写入 DB
	pos, _ := record.PosAndSize()
	return i.storage.SaveClass(pos, int64(record.ClassObjectId), int(record.InstanceSize))
}

func (i *Indexer) onInstanceRecord(r *hprof.HProfInstanceRecord) error {
	pos, _ := r.PosAndSize()
	size := len(r.Values) + 16
	return i.storage.SaveInstance(pos, int64(r.ObjectId), int64(r.ClassObjectId), size)
}

func (i *Indexer) onObjectArrayRecord(record *hprof.HProfObjectArrayRecord) error {
	pos, _ := record.PosAndSize()
	size := len(record.ElementObjectIds)*8 + 16
	return i.storage.SaveObjectArray(pos, int64(record.ArrayObjectId), int64(record.ArrayClassObjectId), size)
}

func (i *Indexer) onPrimitiveArrayRecord(record *hprof.HProfPrimitiveArrayRecord) error {
	// 这里用 ElementType 作为 classId，后续再替换
	pos, _ := record.PosAndSize()
	return i.storage.SavePrimitiveArray(pos, int64(record.ArrayObjectId), int64(record.ElementType), len(record.Values))
}

func (i *Indexer) onRootJNIGlobalRecord(record *hprof.HProfRootJNIGlobal) error {
	pos, _ := record.PosAndSize()
	return i.storage.SaveGCRoot(hprof.GCRootType_NATIVE_STATIC, pos)
}

func (i *Indexer) onRootJNILocalRecord(record *hprof.HProfRootJNILocal) error {
	pos, _ := record.PosAndSize()
	return i.storage.SaveGCRoot(hprof.GCRootType_NATIVE_LOCAL, pos)
}

func (i *Indexer) onRootJavaFrameRecord(record *hprof.HProfRootJavaFrame) error {
	pos, _ := record.PosAndSize()
	return i.storage.SaveGCRoot(hprof.GCRootType_JAVA_LOCAL, pos)
}

func (i *Indexer) onRootStickyClassRecord(record *hprof.HProfRootStickyClass) error {
	pos, _ := record.PosAndSize()
	return i.storage.SaveGCRoot(hprof.GCRootType_SYSTEM_CLASS, pos)
}

func (i *Indexer) onRootThreadObjRecord(record *hprof.HProfRootThreadObj) error {
	pos, _ := record.PosAndSize()
	return i.storage.SaveGCRoot(hprof.GCRootType_THREAD_OBJ, pos)
}

func (i *Indexer) onRootMonitorUsedRecord(record *hprof.HProfRootMonitorUsed) error {
	pos, _ := record.PosAndSize()
	return i.storage.SaveGCRoot(hprof.GCRootType_BUSY_MONITOR, pos)
}
