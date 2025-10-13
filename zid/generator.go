package zid

import (
	"time"

	"github.com/zohu/zgin/zutil"
)

type DefaultIdGenerator struct {
	Options    *Options
	SnowWorker ISnowWorker
}

func NewDefaultIdGenerator(options *Options) *DefaultIdGenerator {
	options = zutil.FirstTruth(options, new(Options))
	options.Validate()
	return &DefaultIdGenerator{
		Options:    options,
		SnowWorker: NewSnowWorkerM1(options),
	}
}

func (dig DefaultIdGenerator) NextId() int64 {
	return dig.SnowWorker.NextId()
}
func (dig DefaultIdGenerator) ExtractTime(id int64) time.Time {
	return dig.SnowWorker.ExtractTime(id)
}
func (dig DefaultIdGenerator) ExtractWorkerId(id int64) int64 {
	return dig.SnowWorker.ExtractWorkerId(id)
}
