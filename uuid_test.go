package quuid

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestUUIDVersionsAndVariants(t *testing.T) {
	constructors := []struct {
		name    string
		version Version
		new     func() (UUID, error)
	}{
		{"v4", Version4, NewUUIDv4},
		{"v6", Version6, NewUUIDv6},
		{"v7", Version7, NewUUIDv7},
		{"v8", Version8, NewUUIDv8},
	}
	for _, tc := range constructors {
		t.Run(tc.name, func(t *testing.T) {
			id, err := tc.new()
			if err != nil {
				t.Fatal(err)
			}
			if id.Version() != tc.version {
				t.Fatalf("got version %d, want %d", id.Version(), tc.version)
			}
			if id.Variant() != VariantRFC9562 {
				t.Fatalf("got variant %d", id.Variant())
			}
			if err := ValidateRFC9562UUID(id.String()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUUIDv6RFC9562Example(t *testing.T) {
	timestamp := time.Date(2022, 2, 22, 19, 22, 22, 0, time.UTC)
	tail := []byte{0xb3, 0xc8, 0x9e, 0x6b, 0xde, 0xce, 0xd8, 0x46}
	id, err := NewUUIDv6At(timestamp, bytes.NewReader(tail))
	if err != nil {
		t.Fatal(err)
	}
	const want = "1ec9414c-232a-6b00-b3c8-9e6bdeced846"
	if id.String() != want {
		t.Fatalf("got %s, want %s", id, want)
	}
	decoded, err := id.Time()
	if err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(timestamp) {
		t.Fatalf("got %s, want %s", decoded, timestamp)
	}
}

func TestUUIDv6InjectedClockAndMonotonicity(t *testing.T) {
	fixed := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	g, err := NewGeneratorWith(
		bytes.NewReader(bytes.Repeat([]byte{0x33}, 8)),
		func() time.Time { return fixed },
	)
	if err != nil {
		t.Fatal(err)
	}
	first, err := g.NewUUIDv6()
	if err != nil {
		t.Fatal(err)
	}
	second, err := g.NewUUIDv6()
	if err != nil {
		t.Fatal(err)
	}
	if first.Compare(second) >= 0 {
		t.Fatalf("UUIDv6 values not monotonic: %s >= %s", first, second)
	}
	firstTime, err := first.Time()
	if err != nil {
		t.Fatal(err)
	}
	secondTime, err := second.Time()
	if err != nil {
		t.Fatal(err)
	}
	if !secondTime.After(firstTime) || secondTime.Sub(firstTime) != 100*time.Nanosecond {
		t.Fatalf("unexpected monotonic step: first=%s second=%s", firstTime, secondTime)
	}
}

func TestUUIDv6PreUnixRoundTrip(t *testing.T) {
	want := time.Date(1900, 1, 2, 3, 4, 5, 600, time.UTC)
	id, err := NewUUIDv6At(want, bytes.NewReader(bytes.Repeat([]byte{0x44}, 8)))
	if err != nil {
		t.Fatal(err)
	}
	got, err := id.Time()
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestUUIDv7InjectedClockAndMonotonicity(t *testing.T) {
	fixed := time.UnixMilli(1_700_000_000_123)
	g, err := NewGeneratorWith(
		bytes.NewReader(bytes.Repeat([]byte{0x11}, 20)),
		func() time.Time { return fixed },
	)
	if err != nil {
		t.Fatal(err)
	}
	first, err := g.NewUUIDv7()
	if err != nil {
		t.Fatal(err)
	}
	second, err := g.NewUUIDv7()
	if err != nil {
		t.Fatal(err)
	}
	if first.Compare(second) >= 0 {
		t.Fatalf("UUIDv7 values not monotonic: %s >= %s", first, second)
	}
	for _, id := range []UUID{first, second} {
		got, err := id.Time()
		if err != nil {
			t.Fatal(err)
		}
		if !got.Equal(fixed.UTC()) {
			t.Fatalf("got timestamp %s, want %s", got, fixed.UTC())
		}
	}
}

func TestUUIDStrictAndLooseParsing(t *testing.T) {
	canonical := "a3bb189e-8bf9-4888-9912-ace4e6543002"
	id, err := ParseUUID(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if id.String() != canonical {
		t.Fatalf("got %s", id)
	}

	invalid := []string{
		" " + canonical,
		canonical + " ",
		"{" + canonical + "}",
		"urn:uuid:" + canonical,
		strings.ReplaceAll(canonical, "-", ""),
		"a3bb189e-8bf9-4888-9912-ace4e654300z",
	}
	for _, input := range invalid {
		if _, err := ParseUUID(input); err == nil {
			t.Fatalf("strict parser accepted %q", input)
		}
	}

	loose := []string{
		" " + canonical + " ",
		"{" + canonical + "}",
		"urn:uuid:" + canonical,
		strings.ReplaceAll(canonical, "-", ""),
	}
	for _, input := range loose {
		parsed, err := ParseUUIDLoose(input)
		if err != nil {
			t.Fatalf("loose parser rejected %q: %v", input, err)
		}
		if parsed != id {
			t.Fatalf("loose parse mismatch for %q", input)
		}
	}
}

func TestUUIDSQLTextAndBinaryRepresentations(t *testing.T) {
	id := MustParseUUID("018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22")

	textValue, err := id.Value()
	if err != nil {
		t.Fatal(err)
	}
	if value, ok := textValue.(string); !ok || value != id.String() {
		t.Fatalf("unexpected text value %#v", textValue)
	}

	binaryValue, err := id.Binary().Value()
	if err != nil {
		t.Fatal(err)
	}
	raw, ok := binaryValue.([]byte)
	if !ok || len(raw) != UUIDSize || !bytes.Equal(raw, id[:]) {
		t.Fatalf("unexpected binary value %#v", binaryValue)
	}

	var fromText UUID
	if err := fromText.Scan(id.String()); err != nil || fromText != id {
		t.Fatalf("text scan: %v", err)
	}
	var fromBinary BinaryUUID
	if err := fromBinary.Scan(raw); err != nil || fromBinary.UUID() != id {
		t.Fatalf("binary scan: %v", err)
	}

	var _ driver.Valuer = id
	var _ driver.Valuer = id.Binary()
}

func TestNullUUIDEmptyValuesAreInvalid(t *testing.T) {
	for _, source := range []any{nil, "", []byte{}} {
		n := NullUUID{UUID: MaxUUID, Valid: true}
		if err := n.Scan(source); err != nil {
			t.Fatalf("scan %#v: %v", source, err)
		}
		if n.Valid || !n.UUID.IsZero() {
			t.Fatalf("source %#v produced valid value", source)
		}
	}

	id := MustNewUUIDv4()
	n := ToNullUUID(id)
	if !n.Valid || n.UUID != id {
		t.Fatal("ToNullUUID lost a valid UUID")
	}
	if ToNullUUID(NilUUID).Valid {
		t.Fatal("NilUUID must map to invalid ToNullUUID")
	}

	data, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip NullUUID
	if err := json.Unmarshal(data, &roundTrip); err != nil || !roundTrip.Valid || roundTrip.UUID != id {
		t.Fatalf("JSON round trip: %v", err)
	}
	if err := json.Unmarshal([]byte(`""`), &roundTrip); err != nil || roundTrip.Valid {
		t.Fatalf("empty JSON string must be invalid: %v", err)
	}
}

func TestUUIDAppendMethods(t *testing.T) {
	id := MustParseUUID("018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22")
	text, err := id.AppendText([]byte("prefix:"))
	if err != nil {
		t.Fatal(err)
	}
	if string(text) != "prefix:"+id.String() {
		t.Fatalf("unexpected text %q", text)
	}
	binary, err := id.AppendBinary([]byte{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(binary) != 2+UUIDSize || !bytes.Equal(binary[2:], id[:]) {
		t.Fatal("unexpected binary append")
	}
}

func TestUUIDv8UsesModernDerivation(t *testing.T) {
	a := DeriveUUIDv8("invoice", []byte("INV-42"))
	b := DeriveUUIDv8("invoice", []byte("INV-42"))
	c := DeriveUUIDv8("customer", []byte("INV-42"))
	if a != b || a == c {
		t.Fatal("deterministic UUIDv8 domain separation failed")
	}
	if a.Version() != Version8 || a.Variant() != VariantRFC9562 {
		t.Fatal("invalid UUIDv8 fields")
	}
	if _, err := DeriveUUIDv8Keyed([]byte("short"), "x", nil); !errors.Is(err, ErrWeakSecret) {
		t.Fatalf("expected ErrWeakSecret, got %v", err)
	}
}

func TestUUIDRejectsNilAndInvalidSources(t *testing.T) {
	var id UUID
	if err := id.Scan(nil); !errors.Is(err, ErrNilValue) {
		t.Fatalf("got %v", err)
	}
	if err := id.Scan(123); !errors.Is(err, ErrInvalidSource) {
		t.Fatalf("got %v", err)
	}
	if err := json.Unmarshal([]byte("null"), &id); !errors.Is(err, ErrNilValue) {
		t.Fatalf("got %v", err)
	}
}
