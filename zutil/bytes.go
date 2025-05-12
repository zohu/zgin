package zutil

import (
	"bytes"
	"sync"
)

var buff = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

func NewByteBuff() *bytes.Buffer {
	return buff.Get().(*bytes.Buffer)
}
func ReleaseByteBuff(b *bytes.Buffer) {
	b.Reset()
	buff.Put(b)
}
