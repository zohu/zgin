package zid

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
)

func AutoWorkerId(r redis.UniversalClient, options *Options) {
	options = zutil.FirstTruth(options, new(Options))
	options.Validate()

	options.WorkerId = findIdx(r, options)
	GeneratorWithOptions(options)

	go alive(r, options.prefix(options.WorkerId))
	zlog.Infof("init zid success, workerid=%d", options.WorkerId)
}

func findIdx(r redis.UniversalClient, ops *Options) int64 {
	mw := ops.MaxWorkerIdNumber()
	for i := int64(0); i <= mw; i++ {
		if r.SetNX(context.Background(), ops.prefix(i), "occupied", time.Second*60).Val() {
			return i
		}
	}
	zlog.Fatalf("all worker id [0-%d] are occupied, please extend WorkerIdBitLength", mw)
	return 0
}
func alive(r redis.UniversalClient, prefix string) {
	for range time.NewTicker(time.Second * 40).C {
		r.Set(context.Background(), prefix, "occupied", time.Second*60)
	}
}
