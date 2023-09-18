package db

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"hprof-tool/pkg/hprof"
	"strings"
)

var schema = strings.ReplaceAll(`
CREATE TABLE IF NOT EXISTS kvs (
    'key' TEXT PRIMARY KEY,
    'value' BLOB NOT NULL
);
CREATE INDEX kvs_idx ON kvs ('key');
CREATE TABLE IF NOT EXISTS texts (
    id INTEGER PRIMARY KEY,
    -- 文件位置，如果是 -1 则表示新增文本
    pos INTEGER NOT NULL,
    -- 新增文本内容
    txt TEXT 
);
CREATE TABLE IF NOT EXISTS load_classes (
    id INTEGER PRIMARY KEY,
    cid INTEGER NOT NULL,
    nameId INTEGER NOT NULL
);
CREATE INDEX load_classes_idx ON load_classes ('cid');
CREATE TABLE IF NOT EXISTS classes (
    id INTEGER PRIMARY KEY,
    -- 文件位置, 如果是 -1 则是 fake class
    pos INTEGER NOT NULL,
	-- fake class data
	'raw' BLOB,
	instance_size INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS instances (
    id INTEGER PRIMARY KEY,
    -- 文件位置
    pos INTEGER NOT NULL,
    -- ClassId
    cid INTEGER NOT NULL, 
    -- 对象大小
    'size' INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS object_arrays (
	-- Object Id
	id INTEGER PRIMARY KEY,
    -- 文件位置
    pos INTEGER NOT NULL,
    -- ClassId
    cid INTEGER NOT NULL,
    -- 对象大小
    'size' INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS primitive_arrays (
	-- Object Id
    id INTEGER PRIMARY KEY,
    -- 文件位置
    pos INTEGER NOT NULL,
    -- type
    'type' INTEGER NOT NULL,
    -- 对象大小
    'size' INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS gcroots (
    id INTEGER PRIMARY KEY,
    -- 文件位置
    pos INTEGER NOT NULL,
    -- 类型
    'type' INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS threads (
    id INTEGER PRIMARY KEY,
	'raw' BLOB NOT NULL
);
CREATE TABLE IF NOT EXISTS thread_traces (
    id INTEGER PRIMARY KEY,
	'raw' BLOB NOT NULL
);
CREATE TABLE IF NOT EXISTS thread_frames (
    id INTEGER PRIMARY KEY,
	'raw' BLOB NOT NULL
);
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
	fmt.Printf("Init sqlite\n")
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
	fmt.Printf("Init sqlite finished. %d %d\n", insertId, rowsAffected) // TODO

	return nil
}

func (s *SqliteStorage) Close() error {
	return s.db.Close()
}

func (s *SqliteStorage) PutKV(key string, value interface{}) error {
	body := encodeGob(value)
	_, err := s.db.Exec("INSERT OR REPLACE INTO kvs VALUES (?, ?)", key, body)
	return err
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

// GetText 获取文本
func (s *SqliteStorage) GetText(id uint64) (int64, string, error) {
	row := s.db.QueryRow("SELECT pos, txt FROM texts WHERE id=?", id)
	var err error
	var pos int64
	var raw []byte
	if err = row.Scan(&pos, &raw); err == sql.ErrNoRows {
		// TODO NotFound error
		return 0, "", err
	}
	if raw != nil {
		return -1, string(raw), nil
	}
	return pos, "", err
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

// GetLoadClassByClassId load class by class id
func (s *SqliteStorage) GetLoadClassById(id uint64) (uint64, uint64, error) {
	row := s.db.QueryRow("SELECT cid, nameId FROM load_classes WHERE id=?", id)
	var err error
	var cid uint64
	var nameId uint64
	if err = row.Scan(&cid, &nameId); err == sql.ErrNoRows {
		// TODO NotFound error
		return 0, 0, err
	}
	return cid, nameId, err
}

// GetLoadClassByClassId load class by class id
func (s *SqliteStorage) GetLoadClassByClassId(cid uint64) (uint64, uint64, error) {
	row := s.db.QueryRow("SELECT id, nameId FROM load_classes WHERE cid=?", cid)
	var err error
	var id uint64
	var nameId uint64
	if err = row.Scan(&id, &nameId); err == sql.ErrNoRows {
		// TODO NotFound error
		return 0, 0, err
	}
	return id, nameId, err
}

// SaveClass 记录 Classes 索引
func (s *SqliteStorage) SaveClass(pos, cid int64, instanceSize int) error {
	_, err := s.db.Exec("INSERT INTO classes (id, `pos`, instance_size) VALUES (?, ?, ?)", cid, pos, instanceSize)
	return err
}

// AddClass 记录 Classes 索引
func (s *SqliteStorage) AddClass(fakeClass *hprof.HProfClassRecord) (uint64, error) {
	value := encodeGob(fakeClass)
	_, err := s.db.Exec("INSERT INTO classes (`pos`, `raw`) VALUES (?, ?)", -1, value)
	if err != nil {
		return 0, err
	}
	row := s.db.QueryRow("SELECT last_insert_rowid();")
	var lastId uint64 = 0
	err = row.Scan(&lastId)
	return lastId, err
}

// GetClass 记录 Classes 索引
func (s *SqliteStorage) GetClass(cid uint64) (int64, *hprof.HProfClassRecord, error) {
	row := s.db.QueryRow("SELECT `pos`, `raw` FROM classes WHERE id=?", cid)
	var err error
	var pos int64
	var raw []byte
	if err = row.Scan(&pos, &raw); err == sql.ErrNoRows {
		// TODO NotFound error
		return -1, nil, err
	}
	if raw != nil {
		cla := &hprof.HProfClassRecord{}
		err = decodeGob(raw, cla)
		return pos, cla, err
	}
	return pos, nil, nil
}

func (s *SqliteStorage) ListClasses(fn func(id uint64, pos int64, cla *hprof.HProfClassRecord) error) error {
	rows, err := s.db.Query("SELECT id, `pos`, `raw` FROM classes ORDER BY id")
	if err != nil {
		return err
	}
	defer rows.Close()
	var id uint64
	var pos int64
	var raw []byte
	cla := &hprof.HProfClassRecord{}
	for rows.Next() {
		err = rows.Scan(&id, &pos, &raw)
		if err != nil {
			return err
		}
		if raw != nil {
			err = decodeGob(raw, cla)
			if err != nil {
				return err
			}
		}
		err = fn(id, pos, cla)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveInstance 记录 Instances 索引
func (s *SqliteStorage) SaveInstance(pos, oid, cid int64, size int) error {
	_, err := s.db.Exec("INSERT INTO instances (id, `pos`, cid, `size`) VALUES (?, ?, ?, ?)",
		oid, pos, cid, size)
	return err
}

func (s *SqliteStorage) ListInstances(fn func(id uint64, pos int64, cid uint64) error) error {
	rows, err := s.db.Query("SELECT id, `pos`, cid FROM instances ORDER BY id")
	if err != nil {
		return err
	}
	defer rows.Close()
	var id uint64
	var pos int64
	var cid uint64
	for rows.Next() {
		err = rows.Scan(&id, &pos, &cid)
		if err != nil {
			return err
		}
		err = fn(id, pos, cid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SqliteStorage) CountInstancesByClass(fn func(cid uint64, count, size int64) error) error {
	rows, err := s.db.Query("SELECT cid, COUNT(id) as c, SUM(`size`) as s FROM instances GROUP BY cid")
	if err != nil {
		return err
	}
	defer rows.Close()
	var cid uint64
	var count int64
	var size int64
	for rows.Next() {
		err = rows.Scan(&cid, &count, &size)
		if err != nil {
			return err
		}
		err = fn(cid, count, size)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveObjectArray 记录 ObjectArray 索引
func (s *SqliteStorage) SaveObjectArray(pos, oid, cid int64, size int) error {
	_, err := s.db.Exec("INSERT INTO object_arrays (id, `pos`, cid, `size`) VALUES (?, ?, ?, ?)", oid, pos, cid, size)
	return err
}

func (s *SqliteStorage) CountObjectArrayByClass(fn func(cid uint64, count, size int64) error) error {
	rows, err := s.db.Query("SELECT cid, COUNT(id) as c, SUM(`size`) as s FROM object_arrays GROUP BY cid")
	if err != nil {
		return err
	}
	defer rows.Close()
	var cid uint64
	var count int64
	var size int64
	for rows.Next() {
		err = rows.Scan(&cid, &count, &size)
		if err != nil {
			return err
		}
		err = fn(cid, count, size)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveInstance 记录 Instances 索引
func (s *SqliteStorage) SavePrimitiveArray(pos, oid, typ int64, size int) error {
	_, err := s.db.Exec("INSERT INTO primitive_arrays (id, `pos`, `type`, `size`) VALUES (?, ?, ?, ?)", oid, pos, typ, size)
	return err
}

func (s *SqliteStorage) CountPrimitiveArrayByType(fn func(cid uint64, count, size int64) error) error {
	rows, err := s.db.Query("SELECT `type`, COUNT(id) as c, SUM(`size`) as s FROM primitive_arrays GROUP BY `type`")
	if err != nil {
		return err
	}
	defer rows.Close()
	var typ uint64
	var count int64
	var size int64
	for rows.Next() {
		err = rows.Scan(&typ, &count, &size)
		if err != nil {
			return err
		}
		err = fn(typ, count, size)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveGCRoot 记录 gc roots 索引
func (s *SqliteStorage) SaveGCRoot(typ int, pos int64) error {
	_, err := s.db.Exec("INSERT INTO gcroots (`pos`, `type`) VALUES (?, ?)",
		pos, typ)
	return err
}

func (s *SqliteStorage) ListGCRoots(fn func(pos int64, typ int) error) error {
	rows, err := s.db.Query("SELECT pos, `type` FROM gcroots ORDER BY id")
	if err != nil {
		return err
	}
	defer rows.Close()
	var pos int64
	var typ int
	for rows.Next() {
		err = rows.Scan(&pos, &typ)
		if err != nil {
			return err
		}
		err = fn(pos, typ)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SqliteStorage) SaveThread(r *hprof.HProfThreadRecord) error {
	body := encodeGob(r)
	_, err := s.db.Exec("INSERT INTO threads (`raw`) VALUES (?)", body)
	return err
}

func (s *SqliteStorage) ListThreads(fn func(r *hprof.HProfThreadRecord) error) error {
	rows, err := s.db.Query("SELECT `raw` FROM threads ORDER BY id")
	if err != nil {
		return err
	}
	defer rows.Close()
	var raw []byte
	var record = &hprof.HProfThreadRecord{}
	for rows.Next() {
		err = rows.Scan(&raw)
		if err != nil {
			return err
		}
		err = decodeGob(raw, record)
		if err != nil {
			return err
		}
		err = fn(record)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveThreadTrace 记录 thread trace 索引
func (s *SqliteStorage) SaveThreadTrace(r *hprof.HProfTraceRecord) error {
	body := encodeGob(r)
	_, err := s.db.Exec("INSERT INTO thread_traces (`raw`) VALUES (?)", body)
	return err
}

func (s *SqliteStorage) ListThreadTraces(fn func(r *hprof.HProfTraceRecord) error) error {
	rows, err := s.db.Query("SELECT `raw` FROM thread_traces ORDER BY id")
	if err != nil {
		return err
	}
	defer rows.Close()
	var raw []byte
	var record = &hprof.HProfTraceRecord{}
	for rows.Next() {
		err = rows.Scan(&raw)
		if err != nil {
			return err
		}
		err = decodeGob(raw, record)
		if err != nil {
			return err
		}
		err = fn(record)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveThreadTrace 记录 thread trace 索引
func (s *SqliteStorage) SaveThreadFrame(r *hprof.HProfFrameRecord) error {
	body := encodeGob(r)
	_, err := s.db.Exec("INSERT INTO thread_frames (`raw`) VALUES (?)", body)
	return err
}

func (s *SqliteStorage) ListThreadFrames(fn func(r *hprof.HProfFrameRecord) error) error {
	rows, err := s.db.Query("SELECT `raw` FROM thread_frames ORDER BY id")
	if err != nil {
		return err
	}
	defer rows.Close()
	var raw []byte
	var record = &hprof.HProfFrameRecord{}
	for rows.Next() {
		err = rows.Scan(&raw)
		if err != nil {
			return err
		}
		err = decodeGob(raw, record)
		if err != nil {
			return err
		}
		err = fn(record)
		if err != nil {
			return err
		}
	}
	return nil
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
