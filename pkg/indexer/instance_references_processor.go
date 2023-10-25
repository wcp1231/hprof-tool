package indexer

import "hprof-tool/pkg/hprof"

// InstanceReferencesProcessor 计算 instance 的 references
type InstanceReferencesProcessor struct {
	i *Indexer
}

func newInstanceReferencesProcessor(i *Indexer) *InstanceReferencesProcessor {
	return &InstanceReferencesProcessor{i}
}

func (p *InstanceReferencesProcessor) process() error {
	println("InstanceReferencesProcessor start")
	return p.i.ForEachInstanceRecords(func(record *hprof.HProfInstanceRecord) error {
		references, err := p.getReferences(record)
		if err != nil {
			return err
		}
		p.i.ctx.instanceReferences[record.ObjectId] = references
		return p.saveReferences(record.ObjectId, references)
	})
}

func (p *InstanceReferencesProcessor) saveReferences(rid uint64, references []uint64) error {
	for _, ref := range references {
		err := p.i.AppendReference(rid, ref, 0) // TODO type
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *InstanceReferencesProcessor) getReferences(instance *hprof.HProfInstanceRecord) ([]uint64, error) {
	references := []uint64{}
	references = append(references, instance.ClassObjectId)

	class, err := p.i.getClassById(instance.ClassObjectId)
	if err != nil {
		return nil, err
	}
	classes, err := p.resolveClassHierarchy(class)
	if err != nil {
		return nil, err
	}
	instanceFields := []*hprof.HProfClass_InstanceField{}
	for _, class := range classes {
		instanceFields = append(instanceFields, class.InstanceFields...)
	}

	fiedValues, err := instance.ReadValues(instanceFields)
	if err != nil {
		return nil, err
	}

	for _, fiedValue := range fiedValues {
		if fiedValue.ValueType() == hprof.HProfValueType_OBJECT {
			objectValue := fiedValue.(*hprof.HProfInstanceObjectValue)
			references = append(references, objectValue.Value)
		}
	}

	return references, nil
}

func (p *InstanceReferencesProcessor) resolveClassHierarchy(class *hprof.HProfClassRecord) ([]*hprof.HProfClassRecord, error) {
	var classes []*hprof.HProfClassRecord
	classes = append(classes, class)
	c := class
	var err error
	for {
		if c.SuperClassObjectId == 0 {
			break
		}
		c, err = p.i.getClassById(c.SuperClassObjectId)
		if err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}

	return classes, nil
}
