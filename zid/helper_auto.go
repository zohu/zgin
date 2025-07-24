package zid

import (
	"context"
	"github.com/zohu/zgin/zch"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"time"
)

func AutoWorkerId(r *zch.Redis, options *Options) {
	options = zutil.FirstTruth(options, new(Options))
	options.Validate()

	ctx := context.TODO()
	options.WorkerId = findIdx(ctx, r, options, 0)

	singletonMutex.Lock()
	idGenerator = NewDefaultIdGenerator(options)
	singletonMutex.Unlock()

	go alive(ctx, r, options.prefix(options.WorkerId))
	zlog.Infof("init zid success, workerid=%d", options.WorkerId)
}

func findIdx(ctx context.Context, r *zch.Redis, ops *Options, retry uint16) uint16 {
	if retry > ops.maxWorkerIdNumber() {
		zlog.Fatalf("all worker id [0-%d] are occupied, please extend WorkerIdBitLength", retry-1)
	}
	ok, err := r.SetNX(ctx, ops.prefix(retry), "occupied", time.Second*60).Result()
	if ok {
		return retry
	}
	if err != nil {
		zlog.Warnf("find worker id error: %v", err)
	}
	return findIdx(ctx, r, ops, retry+1)
}
func alive(ctx context.Context, r *zch.Redis, prefix string) {
	for range time.NewTicker(time.Second * 40).C {
		r.Set(ctx, prefix, "occupied", time.Second*60)
	}
}
