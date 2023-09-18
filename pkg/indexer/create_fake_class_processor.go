package indexer

import (
	"hprof-tool/pkg/hprof"
)

// CreateFakeClassesProcessor 创建缺失的类
// TODO 但我看这些假类没啥引用，也许可以不用关心？
type CreateFakeClassesProcessor struct {
	i *Indexer
}

func newCreateFakeClassesProcessor(i *Indexer) *CreateFakeClassesProcessor {
	return &CreateFakeClassesProcessor{i}
}

func (p *CreateFakeClassesProcessor) process() error {
	err := p.createSystemClass()
	if err != nil {
		return err
	}
	// TODO 查找不存在的 class 并创建
	return nil
}

func (p *CreateFakeClassesProcessor) createSystemClass() error {
	println("CreateFakeClassesProcessor start")
	jlo := p.i.getClassIdByName("java.lang.Object", 0)

	cid := p.i.getClassIdByName("java.lang.Class", 0)
	if cid == 0 {
		err := p.createFakeSystemClass("java.lang.Class", jlo)
		if err != nil {
			return err
		}
	}

	cid = p.i.getClassIdByName("java.lang.ClassLoader", 0)
	if cid == 0 {
		err := p.createFakeSystemClass("java.lang.ClassLoader", jlo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *CreateFakeClassesProcessor) createFakeSystemClass(name string, scid uint64) error {
	fake := &hprof.HProfClassRecord{
		SuperClassObjectId:  scid,
		ClassLoaderObjectId: 0,
	}
	nameId, err := p.i.storage.AddText(name)
	if err != nil {
		return err
	}
	cid, err := p.i.storage.AddClass(fake)
	if err != nil {
		return err
	}
	err = p.i.storage.AddLoadClass(cid, nameId)
	if err != nil {
		return err
	}
	p.i.ctx.className2Cid[name] = []uint64{cid}
	p.i.ctx.classId2Name[cid] = name
	return nil
}
