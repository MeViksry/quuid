package quuid

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

var (
	_ encoding.TextMarshaler     = UUID{}
	_ encoding.TextUnmarshaler   = (*UUID)(nil)
	_ encoding.BinaryMarshaler   = UUID{}
	_ encoding.BinaryUnmarshaler = (*UUID)(nil)
	_ json.Marshaler             = UUID{}
	_ json.Unmarshaler           = (*UUID)(nil)
	_ sql.Scanner                = (*UUID)(nil)
	_ driver.Valuer              = UUID{}

	_ encoding.TextMarshaler     = BinaryUUID{}
	_ encoding.TextUnmarshaler   = (*BinaryUUID)(nil)
	_ encoding.BinaryMarshaler   = BinaryUUID{}
	_ encoding.BinaryUnmarshaler = (*BinaryUUID)(nil)
	_ json.Marshaler             = BinaryUUID{}
	_ json.Unmarshaler           = (*BinaryUUID)(nil)
	_ sql.Scanner                = (*BinaryUUID)(nil)
	_ driver.Valuer              = BinaryUUID{}

	_ json.Marshaler   = NullUUID{}
	_ json.Unmarshaler = (*NullUUID)(nil)
	_ sql.Scanner      = (*NullUUID)(nil)
	_ driver.Valuer    = NullUUID{}

	_ json.Marshaler   = NullBinaryUUID{}
	_ json.Unmarshaler = (*NullBinaryUUID)(nil)
	_ sql.Scanner      = (*NullBinaryUUID)(nil)
	_ driver.Valuer    = NullBinaryUUID{}
)

func TestUUIDEncodingInterfaces(t *testing.T) {
	id := MustParseUUID("018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22")
	if id.IsZero() || !id.Equal(id) || id.Compare(id) != 0 {
		t.Fatal("UUID comparison helpers failed")
	}
	if len(id.Bytes()) != UUIDSize {
		t.Fatal("UUID byte length mismatch")
	}

	text, err := id.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	var fromText UUID
	if err := fromText.UnmarshalText(text); err != nil || fromText != id {
		t.Fatalf("text round trip: %v", err)
	}

	binary, err := id.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	var fromBinary UUID
	if err := fromBinary.UnmarshalBinary(binary); err != nil || fromBinary != id {
		t.Fatalf("binary round trip: %v", err)
	}
	if err := fromBinary.UnmarshalBinary(binary[:1]); !errors.Is(err, ErrInvalidLength) {
		t.Fatalf("expected length error, got %v", err)
	}

	encoded, err := json.Marshal(id)
	if err != nil {
		t.Fatal(err)
	}
	var fromJSON UUID
	if err := json.Unmarshal(encoded, &fromJSON); err != nil || fromJSON != id {
		t.Fatalf("JSON round trip: %v", err)
	}
	if err := json.Unmarshal([]byte(`123`), &fromJSON); err == nil {
		t.Fatal("expected JSON type error")
	}

	var nilUUID *UUID
	if err := nilUUID.UnmarshalText(text); err == nil {
		t.Fatal("expected nil receiver error")
	}
	if err := nilUUID.UnmarshalBinary(binary); err == nil {
		t.Fatal("expected nil receiver error")
	}
	if err := nilUUID.UnmarshalJSON(encoded); err == nil {
		t.Fatal("expected nil receiver error")
	}
}

func TestBinaryUUIDInterfaces(t *testing.T) {
	logical := MustParseUUID("018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22")
	binaryID := logical.Binary()
	if binaryID.String() != logical.String() || binaryID.IsZero() {
		t.Fatal("BinaryUUID helpers failed")
	}
	if binaryID.Version() != logical.Version() || binaryID.Variant() != logical.Variant() {
		t.Fatal("BinaryUUID metadata mismatch")
	}
	if !binaryID.Equal(binaryID) || binaryID.Compare(binaryID) != 0 {
		t.Fatal("BinaryUUID comparisons failed")
	}
	if !bytes.Equal(binaryID.Bytes(), logical[:]) {
		t.Fatal("BinaryUUID bytes mismatch")
	}

	text, err := binaryID.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	appendedText, err := binaryID.AppendText([]byte("x"))
	if err != nil || string(appendedText) != "x"+logical.String() {
		t.Fatalf("AppendText: %v", err)
	}
	var fromText BinaryUUID
	if err := fromText.UnmarshalText(text); err != nil || fromText != binaryID {
		t.Fatalf("text round trip: %v", err)
	}

	raw, err := binaryID.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	appendedBinary, err := binaryID.AppendBinary([]byte{1})
	if err != nil || len(appendedBinary) != 1+UUIDSize {
		t.Fatalf("AppendBinary: %v", err)
	}
	var fromBinary BinaryUUID
	if err := fromBinary.UnmarshalBinary(raw); err != nil || fromBinary != binaryID {
		t.Fatalf("binary round trip: %v", err)
	}
	if err := fromBinary.UnmarshalBinary(raw[:1]); !errors.Is(err, ErrInvalidLength) {
		t.Fatalf("expected length error, got %v", err)
	}

	encoded, err := binaryID.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var fromJSON BinaryUUID
	if err := fromJSON.UnmarshalJSON(encoded); err != nil || fromJSON != binaryID {
		t.Fatalf("JSON round trip: %v", err)
	}
	if err := fromJSON.UnmarshalJSON([]byte("null")); !errors.Is(err, ErrNilValue) {
		t.Fatalf("expected null error, got %v", err)
	}

	var nilBinary *BinaryUUID
	if err := nilBinary.UnmarshalText(text); err == nil {
		t.Fatal("expected nil receiver error")
	}
	if err := nilBinary.UnmarshalBinary(raw); err == nil {
		t.Fatal("expected nil receiver error")
	}
	if err := nilBinary.UnmarshalJSON(encoded); err == nil {
		t.Fatal("expected nil receiver error")
	}
	if err := nilBinary.Scan(raw); err == nil {
		t.Fatal("expected nil receiver error")
	}
}

func TestNullBinaryUUIDInterfaces(t *testing.T) {
	id := MustParseUUID("018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22")
	n := NewNullBinaryUUID(id)
	if !n.Valid || n.UUID != id {
		t.Fatal("NewNullBinaryUUID failed")
	}
	if ToNullBinaryUUID(NilUUID).Valid {
		t.Fatal("Nil UUID must produce invalid nullable binary UUID")
	}

	value, err := n.Value()
	if err != nil {
		t.Fatal(err)
	}
	raw, ok := value.([]byte)
	if !ok || len(raw) != UUIDSize {
		t.Fatalf("unexpected value %#v", value)
	}

	var scanned NullBinaryUUID
	if err := scanned.Scan(raw); err != nil || !scanned.Valid || scanned.UUID != id {
		t.Fatalf("binary scan: %v", err)
	}
	if err := scanned.Scan(id.String()); err != nil || !scanned.Valid || scanned.UUID != id {
		t.Fatalf("text scan: %v", err)
	}
	if err := scanned.Scan(nil); err != nil || scanned.Valid {
		t.Fatalf("null scan: %v", err)
	}
	if value, err := scanned.Value(); err != nil || value != nil {
		t.Fatalf("invalid value: %#v, %v", value, err)
	}

	encoded, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &scanned); err != nil || !scanned.Valid || scanned.UUID != id {
		t.Fatalf("JSON round trip: %v", err)
	}
	if err := json.Unmarshal([]byte("null"), &scanned); err != nil || scanned.Valid {
		t.Fatalf("JSON null: %v", err)
	}
	if err := json.Unmarshal([]byte(`"invalid"`), &scanned); err == nil || scanned.Valid {
		t.Fatal("malformed JSON must fail closed")
	}

	var nilNullable *NullBinaryUUID
	if err := nilNullable.Scan(nil); err == nil {
		t.Fatal("expected nil receiver error")
	}
	if err := nilNullable.UnmarshalJSON([]byte("null")); err == nil {
		t.Fatal("expected nil receiver error")
	}
}

func TestGeneratorUUIDConvenienceMethods(t *testing.T) {
	reader := bytes.NewReader(bytes.Repeat([]byte{0x42}, 64))
	clock := func() time.Time { return time.UnixMilli(1_700_000_000_000) }
	g, err := NewGeneratorWith(reader, clock)
	if err != nil {
		t.Fatal(err)
	}

	v4, err := g.NewUUIDv4()
	if err != nil || v4.Version() != Version4 {
		t.Fatalf("v4: %v", err)
	}
	v6, err := g.NewUUIDv6()
	if err != nil || v6.Version() != Version6 {
		t.Fatalf("v6: %v", err)
	}
	v7, err := g.NewUUIDv7()
	if err != nil || v7.Version() != Version7 {
		t.Fatalf("v7: %v", err)
	}
	v8, err := g.NewUUIDv8()
	if err != nil || v8.Version() != Version8 {
		t.Fatalf("v8: %v", err)
	}

	var nilGenerator *Generator
	if _, err := nilGenerator.NewUUIDv4(); err == nil {
		t.Fatal("expected nil generator error")
	}
	if _, err := nilGenerator.NewUUIDv6(); err == nil {
		t.Fatal("expected nil generator error")
	}
	if _, err := nilGenerator.NewUUIDv7(); err == nil {
		t.Fatal("expected nil generator error")
	}
	if _, err := nilGenerator.NewUUIDv8(); err == nil {
		t.Fatal("expected nil generator error")
	}
}

func TestUUIDVersionAndVariantStrings(t *testing.T) {
	if Version7.String() != "UUIDv7" || VersionUnknown.String() != "unknown" {
		t.Fatal("unexpected Version string")
	}
	if VariantRFC9562.String() != "RFC9562" || Variant(99).String() != "unknown" {
		t.Fatal("unexpected Variant string")
	}
}

func TestUUIDVariantAndValidationBranches(t *testing.T) {
	variants := []struct {
		value byte
		want  Variant
	}{
		{0x00, VariantNCS},
		{0x80, VariantRFC9562},
		{0xc0, VariantMicrosoft},
		{0xe0, VariantFuture},
	}
	for _, tc := range variants {
		id := NilUUID
		id[8] = tc.value
		if got := id.Variant(); got != tc.want {
			t.Fatalf("variant byte %#x: got %d, want %d", tc.value, got, tc.want)
		}
	}

	invalidVariant := MustParseUUID("018fbd2e-7b46-7cc0-18c4-89e6f6dc0c22")
	if err := ValidateRFC9562UUID(invalidVariant.String()); err == nil {
		t.Fatal("expected variant validation error")
	}
	invalidVersion := MustParseUUID("018fbd2e-7b46-fcc0-98c4-89e6f6dc0c22")
	if err := ValidateRFC9562UUID(invalidVersion.String()); err == nil {
		t.Fatal("expected version validation error")
	}
}

func TestUUIDTimeAndRangeErrors(t *testing.T) {
	if _, err := NilUUID.Time(); err == nil {
		t.Fatal("expected unsupported timestamp error")
	}
	beforeEpoch := time.Date(1500, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := NewUUIDv6At(beforeEpoch, bytes.NewReader(bytes.Repeat([]byte{1}, 8))); !errors.Is(err, ErrTimeOutOfRange) {
		t.Fatalf("expected v6 time range error, got %v", err)
	}
	if _, err := NewUUIDv7At(time.UnixMilli(-1), bytes.NewReader(bytes.Repeat([]byte{1}, 10))); !errors.Is(err, ErrTimeOutOfRange) {
		t.Fatalf("expected v7 time range error, got %v", err)
	}
}

func TestMustUUIDConstructors(t *testing.T) {
	if MustNewUUIDv4().Version() != Version4 {
		t.Fatal("MustNewUUIDv4 failed")
	}
	if MustNewUUIDv6().Version() != Version6 {
		t.Fatal("MustNewUUIDv6 failed")
	}
	if MustNewUUIDv7().Version() != Version7 {
		t.Fatal("MustNewUUIDv7 failed")
	}
	if MustNewUUIDv8().Version() != Version8 {
		t.Fatal("MustNewUUIDv8 failed")
	}
}
