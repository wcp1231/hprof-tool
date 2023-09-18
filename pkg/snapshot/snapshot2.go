package snapshot

import (
	"fmt"
	"hprof-tool/pkg/db"
	"hprof-tool/pkg/hprof"
	"hprof-tool/pkg/indexer"
	"hprof-tool/pkg/model"
	"os"
)

type Snapshot2 struct {
	i *indexer.Indexer
}

func NewSnapshot2(fileName string) (*Snapshot2, error) {
	hFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	hreader := hprof.NewReader(hFile)

	s, err := db.NewSqliteStorage(":memory:")
	if err != nil {
		return nil, err
	}

	fmt.Println("Init tables")
	err = s.Init()
	if err != nil {
		return nil, err
	}

	i := indexer.NewSqliteIndexer(hreader, s)

	return &Snapshot2{i}, nil
}

func (s *Snapshot2) EnsureCreateIndex() error {
	// TODO 判断 sqlite 是否有数据
	err := s.i.CreateIndex()
	if err != nil {
		return err
	}
	println("CreateIndex done")

	return s.i.Processor()
}

// GetThreads 返回线程信息
func (s *Snapshot2) GetThreads() map[uint32]*model.Thread {
	return s.i.GetThreads()
}

func (s *Snapshot2) GetText(tId uint64) (string, error) {
	return s.i.GetText(tId)
}

func (s *Snapshot2) GetClassNameByClassSerialNumber(csn uint64) (string, error) {
	return s.i.GetClassNameByClassSerialNumber(csn)
}

func (s *Snapshot2) ListClassesStatistics() ([]ClassStatistics, error) {
	var result []ClassStatistics
	err := s.i.GetClassesStatistics(func(cid uint64, cname string, count, size int64) error {
		result = append(result, ClassStatistics{
			Id:            cid,
			Name:          cname,
			InstanceCount: count,
			InstanceSize:  size,
		})
		return nil
	})
	return result, err
}
