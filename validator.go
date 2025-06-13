package zgin

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"regexp"
	"time"
)

/**
 * Tag说明：
 *  - json: json字段名
 *  - message: 自定义错误信息，会覆盖系统错误
 *  - regular: 正则校验，参数为正则表达式
 *  - datetime: 时间格式校验，参数可省略或RFC3339
 */

func init() {
	validate := Validator()
	_ = validate.RegisterValidation("datetime", datetime)
	_ = validate.RegisterValidation("regular", regular)
}

func Validator() *validator.Validate {
	validate, _ := binding.Validator.Engine().(*validator.Validate)
	return validate
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
