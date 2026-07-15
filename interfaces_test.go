package quuid

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestIDInterfacesAndOrdering(t *testing.T) {
	randomID, err := NewID()
	if err != nil || randomID.IsZero() {
		t.Fatalf("NewID: %v, zero=%v", err, randomID.IsZero())
	}
	if MustNewID().IsZero() {
		t.Fatal("MustNewID returned zero")
	}

	var low, high ID
	low[IDSize-1] = 1
	high[IDSize-1] = 2
	if low.Compare(high) != -1 || high.Compare(low) != 1 || low.Compare(low) != 0 {
		t.Fatal("ID Compare failed")
	}
	if len(low.Bytes()) != IDSize {
		t.Fatal("ID Bytes length")
	}

	text, err := low.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	var fromText ID
	if err := fromText.UnmarshalText(text); err != nil || !fromText.Equal(low) {
		t.Fatalf("ID text round trip: %v", err)
	}

	binary, err := low.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	var fromBinary ID
	if err := fromBinary.UnmarshalBinary(binary); err != nil || !fromBinary.Equal(low) {
		t.Fatalf("ID binary round trip: %v", err)
	}
	if err := fromBinary.UnmarshalBinary(binary[:len(binary)-1]); !errors.Is(err, ErrInvalidLength) {
		t.Fatalf("expected invalid length, got %v", err)
	}

	if _, err := ParseID("wrong_prefix_value"); !errors.Is(err, ErrInvalidPrefix) {
		t.Fatalf("expected prefix error, got %v", err)
	}
	if _, err := ParseID("qid_short"); !errors.Is(err, ErrInvalidLength) {
		t.Fatalf("expected length error, got %v", err)
	}
	if err := json.Unmarshal([]byte("null"), &fromText); !errors.Is(err, ErrNilValue) {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := fromText.Scan(123); !errors.Is(err, ErrInvalidSource) {
		t.Fatalf("expected source error, got %v", err)
	}
	if err := fromText.Scan(nil); !errors.Is(err, ErrNilValue) {
		t.Fatalf("expected nil scan error, got %v", err)
	}
}

func TestStrongIDInterfaces(t *testing.T) {
	id, err := NewStrongID()
	if err != nil || id.IsZero() {
		t.Fatalf("NewStrongID: %v", err)
	}
	if MustNewStrongID().IsZero() {
		t.Fatal("MustNewStrongID returned zero")
	}
	if MustParseStrongID(id.String()) != id {
		t.Fatal("MustParseStrongID mismatch")
	}
	if len(id.Bytes()) != StrongIDSize || len(id.Hex()) != StrongIDSize*2 {
		t.Fatal("StrongID encoding lengths")
	}

	var other StrongID
	other[StrongIDSize-1] = 1
	if (StrongID{}).Compare(other) != -1 || other.Compare(StrongID{}) != 1 || other.Compare(other) != 0 {
		t.Fatal("StrongID Compare failed")
	}

	text, _ := id.MarshalText()
	var fromText StrongID
	if err := fromText.UnmarshalText(text); err != nil || !fromText.Equal(id) {
		t.Fatalf("text: %v", err)
	}
	data, _ := json.Marshal(id)
	var fromJSON StrongID
	if err := json.Unmarshal(data, &fromJSON); err != nil || !fromJSON.Equal(id) {
		t.Fatalf("json: %v", err)
	}
	if err := json.Unmarshal([]byte("null"), &fromJSON); !errors.Is(err, ErrNilValue) {
		t.Fatalf("expected nil error, got %v", err)
	}
	binary, _ := id.MarshalBinary()
	var fromBinary StrongID
	if err := fromBinary.UnmarshalBinary(binary); err != nil || !fromBinary.Equal(id) {
		t.Fatalf("binary: %v", err)
	}
	if err := fromBinary.UnmarshalBinary(binary[:1]); !errors.Is(err, ErrInvalidLength) {
		t.Fatalf("expected length error, got %v", err)
	}
	if err := fromBinary.Scan(id.String()); err != nil || !fromBinary.Equal(id) {
		t.Fatalf("scan string: %v", err)
	}
	if err := fromBinary.Scan(id.Bytes()); err != nil || !fromBinary.Equal(id) {
		t.Fatalf("scan binary: %v", err)
	}
	if _, err := id.Value(); err != nil {
		t.Fatal(err)
	}
	if err := fromBinary.Scan(nil); !errors.Is(err, ErrNilValue) {
		t.Fatalf("expected nil scan error, got %v", err)
	}
	if err := fromBinary.Scan(1.5); !errors.Is(err, ErrInvalidSource) {
		t.Fatalf("expected source error, got %v", err)
	}
}

func TestSortableIDInterfaces(t *testing.T) {
	id, err := NewSortableID()
	if err != nil || id.IsZero() {
		t.Fatalf("NewSortableID: %v", err)
	}
	if MustNewSortableID().IsZero() {
		t.Fatal("MustNewSortableID returned zero")
	}
	if MustParseSortableID(id.String()) != id {
		t.Fatal("MustParseSortableID mismatch")
	}
	if len(id.Bytes()) != SortableIDSize || len(id.Hex()) != SortableIDSize*2 {
		t.Fatal("SortableID encoding lengths")
	}
	if id.Entropy().IsZero() {
		t.Fatal("entropy unexpectedly zero")
	}

	text, _ := id.MarshalText()
	var fromText SortableID
	if err := fromText.UnmarshalText(text); err != nil || !fromText.Equal(id) {
		t.Fatalf("text: %v", err)
	}
	data, _ := json.Marshal(id)
	var fromJSON SortableID
	if err := json.Unmarshal(data, &fromJSON); err != nil || !fromJSON.Equal(id) {
		t.Fatalf("json: %v", err)
	}
	if err := json.Unmarshal([]byte("null"), &fromJSON); !errors.Is(err, ErrNilValue) {
		t.Fatalf("expected nil error, got %v", err)
	}
	binary, _ := id.MarshalBinary()
	var fromBinary SortableID
	if err := fromBinary.UnmarshalBinary(binary); err != nil || !fromBinary.Equal(id) {
		t.Fatalf("binary: %v", err)
	}
	if err := fromBinary.UnmarshalBinary(binary[:1]); !errors.Is(err, ErrInvalidLength) {
		t.Fatalf("expected length error, got %v", err)
	}
	if err := fromBinary.Scan(id.String()); err != nil || !fromBinary.Equal(id) {
		t.Fatalf("scan string: %v", err)
	}
	if err := fromBinary.Scan(id.Bytes()); err != nil || !fromBinary.Equal(id) {
		t.Fatalf("scan bytes: %v", err)
	}
	if _, err := id.Value(); err != nil {
		t.Fatal(err)
	}
	if err := fromBinary.Scan(nil); !errors.Is(err, ErrNilValue) {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := fromBinary.Scan(struct{}{}); !errors.Is(err, ErrInvalidSource) {
		t.Fatalf("expected source error, got %v", err)
	}

	if _, err := NewSortableIDAt(time.UnixMilli(-1), bytes.NewReader(bytes.Repeat([]byte{1}, IDSize))); !errors.Is(err, ErrTimeOutOfRange) {
		t.Fatalf("expected time error, got %v", err)
	}
}

func TestGeneratorConvenienceMethods(t *testing.T) {
	reader := bytes.NewReader(bytes.Repeat([]byte{0x77}, IDSize+StrongIDSize+IDSize))
	g, err := NewGeneratorWith(reader, func() time.Time { return time.UnixMilli(42) })
	if err != nil {
		t.Fatal(err)
	}
	if id, err := g.NewID(); err != nil || id.IsZero() {
		t.Fatalf("NewID: %v", err)
	}
	if id, err := g.NewStrongID(); err != nil || id.IsZero() {
		t.Fatalf("NewStrongID: %v", err)
	}
	if id, err := g.NewSortableID(); err != nil || id.IsZero() {
		t.Fatalf("NewSortableID: %v", err)
	}
	if _, err := NewGeneratorWith(nil, time.Now); err == nil {
		t.Fatal("expected nil reader error")
	}
	if _, err := NewGeneratorWith(bytes.NewReader(nil), nil); err == nil {
		t.Fatal("expected nil clock error")
	}
}

func TestNullIDJSONAndValue(t *testing.T) {
	var n NullID
	value, err := n.Value()
	if err != nil || value != nil {
		t.Fatalf("invalid NullID value: %v, %v", value, err)
	}
	if err := n.UnmarshalJSON([]byte("null")); err != nil || n.Valid {
		t.Fatalf("null JSON: %v", err)
	}
	id := DeriveID("null", []byte("valid"))
	data, _ := json.Marshal(id.String())
	if err := n.UnmarshalJSON(data); err != nil || !n.Valid || !n.ID.Equal(id) {
		t.Fatalf("valid JSON: %v", err)
	}
	value, err = n.Value()
	if err != nil || value != id.String() {
		t.Fatalf("valid value: %v, %v", value, err)
	}
}

func TestUUIDWrapperCoverage(t *testing.T) {
	if MustNewUUIDv4().Version() != 4 {
		t.Fatal("MustNewUUIDv4")
	}
	v4, err := NewUUIDv4FromReader(bytes.NewReader(bytes.Repeat([]byte{0x12}, 16)))
	if err != nil || v4.Version() != 4 {
		t.Fatalf("NewUUIDv4FromReader: %v", err)
	}
	if MustNewUUIDv7().Version() != 7 {
		t.Fatal("MustNewUUIDv7")
	}
	v7, err := NewUUIDv7FromReader(bytes.NewReader(bytes.Repeat([]byte{0x23}, 16)))
	if err != nil || v7.Version() != 7 {
		t.Fatalf("NewUUIDv7FromReader: %v", err)
	}
	v8, err := NewUUIDv8()
	if err != nil || v8.Version() != 8 {
		t.Fatalf("NewUUIDv8: %v", err)
	}
	if MustNewUUIDv8().Version() != 8 {
		t.Fatal("MustNewUUIDv8")
	}
	keyed, err := DeriveUUIDv8Keyed(bytes.Repeat([]byte{0x99}, 32), "ns", []byte("value"))
	if err != nil || keyed.Version() != 8 {
		t.Fatalf("DeriveUUIDv8Keyed: %v", err)
	}
	if _, err := DeriveUUIDv8Keyed([]byte("short"), "ns", nil); !errors.Is(err, ErrWeakSecret) {
		t.Fatalf("expected weak secret, got %v", err)
	}
	if ValidateUUID(v8.String()) != nil {
		t.Fatal("ValidateUUID rejected valid UUID")
	}
	if MustParseUUID(v8.String()) != v8 {
		t.Fatal("MustParseUUID mismatch")
	}
}

func TestTokenConstructorsAndParsing(t *testing.T) {
	token, err := NewToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(token.Bytes()) != TokenSize {
		t.Fatal("Token Bytes length")
	}
	parsed, err := ParseToken(token.Secret())
	if err != nil || parsed != token {
		t.Fatalf("ParseToken: %v", err)
	}
	if MustNewToken() == (Token{}) {
		t.Fatal("MustNewToken returned zero")
	}
	if VerifyToken("not-a-token", token.Digest()) {
		t.Fatal("invalid token verified")
	}
}
