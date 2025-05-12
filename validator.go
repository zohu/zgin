package zgin

import (
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	tzh "github.com/go-playground/validator/v10/translations/zh"
	"regexp"
	"time"
)

/**
 * Tag说明：
 *  - json: json字段名
 *  - message: 自定义错误信息，会覆盖系统错误
 *  - note: 字段中文名，如果没有，则会取gorm内的comment
 *  - regular: 正则校验，参数为正则表达式
 *  - datetime: 时间格式校验，参数可省略或RFC3339
 */
var (
	trans    ut.Translator
	validate *validator.Validate
)

func init() {
	uni := ut.New(zh.New())
	trans, _ = uni.GetTranslator("zh")
	validate = validator.New()
	_ = tzh.RegisterDefaultTranslations(validate, trans)
	// 扩展
	_ = validate.RegisterValidation("datetime", datetime)
	_ = validate.RegisterValidation("regular", regular)
}
func Trans() ut.Translator {
	return trans
}
func Validator() *validator.Validate {
	return validate
}

type FiberValidator struct {
	validate *validator.Validate
}

func (v *FiberValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

func NewFiberValidator() *FiberValidator {
	return &FiberValidator{
		validate: validate,
	}
}

// datetime
// @Description: 时间格式校验 datetime RFC3339
// @param fl
// @return bool
func datetime(fl validator.FieldLevel) bool {
	v := fl.Field().String()
	var layout string
	switch fl.Param() {
	case "RFC3339":
		layout = time.RFC3339
	default:
		layout = time.DateTime
	}
	if _, err := time.Parse(layout, v); err != nil {
		return false
	}
	return true
}

// regular
// @Description: 正则校验 regular
// @param fl
// @return bool
func regular(fl validator.FieldLevel) bool {
	v := fl.Field().String()
	ok, _ := regexp.MatchString(fl.Param(), v)
	return ok
}
