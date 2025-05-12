package ztime

// #include "build_time.h"
import "C"
import (
	"time"
)

var ct = C.GoString(C.build_time())

func BuildTime() time.Time {
	bt, _ := time.ParseInLocation("Jan 02 2006 15:04:05", ct, time.Local)
	return bt
}
