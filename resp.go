package zgin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/zohu/zgin/zutil"
	"gorm.io/gorm"
	"net/http"
	"reflect"
	"strings"
)

type Empty struct{}

type RespBean struct {
	Code    int               `json:"code" xml:"code"`
	Data    any               `json:"data" xml:"data"`
	Message string            `json:"message" xml:"message"`
	Notes   map[string]string `json:"notes" xml:"notes"`
}

type RespListBean[T any] struct {
	Page  int `json:"page" xml:"page"`
	Size  int `json:"size" xml:"size"`
	Total int `json:"total" xml:"total"`
	List  []T `json:"list" xml:"list"`
}

func NewResp(code int, message string, data any, notes map[string]string) RespBean {
	return RespBean{
		Code:    zutil.FirstTruth(code, 1),
		Data:    data,
		Message: zutil.FirstTruth(message, "ok"),
		Notes:   notes,
	}
}
func NewRespWithData(data any) RespBean {
	return NewResp(1, "", data, nil)
}
func NewRespWithList[T any](data *RespListBean[T]) RespBean {
	return NewRespWithData(data)
}

func AbortHttpCode(c *gin.Context, code int, resp RespBean) {
	switch c.ContentType() {
	case gin.MIMEPlain:
		c.String(code, fmt.Sprintf("%s", resp.Data))
	case gin.MIMEXML, gin.MIMEXML2:
		c.XML(code, resp)
	default:
		c.JSON(code, resp)
	}
	c.Abort()
}
func Abort(c *gin.Context, resp RespBean) {
	AbortHttpCode(c, http.StatusOK, resp)
}

func translateErrors(h any, errs validator.ValidationErrors) map[string]string {
	ets := make(map[string]string)
	elem := reflect.TypeOf(h)
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}
	for _, err := range errs {
		field, _ := elem.FieldByName(err.StructField())
		key := strings.Split(field.Tag.Get("json"), ",")[0]
		if msg := field.Tag.Get("message"); msg != "" {
			ets[key] = msg
		} else {
			ets[key] = strings.ReplaceAll(err.Translate(Trans()), err.StructField(), fieldNameZh(field))
		}
	}
	return ets
}

// fieldNameZh
// @Description: 查找字段中文名
// @param field
// @return v
func fieldNameZh(field reflect.StructField) (v string) {
	if v = field.Tag.Get("note"); v != "" {
		return v
	}
	gormTag := field.Tag.Get("gorm")
	if gormTag != "" {
		arr := strings.Split(gormTag, ";")
		for _, tag := range arr {
			if strings.HasPrefix(tag, "comment:") {
				v = strings.TrimPrefix(tag, "comment:")
				v = strings.ReplaceAll(v, " ", "")
				return v
			}
		}
	}
	return field.Name
}

type Pages struct {
	Page int `json:"page" xml:"page" note:"页码"`
	Size int `json:"size" xml:"size" note:"每页数量"`
}

func (p *Pages) PageSizes() (int, int) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Size <= 0 {
		p.Size = 50
	}
	if p.Size > 1000 {
		p.Size = 1000
	}
	return p.Page, zutil.FirstTruth(p.Size, 50)
}
func (p *Pages) ScopePage(db *gorm.DB) *gorm.DB {
	page, size := p.PageSizes()
	return db.Offset((page - 1) * size).Limit(size)
}
