package quuid

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestIDRoundTrip(t *testing.T) {
	raw := bytes.Repeat([]byte{0x42}, IDSize)
	id, err := NewIDFromReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(id.String(), idPrefix) {
		t.Fatalf("missing prefix: %s", id)
	}
	parsed, err := ParseID(id.String())
	if err != nil {
		t.Fatal(err)
	}
	if !id.Equal(parsed) {
		t.Fatalf("round trip mismatch")
	}

	fromHex, err := ParseID(id.Hex())
	if err != nil {
		t.Fatal(err)
	}
	if !id.Equal(fromHex) {
		t.Fatalf("hex round trip mismatch")
	}
}

func TestIDCrockfordAliases(t *testing.T) {
	var id ID
	for i := range id {
		id[i] = byte(i + 1)
	}
	canonical := id.String()
	upper := strings.ToUpper(canonical)
	parsed, err := ParseID(upper)
	if err != nil {
		t.Fatal(err)
	}
	if !id.Equal(parsed) {
		t.Fatal("uppercase parsing mismatch")
	}
}

func TestIDJSONAndSQL(t *testing.T) {
	id := DeriveID("user", []byte("42"))
	data, err := json.Marshal(id)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !id.Equal(decoded) {
		t.Fatal("JSON mismatch")
	}

	value, err := id.Value()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := value.(string); !ok {
		t.Fatalf("unexpected driver value %T", value)
	}
	var scanned ID
	if err := scanned.Scan(value); err != nil {
		t.Fatal(err)
	}
	if !id.Equal(scanned) {
		t.Fatal("SQL scan mismatch")
	}

	var _ driver.Valuer = id
}

func TestNullID(t *testing.T) {
	var nullable NullID
	if err := nullable.Scan(nil); err != nil {
		t.Fatal(err)
	}
	if nullable.Valid {
		t.Fatal("expected invalid NullID")
	}
	data, err := json.Marshal(nullable)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "null" {
		t.Fatalf("got %s", data)
	}

	id := DeriveID("x", []byte("y"))
	if err := nullable.Scan(id.String()); err != nil {
		t.Fatal(err)
	}
	if !nullable.Valid || !nullable.ID.Equal(id) {
		t.Fatal("expected valid NullID")
	}
}

func TestWeakSecretRejected(t *testing.T) {
	_, err := DeriveIDKeyed([]byte("too short"), "ns", []byte("data"))
	if !errors.Is(err, ErrWeakSecret) {
		t.Fatalf("got %v", err)
	}
}

func TestZeroEntropyRejected(t *testing.T) {
	_, err := NewIDFromReader(bytes.NewReader(make([]byte, IDSize)))
	if !errors.Is(err, ErrZeroEntropy) {
		t.Fatalf("got %v", err)
	}
}

func TestDeterministicID(t *testing.T) {
	a := DeriveID("orders", []byte("INV-42"))
	b := DeriveID("orders", []byte("INV-42"))
	c := DeriveID("users", []byte("INV-42"))
	if !a.Equal(b) {
		t.Fatal("same input must derive same ID")
	}
	if a.Equal(c) {
		t.Fatal("domain separation failed")
	}
}
