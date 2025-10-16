package zfile

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/dromara/carbon/v2"
	"github.com/go-playground/validator/v10"
	"github.com/zohu/zgin/zutil"
)

type ProviderType string

const (
	ProviderTypeOss ProviderType = "oss"
	ProviderTypeS3  ProviderType = "s3"
)

type Options struct {
	Provider     ProviderType `json:"provider" yaml:"provider" validate:"required" note:"存储服务商类型"`
	AccessKey    string       `json:"access_key" yaml:"access_key" validate:"required" note:"AccessKey"`
	AccessSecret string       `json:"access_secret" yaml:"access_secret" validate:"required" note:"AccessSecret"`
	Region       string       `json:"region" yaml:"region" validate:"required" note:"区域"`
	Endpoint     string       `json:"endpoint" yaml:"endpoint" validate:"required" note:"接入点"`
	Domain       string       `json:"domain" yaml:"domain" note:"访问时域名，如果为空，则使用接入点"`
	Bucket       string       `json:"bucket" yaml:"bucket" validate:"required" note:"存储桶"`
	Prefix       string       `json:"prefix" yaml:"prefix" note:"存储桶前缀"`
	IdleDays     int64        `json:"idle_days" yaml:"idle_days"  gorm:"comment:最长闲置时间，不设置则永久"`
	MaxRetry     int          `json:"max_retry" yaml:"max_retry" note:"最大重试次数"`
}

func (c *Options) Validate() error {
	c.Domain = zutil.FirstTruth(c.Domain, c.Endpoint)
	if !strings.HasPrefix(c.Domain, "http") {
		c.Domain = fmt.Sprintf("https://%s", strings.TrimSuffix(c.Domain, "/"))
	}
	c.Prefix = strings.TrimPrefix(c.Prefix, "/")
	c.Prefix = strings.TrimSuffix(c.Prefix, "/")
	c.MaxRetry = zutil.FirstTruth(c.MaxRetry, 3)
	return validator.New().Struct(c)
}
func (c *Options) HTTPDomain(filename string) string {
	return fmt.Sprintf("%s/%s", c.Domain, strings.TrimPrefix(filename, "/"))
}
func (c *Options) FullName(args ...string) string {
	path := c.Prefix
	for _, arg := range args {
		if strings.TrimSpace(arg) != "" {
			if strings.HasPrefix(arg, ".") {
				path = fmt.Sprintf("%s%s", path, arg)
			} else {
				path = fmt.Sprintf("%s/%s", path, arg)
			}
		}
	}
	return path
}

// ZfileRecord
// @Description: 自动化pv管理需要的数据表
type ZfileRecord struct {
	Id        uint64         `json:"id" gorm:"->;primarykey"`
	Fid       string         `json:"fid" gorm:"unique;comment:文件ID"`
	Md5       string         `json:"md5" gorm:"unique;comment:文件MD5"`
	Bucket    string         `json:"bucket" gorm:"comment:存储桶"`
	Name      string         `json:"name" gorm:"comment:桶内名称"`
	Pv        int64          `json:"pv" gorm:"comment:访问次数"`
	Expire    int64          `json:"expire" gorm:"comment:保存天数"`
	CreatedAt *carbon.Carbon `json:"created_at,omitempty" gorm:"autoCreateTime"`
	UpdatedAt *carbon.Carbon `json:"updated_at,omitempty" gorm:"autoUpdateTime"`
}
type Progress func(increment, transferred, total int64)

// iService
// @Description: 服务商要实现的接口
type iService interface {
	upload(ctx context.Context, r io.ReadSeeker, name string, progress Progress) error
	delete(ctx context.Context, name string) (err error)
}
