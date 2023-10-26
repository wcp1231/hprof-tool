package snapshot

import (
	"fmt"
	"hprof-tool/pkg/hprof"
	"hprof-tool/pkg/indexer"
	"hprof-tool/pkg/model"
	"hprof-tool/pkg/storage"
	"os"
	"sort"
)

type Snapshot struct {
	i *indexer.Indexer
}

func NewSnapshot(fileName string) (*Snapshot, error) {
	hFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	hreader := hprof.NewReader(hFile)

	s, err := storage.NewSqliteStorage(":memory:")
	if err != nil {
		return nil, err
	}

	fmt.Println("Init tables")
	err = s.Init()
	if err != nil {
		return nil, err
	}

	i := indexer.NewSqliteIndexer(hreader, s)

	return &Snapshot{i}, nil
}

func (s *Snapshot) EnsureCreateIndex() error {
	// TODO 判断 sqlite 是否有数据
	err := s.i.CreateIndex()
	if err != nil {
		return err
	}
	println("CreateIndex done")

	return s.i.Processor()
}

// GetThreads 返回线程信息
func (s *Snapshot) GetThreads() map[uint32]*model.Thread {
	return s.i.GetThreads()
}

func (s *Snapshot) GetText(tId uint64) (string, error) {
	return s.i.GetText(tId)
}

func (s *Snapshot) GetClassNameByClassSerialNumber(csn uint64) (string, error) {
	return s.i.GetClassNameByClassSerialNumber(csn)
}

func (s *Snapshot) ListClassesStatistics() ([]ClassStatistics, error) {
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
	sort.Slice(result, func(i, j int) bool {
		if result[i].InstanceCount == result[j].InstanceCount {
			return result[i].InstanceSize > result[j].InstanceSize
		}
		return result[i].InstanceCount > result[j].InstanceCount
	})
	return result, err
}

func (s *Snapshot) ListInstancesStatistics(cid uint64, typ int) ([]InstanceStatistics, error) {
	var result []InstanceStatistics
	err := s.i.GetInstancesStatistics(cid, typ, func(cid uint64, size int64) error {
		result = append(result, InstanceStatistics{
			Id:   cid,
			Size: size,
		})
		return nil
	})
	sort.Slice(result, func(i, j int) bool {
		return result[i].Size > result[j].Size
	})
	return result, err
}

func (s *Snapshot) GetInstanceDetail(id uint64) (*indexer.Instance, error) {
	return s.i.GetInstanceDetail(id)
}

func (s *Snapshot) GetRecordInbound(id uint64, fn func(record hprof.HProfRecord) error) error {
	return s.i.GetRecordInbounds(id, fn)
}
