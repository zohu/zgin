package zid

import (
	"fmt"
	"strconv"
)

type SnowWorkerM2 struct {
	*SnowWorkerM1
}

func NewSnowWorkerM2(options *Options) ISnowWorker {
	return &SnowWorkerM2{
		NewSnowWorkerM1(options).(*SnowWorkerM1),
	}
}

func (m2 SnowWorkerM2) NextId() int64 {
	m2.Lock()
	defer m2.Unlock()
	currentTimeTick := m2.GetCurrentTimeTick()
	if m2.lastTimeTick == currentTimeTick {
		m2.currentSeqNumber++
		if m2.currentSeqNumber > m2.maxSeqNumber {
			m2.currentSeqNumber = m2.minSeqNumber
			currentTimeTick = m2.GetNextTimeTick()
		}
	} else {
		m2.currentSeqNumber = m2.minSeqNumber
	}
	if currentTimeTick < m2.lastTimeTick {
		fmt.Println("Time error for {0} milliseconds", strconv.FormatInt(m2.lastTimeTick-currentTimeTick, 10))
	}
	m2.lastTimeTick = currentTimeTick
	result := currentTimeTick<<m2.timestampShift + int64(m2.workerId<<m2.seqBitLength) + int64(m2.currentSeqNumber)
	return result
}
