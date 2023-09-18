package indexer

import (
	"fmt"
	"hprof-tool/pkg/hprof"
	"hprof-tool/pkg/model"
)

type ThreadTracesProcessor struct {
	i *Indexer
}

func newThreadTracesProcessor(i *Indexer) *ThreadTracesProcessor {
	return &ThreadTracesProcessor{i}
}

func (p *ThreadTracesProcessor) process() error {
	println("ThreadTracesProcessor start")
	err := p.i.ForEachThreadFrames(func(r *hprof.HProfFrameRecord) error {
		p.i.ctx.id2frame[r.StackFrameId] = &model.StackFrame{
			FrameId:           r.StackFrameId,
			MethodId:          r.MethodNameId,
			SignatureId:       r.MethodSignatureId,
			SourceFileId:      r.SourceFileNameId,
			ClassSerialNumber: r.ClassSerialNumber,
			Line:              r.LineNumber,
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = p.i.ForEachThreadTraces(func(r *hprof.HProfTraceRecord) error {
		p.i.ctx.serNum2stackTrace[r.StackTraceSerialNumber] = &stackTrace{
			ThreadSerialNumber: r.ThreadSerialNumber,
			FrameIds:           r.StackFrameIds,
		}
		return nil
	})
	if err != nil {
		return err
	}
	return p.i.ForEachThreads(func(r *hprof.HProfThreadRecord) error {
		fmt.Printf("Thread ThreadSerialNumber = %d\n", r.ThreadSerialNumber)
		p.i.ctx.threadSN2thread[r.ThreadSerialNumber] = &thread{
			ObjectId:          r.ThreadObjectId,
			NameId:            r.ThreadNameId,
			GroupNameId:       r.ThreadGroupNameId,
			GroupParentNameId: r.ThreadGroupParentNameId,
		}
		return nil
	})
}
