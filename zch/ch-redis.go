package zch

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/zohu/zgin/zutil"
	"github.com/zohu/zlog"
	"net"
	"strings"
)

type Redis struct {
	redis.UniversalClient
}

func NewRedis(options *Options) *Redis {
	options = zutil.FirstTruth(options, &Options{})
	if err := options.Validate(); err != nil {
		zlog.Fatalf("options is invalid: %v", err)
		return nil
	}
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:      options.Addrs,
		ClientName: options.ClientName,
		DB:         options.Database,
		Password:   options.Password,
	})
	if options.Prefix != "" {
		hook := NewPrefixHook(options.Prefix)
		client.AddHook(hook)
	}
	return &Redis{
		UniversalClient: client,
	}
}

func (r *Redis) BatchDelete(ctx context.Context, pattern string, batchSize ...int64) (int64, error) {
	batchSize = append(batchSize, 1000)
	var cursor uint64
	var keys []string
	var total int64
	for {
		var err error
		keys, cursor, err = r.Scan(ctx, cursor, pattern, batchSize[0]).Result()
		if err != nil {
			return 0, fmt.Errorf("scan failed: %w", err)
		}
		if len(keys) > 0 {
			count, err := r.Del(ctx, keys...).Result()
			if err != nil {
				return 0, fmt.Errorf("delete failed: %w", err)
			}
			total += count
		}
		if cursor == 0 {
			break
		}
	}
	return total, nil
}

/**
 * redis prefix hook
 */

type PrefixHook struct {
	prefix string
}

func NewPrefixHook(prefix string) PrefixHook {
	return PrefixHook{
		prefix: prefix,
	}
}

func (h PrefixHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}
func (h PrefixHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		addPrefix(h.prefix, cmd)
		return next(ctx, cmd)
	}
}
func (h PrefixHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		for _, cmd := range cmds {
			addPrefix(h.prefix, cmd)
		}
		return next(ctx, cmds)
	}
}

func addPrefix(prefix string, cmd redis.Cmder) {
	args := cmd.Args()
	if len(args) <= 1 {
		return
	}
	name := strings.ToUpper(cmd.Name())
	switch name {
	case "MGET", "DEL":
		for i := 1; i < len(args); i++ {
			args[i] = withPrefix(prefix, args[i].(string))
		}
	case "MSET":
		for i := 1; i < len(args); i += 2 {
			args[i] = withPrefix(prefix, args[i].(string))
		}
	case "SCAN":
		if len(args) > 2 {
			for i := 2; i < len(args); i += 2 {
				if args[i] == "match" && len(args) > i+1 {
					args[i+1] = withPrefix(prefix, args[i+1].(string))
					break
				}
			}
		}
	default:
		if canWithPrefix(name) {
			args[1] = withPrefix(prefix, args[1].(string))
		}
	}
}

func canWithPrefix(name string) bool {
	switch name {
	case "GET", "SET", "EXISTS", "DEL", "TYPE",
		"RPUSH", "LPOP", "RPOP", "LLEN", "LRANGE",
		"SADD", "SREM", "SISMEMBER", "SMEMBERS", "SCARD",
		"HSET", "HMSET", "HGET", "HGETALL",
		"ZADD", "ZRANGE", "ZRANGEBYSCORE", "ZREVRANGEBYSCORE", "ZREM",
		"INCR", "INCRBY", "INCRBYFLOAT",
		"WATCH", "MULTI", "EXEC", "EXPIRE":
		return true
	default:
		return false
	}
}
func withPrefix(prefix string, key string) string {
	if strings.HasPrefix(key, prefix+":") {
		return key
	}
	return prefix + ":" + key
}
