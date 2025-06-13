package zmiddle

import (
	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"golang.org/x/text/language"
	"os"
	"path"
	"strings"
	"sync"
)

const LocalizerDir = "./locales"

func NewLocalizer() gin.HandlerFunc {
	if _, err := os.Stat(LocalizerDir); err != nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	files, err := os.ReadDir(LocalizerDir)
	if err != nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	var locals []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".toml") {
			locals = append(locals, path.Join(LocalizerDir, f.Name()))
		}
	}
	if len(locals) == 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	zlog.Infof("middleware localizer enabled")

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	for _, file := range locals {
		if _, err = bundle.LoadMessageFile(file); err != nil {
			zlog.Fatalf("load localizer file failed: %v", err)
		}
		zlog.Infof("load localizer file: %s", file)
	}
	var localizerCache sync.Map
	return func(c *gin.Context) {
		cookie, _ := c.Cookie("lang")
		lang := zutil.FirstTruth(
			c.Query("lang"),
			cookie,
			c.GetHeader("Accept-Language"),
			language.English.String(),
		)
		v, ok := localizerCache.Load(lang)
		if !ok {
			v = i18n.NewLocalizer(bundle, lang)
			localizerCache.Store(lang, v)
		}
		c.Set("localizer", v)
		c.Next()
	}
}
func Localizer(c *gin.Context) *i18n.Localizer {
	return c.MustGet("localizer").(*i18n.Localizer)
}
