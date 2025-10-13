### 增强型雪花 ID 生成器（Snowflake）

> 在标准 Snowflake 的基础上，增加了对时钟回拨（clock drift / time rollback）的精细处理机制，并支持自定义位分配（如机器 ID 位长、序列号位长等）。

一、设计亮点
1. 灵活的位分配
   动态配置机器 ID 与序列号的位数，不再硬编码。
   支持非标准 Snowflake 布局（例如 12 位机器 ID + 10 位序列号）。

2. 时钟回拨处理机制
   引入“唯一性回拨序号”，“容忍回拨 + 避免重复” 的实用策略，比简单抛异常或阻塞更优雅。

3. 过载（OverCost）处理
   当单毫秒内序列号耗尽，进入“过载模式”：
   - 自动推进时间戳；
   - 记录过载次数，超过阈值后强制等待真实时间追上；
   - 防止因突发高并发导致 ID 生成“超前”太多；

二、使用方法
1. 引入依赖
```bash
go get github.com/zohu/zid
```
2. 配置项

| 字段	                | 类型	     | 默认值	                            | 有效范围	                               | 说明                                                               |
|--------------------|---------|---------------------------------|-------------------------------------|------------------------------------------------------------------|
| BaseTime	          | int64	  | 2025-01-01 00:00:00 UTC（毫秒时间戳）	 | [2025-01-01 UTC, 当前时间]	             | Snowflake 的纪元起点（单位：毫秒）。必须 ≤ 当前系统时间，且建议固定以保证 ID 全局可排序             |
| WorkerId	          | int64	  | 0	                              | [0, 2^WorkerIdBitLength - 1]	       | 节点唯一标识符。若启用自动分配（如 AutoWorkerId），无需手动设置；否则需确保集群内唯一                |
| WorkerIdAutoPrefix | string  | "zid"                           | 任意非空字符串                             | 自动分配 WorkerId 时在 Redis 中使用的键前缀，用于隔离不同服务。建议结合服务名使用（如 "order:zid"） |
| WorkerIdBitLength  | byte    | 6                               | [1, 64]                             | WorkerId 所占位数。决定最大支持节点数为 2^WorkerIdBitLength（如 6 位 → 最多 64 个节点）  |
| SeqBitLength       | 	byte   | 	6	                             | [2, 21]                             | 	序列号所占位数。决定每毫秒最大生成 ID 数为 2^SeqBitLength（如 6 位 → 每毫秒 64 个）        | 
| MaxSeqNumber       | 	uint32 | 	2^SeqBitLength - 1（即最大值）       | 	[MinSeqNumber, 2^SeqBitLength - 1] | 	每毫秒允许使用的最大序列号。设为 0 表示使用理论最大值                                    |
| MinSeqNumber       | 	uint32 | 	5                              | 	[5, MaxSeqNumber]                  | 	每毫秒保留的最小序列号。编号 0~4 为系统保留位： 0：手工新值预留  1~4：时间回拨应急预留               |
| TopOverCostCount	  | uint32	 | 2000	                           | [0, 10000]	                         | 允许的最大时钟漂移（回拨）补偿次数。值越大容忍度越高，但内存/计算开销增加。推荐 500~10000               |

3. 使用示例
```
# 内置自动初始化，单体服务可以直接使用，WorkerId固定为0

# 纯数字ID
zid.NextId() int64

# 字符串，内容同数字ID
zid.NextIdStr() string

# 16进制ID字符串
zid.NextIdHex() string

# 36进制ID字符串, 更短
zid.NextIdShort() string

# 提取ID的生成时间
zid.ExtractTime(id int64) time.Time
zid.ExtractTimeHex(hex string) time.Time
zid.ExtractTimeShort(short string) time.Time

# 获取WorkerId
zid.ExtractWorkerId(id int64) int64
zid.ExtractWorkerIdHex(hex string) int64
zid.ExtractWorkerIdShort(short string) int64
```

4. 开启自动分配WorkerId
```
# 内置了一个基于 Redis 的分布式锁自动分配 workerId 逻辑
AutoWorkerId(r redis.UniversalClient, options *Options)
# 然后就可以正常使用了
id := zid.NextId()

----

# 如果不想用内置的，可以自己实现，逻辑如下：
# 1. 自己实现逻辑生成options
# 2. 调用GeneratorWithOptions(options *Options)即可替换全局生成器

# 如 利用 IP + MAC，假设你已经实现了getPrimaryIPAndMAC() (ipStr, macStr string)
ip, mac := getPrimaryIPAndMAC()
# 组合字符串
key := ip + "|" + mac
# SHA256 哈希
hash := sha256.Sum256([]byte(key))
# 取前 8 字节转为 uint64，再取模
id := binary.BigEndian.Uint64(hash[:8])
workerId := int64(id % uint64(maxWorkerId+1))

GeneratorWithOptions(&Options{
    WorkerId: workerId,
})

# 然后就可以正常使用了
id := zid.NextId()

# 如果用在k8s集群中，可以使用 pod name 的 hash
```