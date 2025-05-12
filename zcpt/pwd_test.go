package zcpt

import "testing"

const (
	uid = "0001"
	pwd = "mored@2022"
)

var hashedPwd string

func TestNewPwd(t *testing.T) {
	hashedPwd = NewPwd(uid, pwd)
	t.Log(hashedPwd)
}

func TestVerifyPwd(t *testing.T) {
	if !VerifyPwd(uid, hashedPwd, pwd) {
		t.Errorf("密码校验失败，明文=%s，密文=%s", pwd, hashedPwd)
	}
}
