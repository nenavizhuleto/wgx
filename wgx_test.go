package wgx_test

import (
	"testing"

	"github.com/nenavizhuleto/wgx"
)

const (
	PrivateKeyExample wgx.PrivateKey = "UMigz7Ll3bbACMrbURD9DaBHEfuQ1pstdb6i6Dh2qEM="
	PublicKeyExample  wgx.PublicKey  = "VhYoer3IzIpKmtJ5TPWKXCnJDHpLYjidhJaefVo/DRs="
)

func Test_GenKeyNoErr(t *testing.T) {
	_, err := wgx.GenKey()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}
}

func Test_PubKeyNoErr(t *testing.T) {
	key, err := wgx.PubKey(PrivateKeyExample)
	if err != nil {
		t.Fatalf("failed to generate public key: %v", err)
	}

	if key != PublicKeyExample {
		t.Fatalf("invalid public key generated: expected %s, got %s", PublicKeyExample, key)
	}
}

func Test_GenerateKeyPairNoErr(t *testing.T) {
	_, err := wgx.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
}
