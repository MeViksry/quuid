package quuid

import (
	"bytes"
	"testing"
)

func TestStrongIDRoundTrip(t *testing.T) {
	id, err := NewStrongIDFromReader(bytes.NewReader(bytes.Repeat([]byte{0x7a}, StrongIDSize)))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseStrongID(id.String())
	if err != nil {
		t.Fatal(err)
	}
	if !id.Equal(parsed) {
		t.Fatal("round trip mismatch")
	}
	if len(id.String()) != len(strongIDPrefix)+77 {
		t.Fatalf("unexpected encoded length %d", len(id.String()))
	}
}

func TestStrongIDDerivation(t *testing.T) {
	a := DeriveStrongID("tenant", []byte("alpha"))
	b := DeriveStrongID("tenant", []byte("alpha"))
	if !a.Equal(b) {
		t.Fatal("deterministic derivation mismatch")
	}
	secret := bytes.Repeat([]byte{0x11}, 32)
	keyed, err := DeriveStrongIDKeyed(secret, "tenant", []byte("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if keyed.Equal(a) {
		t.Fatal("keyed and unkeyed derivations must differ")
	}
}
