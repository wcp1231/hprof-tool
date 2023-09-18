package indexer

import "hprof-tool/pkg/hprof"

// ClassReferencesProcessor 计算 class 的 references
type ClassReferencesProcessor struct {
	i *Indexer
}

func newClassReferencesProcessor(i *Indexer) *ClassReferencesProcessor {
	return &ClassReferencesProcessor{i}
}

func (p *ClassReferencesProcessor) process() error {
	println("ClassReferencesProcessor start")
	return p.i.ForEachClassRecords(func(record *hprof.HProfClassRecord) error {
		references := p.getReferences(record)
		p.i.ctx.classReferences[record.ClassObjectId] = references
		return nil
	})
}

func (p *ClassReferencesProcessor) getReferences(cr *hprof.HProfClassRecord) []uint64 {
	references := []uint64{}
	// 所有类都是 java.lang.Class 的实例
	jlc := p.i.getClassIdByName("java.lang.Class", 0)
	references = append(references, jlc)
	if cr.SuperClassObjectId != 0 {
		references = append(references, cr.SuperClassObjectId)
	}
	references = append(references, cr.ClassLoaderObjectId)
	for _, sf := range cr.StaticFields {
		if sf.Type == hprof.HProfValueType_OBJECT {
			references = append(references, sf.Value)
		}
	}
	return references
}
