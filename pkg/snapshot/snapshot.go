package snapshot

import (
	"fmt"
	"hprof-tool/pkg/model"
	"hprof-tool/pkg/parser"
	"hprof-tool/pkg/storage"
	"os"
)

type Snapshot struct {
	hFile *os.File

	parser  *parser.HProfParser
	storage storage.Storage

	header *parser.HProfHeader
}

func NewSnapshot(fileName string) (*Snapshot, error) {
	hFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	s, err := storage.NewSqliteStorage(":memory:")
	if err != nil {
		return nil, err
	}

	fmt.Println("Init tables")
	err = s.Init()
	if err != nil {
		return nil, err
	}

	return &Snapshot{
		hFile: hFile,

		parser:  parser.NewParser(hFile),
		storage: s,
	}, nil
}

// Read 分阶段读取 HProf 文件
func (s *Snapshot) Read() error {
	err := s.firstProcess()
	if err != nil {
		return err
	}

	err = s.secondProcess()
	if err != nil {
		return err
	}

	return nil
}

func (s *Snapshot) firstProcess() error {
	firstProcess := NewFirstProcessor(s, s.parser)
	err := firstProcess.Process()
	if err != nil {
		return err
	}
	return nil
}

func (s *Snapshot) secondProcess() error {

	return nil
}

func (s *Snapshot) ForEachClass(fn func(dump *model.HProfClassDump) error) error {
	return s.storage.ListClasses(func(classDump *model.HProfClassDump) error {
		err := s.parser.Seek(classDump.POS)
		if err != nil {
			return err
		}
		result, err := s.parser.ParseClassDumpRecord()
		if err != nil {
			return err
		}
		return fn(result)
	})
}

func (s *Snapshot) ClassCount() (int, error) {
	st := s.storage.(*storage.SqliteStorage)
	return st.CountClasses()
}

func (s *Snapshot) InstanceCount() (int, error) {
	st := s.storage.(*storage.SqliteStorage)
	return st.CountInstances()
}
