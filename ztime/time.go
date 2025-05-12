package ztime

import (
	"fmt"
	"time"
)

const (
	FormatTimeNoYear = "01-02 15:04:05.000"
	FormatDate       = "2006-01-02"
	FormatString     = "20060102150405"
	FormatDateString = "20060102"
)

func SecondToString(t int64) string {
	return (time.Duration(t) * time.Second).String()
}
func StringToDuration(str string) time.Duration {
	t, err := time.ParseDuration(str)
	if err != nil {
		return time.Duration(0)
	}
	return t
}

func TimePeriodString(t time.Time) string {
	n := time.Now()
	tt := n.Sub(t).Seconds()
	ty, tw := t.ISOWeek()
	ny, nw := n.ISOWeek()
	if tt < 10 {
		return "数秒内"
	} else if tt < 30 {
		return "半分钟内"
	} else if tt < 60 {
		return "一分钟内"
	} else if tt < 60*5 {
		return "五分钟内"
	} else if tt < 60*30 {
		return "半小时内"
	} else if tt < 60*60 {
		return "一小时内"
	} else if tt < 60*60*24 {
		return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
	} else if tt < 60*60*24*2 {
		return fmt.Sprintf("昨天 %02d:%02d", t.Hour(), t.Minute())
	} else if tt < 60*60*24*3 {
		return fmt.Sprintf("前天 %02d:%02d", t.Hour(), t.Minute())
	} else if ty == ny && tw == nw {
		return week2string(t)
	}
	return fmt.Sprintf("%s %02d:%02d", t.Format(FormatDate), t.Hour(), t.Minute())
}

func week2string(t time.Time) string {
	w := t.Weekday()
	h := t.Hour()
	m := t.Minute()
	switch w {
	case 0:
		return fmt.Sprintf("周日 %02d:%02d", h, m)
	case 1:
		return fmt.Sprintf("周一 %02d:%02d", h, m)
	case 2:
		return fmt.Sprintf("周二 %02d:%02d", h, m)
	case 3:
		return fmt.Sprintf("周三 %02d:%02d", h, m)
	case 4:
		return fmt.Sprintf("周四 %02d:%02d", h, m)
	case 5:
		return fmt.Sprintf("周五 %02d:%02d", h, m)
	case 6:
		return fmt.Sprintf("周六 %02d:%02d", h, m)
	}
	return ""
}

func FormatTimeMust(layout, str string) time.Time {
	t, err := time.Parse(layout, str)
	if err != nil {
		return time.Time{}
	}
	return t
}
