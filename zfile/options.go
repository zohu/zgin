package zfile

type ProviderType string

const (
	ProviderTypeOss ProviderType = "oss"
)

type Options struct {
	ProviderType ProviderType `binding:"required" note:"存储类型"`
	AccessKey    string       `binding:"required" note:"访问密钥"`
	AccessSecret string       `binding:"required" note:"访问密钥"`
	Region       string       `binding:"required" note:"区域"`
	Endpoint     string       `binding:"required" note:"接入点"`
	Domain       string       `note:"访问时域名，如果为空，则使用接入点"`
	Bucket       string       `binding:"required" note:"存储桶"`
	Prefix       string       `note:"存储桶前缀"`
	IdleDays     int64        `note:"最长闲置时间，不设置则永久"`
	isPvMode     bool
}
