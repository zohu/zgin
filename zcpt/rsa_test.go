package zcpt

import "testing"

func TestRsa(t *testing.T) {
	v, b, err := GenerateRSAKey(2048)
	if err != nil {
		t.Error(err)
	}
	piv := EncodePrivateRSAKeyToPEM(v)
	pub, _ := EncodePublicRSAKeyToPEM(b)
	t.Logf("pub -> %s", string(pub))
	t.Logf("piv -> %s", string(piv))

	text := "hello world"
	cipher, err := RSAEncrypt([]byte(text), pub)
	if err != nil {
		t.Error(err)
	}
	t.Logf("cipher -> %s", string(cipher))
	str, err := RSADecrypt(cipher, piv)
	if err != nil {
		t.Error(err)
	}
	t.Logf("str -> %s", string(str))
}
