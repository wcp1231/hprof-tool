package indexer

type IndexerProcessor interface {
	process() error
}

func (i *Indexer) Processor() error {
	var processors []IndexerProcessor
	processors = append(processors, newCreateClassIndexesProcessor(i))
	processors = append(processors, newCreateFakeClassesProcessor(i))
	processors = append(processors, newThreadTracesProcessor(i))
	processors = append(processors, newGCRootProcessor(i))
	processors = append(processors, newClassReferencesProcessor(i))
	processors = append(processors, newInstanceReferencesProcessor(i))

	for _, processor := range processors {
		err := processor.process()
		if err != nil {
			return err
		}
	}
	return nil
}
