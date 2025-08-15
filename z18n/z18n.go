package z18n

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/zohu/zgin/zch"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zmap"
	"github.com/zohu/zgin/zutil"
	"golang.org/x/text/language"
	"os"
	"path"
	"strings"
)

/**
i18n 优先从缓存翻译，如果没有则从文件翻译
*/

type Localizer func(ctx context.Context, lang language.Tag, ID string) string

var bundle *i18n.Bundle
var localizers zmap.ConcurrentMap[string, *i18n.Localizer]
var custom Localizer

func init() {
	localizers = zmap.New[*i18n.Localizer]()
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
}

func LoadFile(filepath string) error {
	stat, err := os.Stat(filepath)
	if err != nil {
		return fmt.Errorf("os.stat failed: %v", err)
	}
	var files []string
	if stat.IsDir() {
		de, err := os.ReadDir(filepath)
		if err != nil {
			return fmt.Errorf("os.readdir failed: %v", err)
		}
		for _, f := range de {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".toml") {
				files = append(files, path.Join(filepath, f.Name()))
			}
		}
	} else if strings.HasSuffix(filepath, ".toml") {
		files = append(files, filepath)
	}
	for _, file := range files {
		if _, err = bundle.LoadMessageFile(file); err != nil {
			return fmt.Errorf("load localizer file failed: %v", err)
		}
		zlog.Infof("load localizer file: %s", file)
	}
	return nil
}
func Language(c *gin.Context) language.Tag {
	cookie, _ := c.Cookie("lang")
	lang := zutil.FirstTruth(
		c.Query("lang"),
		cookie,
		c.GetHeader("Accept-Language"),
		language.English.String(),
	)
	t, _, err := language.ParseAcceptLanguage(lang)
	if err != nil {
		zlog.Warnf("parse accept language failed: %v", err)
		return language.English
	}
	t = append(t, language.English)
	switch t[0].String() {
	case "zh", "zh-CN", "zh-Hans", "zh-Hans-CN":
		return language.Chinese
	case "tw", "zh-TW", "zh-HK", "zh-Hant", "zh-Hant-TW":
		return language.TraditionalChinese
	case "en", "en-US", "en-GB":
		return language.English
	case "ja", "ja-JP":
		return language.Japanese
	case "pt", "pt-BR":
		return language.Portuguese
	default:
		return t[0]
	}
}
func NewLocalizer(tag language.Tag) *i18n.Localizer {
	if l, ok := localizers.Get(tag.String()); ok {
		return l
	}
	l := i18n.NewLocalizer(bundle, tag.String())
	localizers.Set(tag.String(), l)
	return l
}
func Localize(c *gin.Context, ID string, kv ...map[string]string) string {
	lang := Language(c)
	if custom != nil && strings.HasPrefix(ID, zch.PrefixI18n.Key()) {
		return custom(c.Request.Context(), lang, ID)
	}
	localizer := NewLocalizer(lang)
	data := map[string]string{}
	if len(kv) > 0 {
		data = kv[0]
	}
	message, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    ID,
		TemplateData: data,
	})
	if err == nil {
		return message
	}
	return ID
}
func LocalizeWithTag(lang language.Tag, ID string, kv ...map[string]string) string {
	if custom != nil && strings.HasPrefix(ID, zch.PrefixI18n.Key()) {
		return custom(context.Background(), lang, ID)
	}
	localizer := NewLocalizer(lang)
	data := map[string]string{}
	if len(kv) > 0 {
		data = kv[0]
	}
	message, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    ID,
		TemplateData: data,
	})
	if err == nil {
		return message
	}
	return ID
}

func AddLocalizer(lc Localizer) {
	custom = lc
}
