package zid

import (
	"strconv"
	"time"
)

var idGenerator *DefaultIdGenerator

func init() {
	if idGenerator == nil {
		idGenerator = NewDefaultIdGenerator(nil)
	}
}

func GeneratorWithOptions(options *Options) {
	idGenerator = NewDefaultIdGenerator(options)
}

// NextId
// @Description: 10进制ID
// @return int64
func NextId() int64 {
	return idGenerator.NextId()
}

// NextIdStr
// @Description: 10进制ID字符串
// @return string
func NextIdStr() string {
	return strconv.FormatInt(NextId(), 10)
}

// NextIdHex
// @Description: 16进制ID字符串
// @return string
func NextIdHex() string {
	return strconv.FormatInt(NextId(), 16)
}

// NextIdShort
// @Description: 36进制ID字符串
// @return string
func NextIdShort() string {
	return strconv.FormatInt(NextId(), 36)
}

// ExtractTime
// @Description: 提取ID时间
// @param id
// @return time.Time
func ExtractTime(id int64) time.Time {
	return idGenerator.ExtractTime(id)
}

func ExtractTimeHex(hex string) time.Time {
	id, _ := strconv.ParseInt(hex, 16, 64)
	return idGenerator.ExtractTime(id)
}

func ExtractTimeShort(short string) time.Time {
	id, _ := strconv.ParseInt(short, 36, 64)
	return idGenerator.ExtractTime(id)
}

// ExtractWorkerId
// @Description: 提取工作节点ID
// @param id
// @return int64
func ExtractWorkerId(id int64) int64 {
	return idGenerator.ExtractWorkerId(id)
}
func ExtractWorkerIdHex(hex string) int64 {
	id, _ := strconv.ParseInt(hex, 16, 64)
	return idGenerator.ExtractWorkerId(id)
}
func ExtractWorkerIdShort(short string) int64 {
	id, _ := strconv.ParseInt(short, 36, 64)
	return idGenerator.ExtractWorkerId(id)
}
