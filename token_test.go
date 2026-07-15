package quuid

import (
	"bytes"
	"testing"
)

func TestTokenVerification(t *testing.T) {
	token, err := NewTokenFromReader(bytes.NewReader(bytes.Repeat([]byte{0x5a}, TokenSize)))
	if err != nil {
		t.Fatal(err)
	}
	digest := token.Digest()
	if !VerifyToken(token.Secret(), digest) {
		t.Fatal("valid token rejected")
	}
	other, err := NewTokenFromReader(bytes.NewReader(bytes.Repeat([]byte{0x5b}, TokenSize)))
	if err != nil {
		t.Fatal(err)
	}
	if VerifyToken(other.Secret(), digest) {
		t.Fatal("invalid token accepted")
	}
	if token.Redacted() == token.Secret() {
		t.Fatal("redaction exposed full token")
	}
}
