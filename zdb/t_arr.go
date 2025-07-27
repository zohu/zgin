package zdb

import (
	"database/sql/driver"
	"github.com/bytedance/sonic"
	"github.com/lib/pq"
	"slices"
	"sort"
	"strings"
)

type StringArray struct {
	pq.StringArray
}

func NewStringArray(arr []string) *StringArray {
	return &StringArray{arr}
}

// gorm 支持

func (a *StringArray) GormDataType() string {
	return "text[]"
}
func (a *StringArray) String() string {
	return strings.Join(a.StringArray, ",")
}
func (a *StringArray) Scan(src interface{}) error {
	return a.StringArray.Scan(src)
}
func (a *StringArray) Value() (driver.Value, error) {
	return a.StringArray.Value()
}

// JSON 支持

func (a *StringArray) UnmarshalJSON(data []byte) error {
	var strs []string
	if err := sonic.Unmarshal(data, &strs); err != nil {
		return err
	}
	a.StringArray = strs
	return nil
}

func (a *StringArray) MarshalJSON() ([]byte, error) {
	return sonic.Marshal([]string(a.StringArray))
}
func (a *StringArray) Contains(v string) bool {
	return slices.Contains(a.StringArray, v)
}
func (a *StringArray) Append(v string) *StringArray {
	if v == "" {
		return a
	}
	a.StringArray = append(a.StringArray, v)
	return a
}
func (a *StringArray) AppendOnce(v string) *StringArray {
	if a.Contains(v) {
		return a
	}
	return a.Append(v)
}
func (a *StringArray) Equal(b *StringArray) bool {
	return slices.Equal(a.StringArray, b.StringArray)
}

// ItemEqual
// @Description: 与顺序无关
// @receiver a
// @param b
// @return bool
func (a *StringArray) ItemEqual(b *StringArray) bool {
	if len(a.StringArray) != len(b.StringArray) {
		return false
	}
	if len(a.StringArray) == 0 {
		return true // 两个空切片视为相等
	}
	aCopy := make([]string, len(a.StringArray))
	copy(aCopy, a.StringArray)
	bCopy := make([]string, len(b.StringArray))
	copy(bCopy, b.StringArray)

	sort.Strings(aCopy)
	sort.Strings(bCopy)
	return slices.Equal(aCopy, bCopy)
}
