package zid

import (
	"strconv"
	"sync"
	"time"
)

type SnowWorkerM1 struct {
	baseTime          int64  // 基础时间
	workerId          uint16 // 机器码
	workerIdBitLength byte   // 机器码位长
	seqBitLength      byte   // 自增序列数位长
	maxSeqNumber      uint32 // 最大序列数（含）
	minSeqNumber      uint32 // 最小序列数（含）
	topOverCostCount  uint32 // 最大漂移次数

	timestampShift         byte
	currentSeqNumber       uint32
	lastTimeTick           int64
	turnBackTimeTick       int64
	turnBackIndex          byte
	isOverCost             bool
	overCostCountInOneTerm uint32

	sync.Mutex
}

func NewSnowWorkerM1(options *Options) ISnowWorker {
	return &SnowWorkerM1{
		baseTime:          options.BaseTime,
		workerIdBitLength: options.WorkerIdBitLength,
		workerId:          options.WorkerId,
		seqBitLength:      options.SeqBitLength,
		maxSeqNumber:      options.maxSeqNumber(),
		minSeqNumber:      options.MinSeqNumber,
		topOverCostCount:  options.TopOverCostCount,
		timestampShift:    options.WorkerIdBitLength + options.SeqBitLength,
		currentSeqNumber:  options.MinSeqNumber,

		lastTimeTick:           0,
		turnBackTimeTick:       0,
		turnBackIndex:          0,
		isOverCost:             false,
		overCostCountInOneTerm: 0,
	}
}

func (m1 *SnowWorkerM1) NextOverCostId() int64 {
	currentTimeTick := m1.GetCurrentTimeTick()
	if currentTimeTick > m1.lastTimeTick {
		m1.lastTimeTick = currentTimeTick
		m1.currentSeqNumber = m1.minSeqNumber
		m1.isOverCost = false
		m1.overCostCountInOneTerm = 0
		return m1.CalcId(m1.lastTimeTick)
	}
	if m1.overCostCountInOneTerm >= m1.topOverCostCount {
		m1.lastTimeTick = m1.GetNextTimeTick()
		m1.currentSeqNumber = m1.minSeqNumber
		m1.isOverCost = false
		m1.overCostCountInOneTerm = 0
		return m1.CalcId(m1.lastTimeTick)
	}
	if m1.currentSeqNumber > m1.maxSeqNumber {
		m1.lastTimeTick++
		m1.currentSeqNumber = m1.minSeqNumber
		m1.isOverCost = true
		m1.overCostCountInOneTerm++
		return m1.CalcId(m1.lastTimeTick)
	}
	return m1.CalcId(m1.lastTimeTick)
}

func (m1 *SnowWorkerM1) NextNormalId() int64 {
	currentTimeTick := m1.GetCurrentTimeTick()
	if currentTimeTick < m1.lastTimeTick {
		if m1.turnBackTimeTick < 1 {
			m1.turnBackTimeTick = m1.lastTimeTick - 1
			m1.turnBackIndex++
			// 每毫秒序列数的前5位是预留位，0用于手工新值，1-4是时间回拨次序
			// 支持4次回拨次序（避免回拨重叠导致ID重复），可无限次回拨（次序循环使用）。
			if m1.turnBackIndex > 4 {
				m1.turnBackIndex = 1
			}
		}
		return m1.CalcTurnBackId(m1.turnBackTimeTick)
	}
	// 时间追平时，_TurnBackTimeTick清零
	if m1.turnBackTimeTick > 0 {
		m1.turnBackTimeTick = 0
	}
	if currentTimeTick > m1.lastTimeTick {
		m1.lastTimeTick = currentTimeTick
		m1.currentSeqNumber = m1.minSeqNumber
		return m1.CalcId(m1.lastTimeTick)
	}
	if m1.currentSeqNumber > m1.maxSeqNumber {
		m1.lastTimeTick++
		m1.currentSeqNumber = m1.minSeqNumber
		m1.isOverCost = true
		m1.overCostCountInOneTerm = 1
		return m1.CalcId(m1.lastTimeTick)
	}
	return m1.CalcId(m1.lastTimeTick)
}
func (m1 *SnowWorkerM1) CalcId(useTimeTick int64) int64 {
	result := useTimeTick<<m1.timestampShift + int64(m1.workerId<<m1.seqBitLength) + int64(m1.currentSeqNumber)
	m1.currentSeqNumber++
	return result
}
func (m1 *SnowWorkerM1) CalcTurnBackId(useTimeTick int64) int64 {
	result := useTimeTick<<m1.timestampShift + int64(m1.workerId<<m1.seqBitLength) + int64(m1.turnBackIndex)
	m1.turnBackTimeTick--
	return result
}
func (m1 *SnowWorkerM1) GetCurrentTimeTick() int64 {
	var millis = time.Now().UnixNano() / 1e6
	return millis - m1.baseTime
}
func (m1 *SnowWorkerM1) GetNextTimeTick() int64 {
	tempTimeTicker := m1.GetCurrentTimeTick()
	for tempTimeTicker <= m1.lastTimeTick {
		time.Sleep(time.Duration(1) * time.Millisecond)
		tempTimeTicker = m1.GetCurrentTimeTick()
	}
	return tempTimeTicker
}
func (m1 *SnowWorkerM1) NextId() int64 {
	m1.Lock()
	defer m1.Unlock()
	if m1.isOverCost {
		return m1.NextOverCostId()
	} else {
		return m1.NextNormalId()
	}
}
func (m1 *SnowWorkerM1) NextIdStr() string {
	return strconv.FormatInt(m1.NextId(), 10)
}
func (m1 *SnowWorkerM1) ExtractTime(id int64) time.Time {
	return time.UnixMilli(id>>(m1.workerIdBitLength+m1.seqBitLength) + m1.baseTime)
}
