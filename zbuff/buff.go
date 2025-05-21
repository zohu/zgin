package zbuff

import (
	"bytes"
	"sync"
)

type Buffer struct {
	*bytes.Buffer
}

var buff = sync.Pool{
	New: func() any {
		return &Buffer{
			bytes.NewBuffer(nil),
		}
	},
}

func New() *Buffer {
	return buff.Get().(*Buffer)
}

func (b *Buffer) Free() {
	b.Reset()
	buff.Put(b)
}

func (b *Buffer) WriteStringIf(ok bool, s string) (int, error) {
	if !ok {
		return 0, nil
	}
	return b.WriteString(s)
}
