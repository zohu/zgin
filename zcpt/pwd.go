package zcpt

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// NewPwd
// @Description: 用uid+pwd对密码加工，使密码数据每次不一样且不可逆
// @param uid
// @param pwd
// @return string
func NewPwd(uid string, pwd string) string {
	pwd = fmt.Sprintf("%s@%s", pwd, uid)
	hash, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(hash)
}

// VerifyPwd
// @Description: 校验密码
// @param uid
// @param cptPwd 密文
// @param pwd 明文
// @return bool
func VerifyPwd(uid string, cptPwd, pwd string) bool {
	pwd = fmt.Sprintf("%s@%s", pwd, uid)
	err := bcrypt.CompareHashAndPassword([]byte(cptPwd), []byte(pwd))
	return err == nil
}
