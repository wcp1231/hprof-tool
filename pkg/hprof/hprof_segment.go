package hprof

import (
	"math"
)

type HProfRecordHeapDumpBoundary struct {
	HProfBasicRecord
}

func (m *HProfRecordHeapDumpBoundary) Id() uint64 {
	return 0
}

func (m *HProfRecordHeapDumpBoundary) Type() HProfRecordType {
	return HProfRecordTypeHeapDump
}

func parseHeapDumpSegment(pr *HProfReader) (*HProfRecordHeapDumpBoundary, error) {
	sz, err := pr.parseRecordSize()
	if err != nil {
		return nil, err
	}

	println("HeapDumpSegmentStart", sz)

	if sz == 0 {
		// Truncated. Set to the max int.
		sz = math.MaxUint32
	}
	pr.heapDumpFrameLeftBytes = sz
	return &HProfRecordHeapDumpBoundary{}, nil
}

func parseHeapDumpSegmentEnd(pr *HProfReader) (*HProfRecordHeapDumpBoundary, error) {
	_, err := pr.parseRecordSize()
	if err != nil {
		return nil, err
	}
	println("HeapDumpSegmentEnd")
	return &HProfRecordHeapDumpBoundary{}, nil
}

func parseHeapDump(pr *HProfReader) (*HProfRecordHeapDumpBoundary, error) {
	sz, err := pr.parseRecordSize()
	if err != nil {
		return nil, err
	}
	pr.heapDumpFrameLeftBytes = sz
	return &HProfRecordHeapDumpBoundary{}, nil
}
