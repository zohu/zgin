package zch

import (
	"github.com/zohu/zgin/zutil"
	"testing"
)

func TestDict(t *testing.T) {
	for i := 0; i < 100; i++ {
		k := zutil.RandomStr(zutil.Random(1, 30))
		t.Logf("%d => %d", len(k), len(dictPrefix(k)))
	}
}
