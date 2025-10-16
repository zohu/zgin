package zcpt

import "testing"

func TestNewPwd(t *testing.T) {
	uid := "0001"
	pwd := "zcpt@2025"
	hashedPwd := NewPwd(uid, pwd)
	t.Log(hashedPwd)
	if !VerifyPwd(uid, hashedPwd, pwd) {
		t.Errorf("密码校验失败，明文=%s，密文=%s", pwd, hashedPwd)
	}
}
