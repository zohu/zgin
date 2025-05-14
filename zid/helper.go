package zid

import (
	"strconv"
	"sync"
	"time"
)

var singletonMutex sync.Mutex
var idGenerator *DefaultIdGenerator

func init() {
	singletonMutex.Lock()
	idGenerator = NewDefaultIdGenerator(nil)
	singletonMutex.Unlock()
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
	return idGenerator.NextIdStr()
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
