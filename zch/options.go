package zch

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/zohu/zgin/zutil"
	"strings"
	"time"
)

type Options struct {
	Expiration    time.Duration `yaml:"expiration"`
	CleanInterval time.Duration `yaml:"clean_interval"`
	Addrs         []string      `validate:"required" yaml:"addrs"`
	Database      int           `yaml:"database"`
	Password      string        `yaml:"password"`
	Prefix        string        `yaml:"prefix"`
	ClientName    string        `yaml:"client_name"`
}

func (o *Options) Validate() error {
	o.Expiration = zutil.FirstTruth(o.Expiration, time.Hour)
	o.CleanInterval = zutil.FirstTruth(o.CleanInterval, time.Minute*5)
	o.Database = zutil.FirstTruth(o.Database, 0)
	o.ClientName = zutil.FirstTruth(o.ClientName, "zch")
	return validator.New().Struct(o)
}

type Prefix string

func (p Prefix) Key(args ...string) string {
	if len(args) == 0 {
		return string(p)
	}
	return fmt.Sprintf("%s:%s", string(p), strings.Join(args, ":"))
}
