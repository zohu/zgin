package zdb

import (
	"database/sql/driver"

	"github.com/zohu/zgin/zcpt"
)

type CptString struct {
	bytes []byte
}

func NewCptString(s, uid string) *CptString {
	b, _ := zcpt.AesEncryptCBC([]byte(s), []byte(zcpt.Md5(uid)))
	return &CptString{
		bytes: b,
	}
}

func (a *CptString) GormDataType() string {
	return "BYTEA"
}
func (a *CptString) String() string {
	return string(a.bytes)
}
func (a *CptString) Scan(src interface{}) error {
	a.bytes = src.([]byte)
	return nil
}
func (a *CptString) Value() (driver.Value, error) {
	return a.bytes, nil
}

func (a *CptString) Decrypt(uid string) string {
	b, _ := zcpt.AesDecryptCBC(a.bytes, []byte(zcpt.Md5(uid)))
	return string(b)
}
