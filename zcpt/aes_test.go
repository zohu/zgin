package zcpt

import "testing"

func TestAes(t *testing.T) {
	text := "hello world"
	key := "315c2wd6vpc7q4hx"
	if d, err := AesEncryptCBC([]byte(text), []byte(key)); err != nil {
		t.Error(err)
	} else {
		t.Logf("cbc cipher -> %s", string(d))
		if d, err := AesDecryptCBC(d, []byte(key)); err != nil {
			t.Error(err)
		} else {
			t.Logf("cbc str -> %s", string(d))
		}
	}
	if d, err := AesEncryptECB([]byte(text), []byte(key)); err != nil {
		t.Error(err)
	} else {
		t.Logf("ecb cipher -> %s", string(d))
		if d, err := AesDecryptECB(d, []byte(key)); err != nil {
			t.Error(err)
		} else {
			t.Logf("ecb str -> %s", string(d))
		}
	}
	if d, err := AesEncryptCFB([]byte(text), []byte(key)); err != nil {
		t.Error(err)
	} else {
		t.Logf("cfb cipher -> %s", string(d))
		if d, err := AesDecryptCFB(d, []byte(key)); err != nil {
			t.Error(err)
		} else {
			t.Logf("cfb str -> %s", string(d))
		}
	}
	key = "1234567890123456"
	if d, err := AesEncryptCBC([]byte(text), []byte(key)); err != nil {
		t.Error(err)
	} else {
		t.Logf("cbc cipher -> %s", string(d))
		if d, err := AesDecryptCBC(d, []byte(key)); err != nil {
			t.Error(err)
		} else {
			t.Logf("cbc str -> %s", string(d))
		}
	}
	if d, err := AesEncryptECB([]byte(text), []byte(key)); err != nil {
		t.Error(err)
	} else {
		t.Logf("ecb cipher -> %s", string(d))
		if d, err := AesDecryptECB(d, []byte(key)); err != nil {
			t.Error(err)
		} else {
			t.Logf("ecb str -> %s", string(d))
		}
	}
	if d, err := AesEncryptCFB([]byte(text), []byte(key)); err != nil {
		t.Error(err)
	} else {
		t.Logf("cfb cipher -> %s", string(d))
		if d, err := AesDecryptCFB(d, []byte(key)); err != nil {
			t.Error(err)
		} else {
			t.Logf("cfb str -> %s", string(d))
		}
	}
}
