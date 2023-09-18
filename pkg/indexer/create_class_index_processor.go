package indexer

import (
	"strings"
)

// CreateClassIndexesProcessor 建立 class 的索引
// 比如 ClassName 和 ClassId 的索引
type CreateClassIndexesProcessor struct {
	i *Indexer
}

func newCreateClassIndexesProcessor(i *Indexer) *CreateClassIndexesProcessor {
	return &CreateClassIndexesProcessor{i}
}

func (p *CreateClassIndexesProcessor) process() error {
	println("CreateClassIndexesProcessor start")
	return p.i.ForEachClassesWithName(func(cid uint64, cname string) error {
		cname = strings.ReplaceAll(cname, "/", ".")
		p.i.ctx.className2Cid[cname] = []uint64{cid}
		p.i.ctx.classId2Name[cid] = cname
		return nil
	})
}
