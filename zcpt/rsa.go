package zcpt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// GenerateRSAKey
// @Description: 生成RSA密钥对
// @param bits
// @return *rsa.PrivateKey
// @return *rsa.PublicKey
// @return error
func GenerateRSAKey(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	private, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return private, &private.PublicKey, nil
}

// EncodePrivateRSAKeyToPEM
// @Description: 将私钥转换为PEM格式
// @param privateKey
// @return []byte
func EncodePrivateRSAKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		Type:  "RSA PRIVATE KEY",
	})
}

// EncodePublicRSAKeyToPEM
// @Description: 将公钥转换为PEM格式
// @param publicKey
// @return []byte
// @return error
func EncodePublicRSAKeyToPEM(publicKey *rsa.PublicKey) ([]byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Bytes: pubBytes,
		Type:  "RSA PUBLIC KEY",
	}), nil
}

// RSAEncrypt
// @Description: RSA加密
// @param src
// @param filename
// @return []byte
// @return error
func RSAEncrypt(src, pub []byte) ([]byte, error) {
	// 从数据中找出pem格式的块
	block, _ := pem.Decode(pub)
	if block == nil {
		return nil, fmt.Errorf("public key error")
	}
	// 解析一个der编码的公钥
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	// 公钥加密
	result, _ := rsa.EncryptPKCS1v15(rand.Reader, publicKey.(*rsa.PublicKey), src)
	return result, nil
}

// RSADecrypt
// @Description: RSA解密
// @param src
// @param filename
// @return []byte
// @return error
func RSADecrypt(src, piv []byte) ([]byte, error) {
	// 从数据中解析出pem块
	block, _ := pem.Decode(piv)
	if block == nil {
		return nil, fmt.Errorf("private key error")
	}
	// 解析出一个der编码的私钥
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	// 私钥解密
	result, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, src)
	if err != nil {
		return nil, err
	}
	return result, nil
}
