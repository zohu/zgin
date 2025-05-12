package zcpt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// ========================= CBC =========================

func AesEncryptCBC(data, key []byte) (encrypted []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData := PKCS5Padding(data, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	encrypted = make([]byte, len(origData))
	blockMode.CryptBlocks(encrypted, origData)
	return encrypted, nil
}

func AesDecryptCBC(encrypted, key []byte) (decrypted []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	decrypted = make([]byte, len(encrypted))
	blockMode.CryptBlocks(decrypted, encrypted)
	decrypted = PKCS5UnPadding(decrypted)
	return decrypted, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	text := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, text...)
}
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}

// ========================= ECB =========================

func AesEncryptECB(data, key []byte) (encrypted []byte, err error) {
	cp, err := aes.NewCipher(generateKey(key))
	if err != nil {
		return nil, err
	}
	length := (len(data) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)
	copy(plain, data)
	pad := byte(len(plain) - len(data))
	for i := len(data); i < len(plain); i++ {
		plain[i] = pad
	}
	encrypted = make([]byte, len(plain))
	for bs, be := 0, cp.BlockSize(); bs <= len(data); bs, be = bs+cp.BlockSize(), be+cp.BlockSize() {
		cp.Encrypt(encrypted[bs:be], plain[bs:be])
	}
	return encrypted, nil
}

func AesDecryptECB(encrypted, key []byte) (decrypted []byte, err error) {
	cp, err := aes.NewCipher(generateKey(key))
	if err != nil {
		return nil, err
	}
	decrypted = make([]byte, len(encrypted))
	for bs, be := 0, cp.BlockSize(); bs < len(encrypted); bs, be = bs+cp.BlockSize(), be+cp.BlockSize() {
		cp.Decrypt(decrypted[bs:be], encrypted[bs:be])
	}
	trim := 0
	if len(decrypted) > 0 {
		trim = len(decrypted) - int(decrypted[len(decrypted)-1])
	}
	if trim < 0 || trim > len(decrypted) {
		return nil, errors.New("aes decrypt error")
	}
	return decrypted[:trim], nil
}

func generateKey(key []byte) []byte {
	genKey := make([]byte, 16)
	copy(genKey, key)
	for i := 16; i < len(key); {
		for j := 0; j < 16 && i < len(key); j, i = j+1, i+1 {
			genKey[j] ^= key[i]
		}
	}
	return genKey
}

// ========================= CFB =========================

func AesEncryptCFB(data, key []byte) (encrypted []byte, err error) {
	cp, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	encrypted = make([]byte, aes.BlockSize+len(data))
	iv := encrypted[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(cp, iv)
	stream.XORKeyStream(encrypted[aes.BlockSize:], data)
	return encrypted, nil
}

func AesDecryptCFB(encrypted, key []byte) (decrypted []byte, err error) {
	cp, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(cp, iv)
	stream.XORKeyStream(encrypted, encrypted)
	return encrypted, nil
}
