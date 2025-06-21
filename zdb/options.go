package zdb

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/zohu/zgin/zutil"
	"time"
)

type Options struct {
	Host              string        `yaml:"host" binding:"required" note:"数据库地址"`
	Port              string        `yaml:"port" binding:"required" note:"数据库端口"`
	User              string        `yaml:"user" binding:"required" note:"数据库用户"`
	Pass              string        `yaml:"pass" binding:"required" note:"数据库密码"`
	DB                string        `yaml:"db" binding:"required" note:"数据库名"`
	Config            string        `yaml:"config" note:"数据库配置"`
	MaxIdle           int           `yaml:"max_idle" note:"最大闲置连接数"`
	MaxAlive          int           `yaml:"max_alive" note:"最大存活连接数"`
	MaxAliveLife      time.Duration `yaml:"max_alive_life" note:"最大存活时间"`
	LogSlow           time.Duration `yaml:"log_slow" note:"慢阈值，秒"`
	LogIgnoreNotFound string        `yaml:"log_ignore_not_found" note:"忽略无记录错误,yes/no"`
	Debug             *bool         `yaml:"debug" note:"是否开启debug日志"`
	Extension         []string      `yaml:"extension" note:"扩展配置"`
}

func (o *Options) Validate() error {
	o.Config = zutil.FirstTruth(o.Config, "sslmode=disable TimeZone=Asia/Shanghai")
	o.MaxIdle = zutil.FirstTruth(o.MaxIdle, 10)
	o.MaxAlive = zutil.FirstTruth(o.MaxAlive, 100)
	o.MaxAliveLife = zutil.FirstTruth(o.MaxAliveLife, time.Hour)
	o.LogSlow = zutil.FirstTruth(o.LogSlow, time.Second*5)
	o.LogIgnoreNotFound = zutil.FirstTruth(o.LogIgnoreNotFound, "yes")
	return validator.New().Struct(o)
}
func (o *Options) Dsn(database string) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s %s",
		o.Host,
		o.Port,
		o.User,
		o.Pass,
		database,
		o.Config,
	)
}
