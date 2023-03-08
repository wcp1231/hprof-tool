package storage

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"hprof-tool/pkg/model"
	"strings"
)

// TODO 只有字符串内存去 hprof 里读取

var schema = strings.ReplaceAll(`
CREATE TABLE IF NOT EXISTS kvs (
    'key' TEXT PRIMARY KEY,
    'value' BLOB NOT NULL
);
CREATE TABLE IF NOT EXISTS texts (
    id INTEGER PRIMARY KEY,
    -- 文件位置，如果是 -1 则表示新增文本
    pos INTEGER NOT NULL,
    -- 新增文本内容
    txt TEXT 
);
CREATE TABLE IF NOT EXISTS load_classes (
    id INTEGER PRIMARY KEY,
    classId INTEGER NOT NULL,
    nameId INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS instances (
    id INTEGER PRIMARY KEY,
    -- 类型
    'type' INTEGER NOT NULL,
    -- ObjectId
    oid INTEGER NOT NULL,
    -- ClassId
    cid INTEGER NOT NULL, 
    -- 原始数据
    'raw' BLOB NOT NULL,
    outbound BLOB,
    inbound BLOB
);
CREATE INDEX instances_type_id_idx ON instances ('type', 'oid');
`, "'", "`")

type SqliteStorage struct {
	db *sql.DB
}

func NewSqliteStorage(dbFile string) (*SqliteStorage, error) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		return nil, err
	}

	return &SqliteStorage{db: db}, nil
}

func (s *SqliteStorage) Init() error {
	result, err := s.db.Exec(schema)
	if err != nil {
		return err
	}
	insertId, err := result.LastInsertId()
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	fmt.Printf("%d %d\n", insertId, rowsAffected) // TODO

	return nil
}

func (s *SqliteStorage) Close() error {
	return s.db.Close()
}

// SaveText 记录文本索引
func (s *SqliteStorage) SaveText(id uint64, pos int64) error {
	_, err := s.db.Exec("INSERT INTO texts (id, pos) VALUES (?, ?)", id, pos)
	return err
}

// AddText 新增文本
func (s *SqliteStorage) AddText(txt string) (uint64, error) {
	_, err := s.db.Exec("INSERT INTO texts (pos, txt) VALUES (?, ?)", -1, txt)
	if err != nil {
		return 0, err
	}
	row := s.db.QueryRow("SELECT last_insert_rowid();")
	var lastId uint64 = 0
	err = row.Scan(&lastId)
	return lastId, err
}

func (s *SqliteStorage) UpdateTextAndPos(id uint32, pos int64, txt string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO texts VALUES (?, ?, ?)", id, pos, txt)
	return err
}

func (s *SqliteStorage) SaveLoadClass(id uint32, classId uint64, nameId uint64) error {
	_, err := s.db.Exec("INSERT INTO load_classes VALUES (?, ?, ?)", id, classId, nameId)
	return err
}

func (s *SqliteStorage) AddLoadClass(classId uint64, nameId uint64) error {
	_, err := s.db.Exec("INSERT INTO load_classes (classId, nameId) VALUES (?, ?)", classId, nameId)
	return err
}

// SaveInstance 记录 Instances 索引
// value 部分根据不同类型存不同的值
func (s *SqliteStorage) SaveInstance(typ model.HProfRecordType, oid, cid int64, value interface{}) error {
	body := encodeGob(value)
	_, err := s.db.Exec("INSERT INTO instances (`type`, oid, cid, `raw`) VALUES (?, ?, ?, ?)",
		typ, oid, cid, body)
	return err
}

// UpdateClassClassId 将 Class 的 classId 改成生成的虚拟类
func (s *SqliteStorage) UpdateClassCidToLandClass(classId uint64) error {
	_, err := s.db.Exec("UPDATE instances SET cid = ? WHERE `type` = ? AND cid = -1",
		classId, model.HProfHDRecordTypeClassDump)
	return err
}

// UpdatePrimitiveArrayClassId 将 PrimitiveArray 的 classId 改成生成的虚拟类
func (s *SqliteStorage) UpdatePrimitiveArrayClassId(elementType, classId uint64) error {
	_, err := s.db.Exec("UPDATE instances SET cid = ? WHERE `type` = ? AND cid = ?",
		classId, model.HProfHDRecordTypePrimitiveArrayDump, elementType)
	return err
}

func (s *SqliteStorage) PutKV(key string, value interface{}) error {
	body := encodeGob(value)
	_, err := s.db.Exec("INSERT OR REPLACE INTO kvs VALUES (?, ?)", key, body)
	return err
}

func (s *SqliteStorage) FindLoadClassByName(nameId uint64) (*model.HProfRecordLoadClass, error) {
	row := s.db.QueryRow("SELECT id, classId, nameId FROM load_classes WHERE nameId = ?", nameId)
	result := &model.HProfRecordLoadClass{}
	err := row.Scan(&result.ClassSerialNumber, &result.ClassObjectId, &result.ClassNameId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func (s *SqliteStorage) FindClass(oid uint64) (*model.HProfClassDump, error) {
	row := s.db.QueryRow("SELECT `raw` FROM instances WHERE `type` = ? AND oid = ?",
		model.HProfHDRecordTypeClassDump, oid)
	var raw []byte
	err := row.Scan(&raw)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var result = &model.HProfClassDump{}
	err = decodeGob(raw, result)
	return result, err
}

func (s *SqliteStorage) ListClasses(fn func(dump *model.HProfClassDump) error) error {
	rows, err := s.db.Query("SELECT `raw` FROM instances WHERE `type` = ?",
		model.HProfHDRecordTypeClassDump)
	if err != nil {
		return err
	}
	defer rows.Close()
	var raw []byte
	var result = &model.HProfClassDump{}
	for rows.Next() {
		err = rows.Scan(&raw)
		if err != nil {
			return err
		}
		err = decodeGob(raw, result)
		if err != nil {
			return err
		}
		err = fn(result)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SqliteStorage) ListInstances(fn func(dump model.HProfObjectRecord) error) error {
	rows, err := s.db.Query("SELECT `type`, `raw` FROM instances WHERE `type` IN (?, ?, ?)",
		model.HProfHDRecordTypeInstanceDump, model.HProfHDRecordTypeObjectArrayDump,
		model.HProfHDRecordTypePrimitiveArrayDump)
	if err != nil {
		return err
	}
	defer rows.Close()
	var typ byte
	var raw []byte
	var instance = &model.HProfInstanceDump{}
	var objectArray = &model.HProfObjectArrayDump{}
	var primitiveArray = &model.HProfObjectArrayDump{}
	for rows.Next() {
		err = rows.Scan(&typ, &raw)
		if err != nil {
			return err
		}
		if typ == model.HProfHDRecordTypeInstanceDump {
			err = decodeGob(raw, instance)
			if err != nil {
				return err
			}
			err = fn(instance)
		} else if typ == model.HProfHDRecordTypeObjectArrayDump {
			err = decodeGob(raw, objectArray)
			if err != nil {
				return err
			}
			err = fn(objectArray)
		} else if typ == model.HProfHDRecordTypePrimitiveArrayDump {
			err = decodeGob(raw, primitiveArray)
			if err != nil {
				return err
			}
			err = fn(primitiveArray)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SqliteStorage) ListHeapObject(fn func(dump *model.HeapObject) error) error {
	rows, err := s.db.Query("SELECT `id`, `type`, `raw`, cid FROM instances WHERE `type` IN (?, ?, ?) ORDER BY id",
		model.HProfHDRecordTypeInstanceDump, model.HProfHDRecordTypeObjectArrayDump,
		model.HProfHDRecordTypePrimitiveArrayDump)
	if err != nil {
		return err
	}
	defer rows.Close()
	var typ byte
	var raw []byte
	var classId uint64
	var instance = &model.HProfInstanceDump{}
	var objectArray = &model.HProfObjectArrayDump{}
	var primitiveArray = &model.HProfObjectArrayDump{}
	var id int
	for rows.Next() {
		err = rows.Scan(&id, &typ, &raw, &classId)
		if err != nil {
			return err
		}
		class, err := s.FindClass(classId)
		if err != nil {
			return err
		}
		object := &model.HeapObject{
			Class:      class,
			References: []uint64{},
		}
		if typ == model.HProfHDRecordTypeInstanceDump {
			err = decodeGob(raw, instance)
			if err != nil {
				return err
			}
			object.Instance = instance
			err = fn(object)
		} else if typ == model.HProfHDRecordTypeObjectArrayDump {
			err = decodeGob(raw, objectArray)
			if err != nil {
				return err
			}
			object.Instance = objectArray
			err = fn(object)
		} else if typ == model.HProfHDRecordTypePrimitiveArrayDump {
			err = decodeGob(raw, primitiveArray)
			if err != nil {
				return err
			}
			object.Instance = primitiveArray
			err = fn(object)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SqliteStorage) CountClasses() (int, error) {
	return s.CountByType(byte(model.HProfHDRecordTypeClassDump))
}

func (s *SqliteStorage) CountInstances() (int, error) {
	row := s.db.QueryRow("SELECT count(id) FROM instances WHERE `type` IN (?, ?, ?)",
		model.HProfHDRecordTypeInstanceDump, model.HProfHDRecordTypeObjectArrayDump,
		model.HProfHDRecordTypePrimitiveArrayDump)
	var count = -1
	err := row.Scan(&count)
	return count, err
}

func (s *SqliteStorage) CountByType(objType byte) (int, error) {
	row := s.db.QueryRow("SELECT count(id) FROM instances WHERE `type` = ?",
		objType)
	var count = -1
	err := row.Scan(&count)
	return count, err
}

func (s *SqliteStorage) GetPosByType(id uint64, objType byte) (int64, error) {
	rows, err := s.db.Query("SELECT pos FROM instances WHERE `type` = ? AND oid = ?",
		objType, id)
	if err != nil {
		return -1, err
	}
	var pos int64 = -1
	if rows.Next() {
		err = rows.Scan(&pos)
	}
	return pos, rows.Close()
}

func (s *SqliteStorage) GetInboundById(id int) []int {
	return nil
}

func (s *SqliteStorage) SetInboundById(id int, ids []int) {

}

func (s *SqliteStorage) GetOutboundById(id int) []int {
	return nil
}

func (s *SqliteStorage) SetOutboundById(id uint64, ids []uint64) error {
	bs := encodeGob(ids)
	_, err := s.db.Exec("UPDATE instances SET outbound = ? WHERE oid = ?",
		bs, id)
	return err
}

// encodeGob 主要用于序列化数组
func encodeGob(v interface{}) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// encodeGob 主要用于反序列化数组
func decodeGob(b []byte, result interface{}) error {
	buf := bytes.NewBuffer(b)
	enc := gob.NewDecoder(buf)

	return enc.Decode(result)
}
