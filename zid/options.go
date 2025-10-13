package zid

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
)

type ISnowWorker interface {
	NextId() int64
	ExtractTime(int64) time.Time
	ExtractWorkerId(id int64) int64
}
type Options struct {
	BaseTime           int64  // 基础时间（ms单位），不能超过当前系统时间
	WorkerId           int64  // 机器码，必须由外部设定，最大值 2^WorkerIdBitLength-1
	WorkerIdAutoPrefix string // 机器码前缀，默认值"zid"
	WorkerIdBitLength  byte   // 机器码位长，默认值6，取值范围 [1, 15]（要求：序列数位长+机器码位长不超过22）
	SeqBitLength       byte   // 序列数位长，默认值6，取值范围 [3, 21]（要求：序列数位长+机器码位长不超过22）
	MaxSeqNumber       uint32 // 最大序列数（含），设置范围 [MinSeqNumber, 2^SeqBitLength-1]，默认值0，表示最大序列数取最大值（2^SeqBitLength-1]）
	MinSeqNumber       uint32 // 最小序列数（含），默认值5，取值范围 [5, MaxSeqNumber]，每毫秒的前5个序列数对应编号0-4是保留位，其中1-4是时间回拨相应预留位，0是手工新值预留位
	TopOverCostCount   uint32 // 最大漂移次数（含），默认2000，推荐范围500-10000（与计算能力有关）
}

func (o *Options) Validate() {
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	o.BaseTime = zutil.FirstTruth(o.BaseTime, baseTime)
	o.WorkerIdAutoPrefix = zutil.FirstTruth(o.WorkerIdAutoPrefix, "zid")
	o.WorkerIdBitLength = zutil.FirstTruth(o.WorkerIdBitLength, 6)
	o.SeqBitLength = zutil.FirstTruth(o.SeqBitLength, 6)
	o.MaxSeqNumber = zutil.FirstTruth(o.MaxSeqNumber, 0)
	o.MinSeqNumber = zutil.FirstTruth(o.MinSeqNumber, 5)
	o.TopOverCostCount = zutil.FirstTruth(o.TopOverCostCount, 2000)

	if o.BaseTime < baseTime || o.BaseTime > time.Now().UnixMilli() {
		zlog.Fatalf("BaseTime range:[2025-01-01 ~ now]")
	}
	if o.WorkerIdBitLength < 1 || o.WorkerIdBitLength > 21 {
		zlog.Fatalf("WorkerIdBitLength range:[1, 21]")
	}
	if o.WorkerIdBitLength+o.SeqBitLength > 22 {
		zlog.Fatalf("WorkerIdBitLength + SeqBitLength <= 22")
	}
	maxWorkerIdNumber := o.MaxWorkerIdNumber()
	if o.WorkerId < 0 || o.WorkerId > maxWorkerIdNumber {
		zlog.Fatalf("WorkerId range:[0, " + strconv.FormatUint(uint64(maxWorkerIdNumber), 10) + "]")
	}
	if o.SeqBitLength < 2 || o.SeqBitLength > 21 {
		zlog.Fatalf("SeqBitLength range:[2, 21]")
	}
	maxSeqNumber := o.maxSeqNumber()
	if o.MaxSeqNumber < 0 || o.MaxSeqNumber > maxSeqNumber {
		zlog.Fatalf("MaxSeqNumber range:[1, " + strconv.FormatUint(uint64(maxSeqNumber), 10) + "]")
	}
	if o.MinSeqNumber < 5 || o.MinSeqNumber > maxSeqNumber {
		zlog.Fatalf("MinSeqNumber range:[5, " + strconv.FormatUint(uint64(maxSeqNumber), 10) + "]")
	}
	if o.TopOverCostCount < 0 || o.TopOverCostCount > 10000 {
		zlog.Fatalf("TopOverCostCount range:[0, 10000]")
	}
}
func (o *Options) MaxWorkerIdNumber() int64 {
	return zutil.FirstTruth(int64(1<<o.WorkerIdBitLength)-1, 63)
}
func (o *Options) maxSeqNumber() uint32 {
	return zutil.FirstTruth(uint32(1<<o.SeqBitLength)-1, 63)
}
func (o *Options) prefix(wid int64) string {
	return fmt.Sprintf("%s:%d", strings.TrimSuffix(o.WorkerIdAutoPrefix, ":"), wid)
}
