package zid

import (
	"github.com/zohu/zgin/zutil"
	"time"
)

type DefaultIdGenerator struct {
	Options    *Options
	SnowWorker ISnowWorker
}

func NewDefaultIdGenerator(options *Options) *DefaultIdGenerator {
	options = zutil.FirstTruth(options, new(Options))
	options.Validate()

	var snowWorker ISnowWorker
	switch options.Method {
	case 1:
		snowWorker = NewSnowWorkerM1(options)
	case 2:
		snowWorker = NewSnowWorkerM2(options)
	default:
		snowWorker = NewSnowWorkerM1(options)
	}
	if options.Method == 1 {
		time.Sleep(time.Duration(500) * time.Microsecond)
	}
	return &DefaultIdGenerator{
		Options:    options,
		SnowWorker: snowWorker,
	}
}

func (dig DefaultIdGenerator) NextId() int64 {
	return dig.SnowWorker.NextId()
}
func (dig DefaultIdGenerator) NextIdStr() string {
	return dig.SnowWorker.NextIdStr()
}
func (dig DefaultIdGenerator) ExtractTime(id int64) time.Time {
	return dig.SnowWorker.ExtractTime(id)
}
