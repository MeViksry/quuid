package quuid

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
)

// StrongIDSize is the binary size of a StrongID in bytes.
const (
	StrongIDSize   = 48
	strongIDPrefix = "qsid_"
)

// StrongID is a 384-bit identifier for systems that want a larger security margin.
// Under generic idealized quantum collision search, 384 output bits target roughly
// 128 bits of collision resistance. It is longer than a UUID and is not RFC 9562 compatible.
type StrongID [StrongIDSize]byte

// NewStrongID returns a new strong identifier using the generator entropy source.
func NewStrongID() (StrongID, error) {
	return NewStrongIDFromReader(rand.Reader)
}

// NewStrongIDFromReader returns a new StrongID using r as its entropy source.
func NewStrongIDFromReader(r io.Reader) (StrongID, error) {
	var id StrongID
	if r == nil {
		return id, fmt.Errorf("quuid: nil entropy reader")
	}
	if _, err := io.ReadFull(r, id[:]); err != nil {
		return StrongID{}, fmt.Errorf("quuid: read entropy: %w", err)
	}
	if allZero(id[:]) {
		return StrongID{}, ErrZeroEntropy
	}
	return id, nil
}

// MustNewStrongID is like NewStrongID but panics when secure randomness fails.
func MustNewStrongID() StrongID {
	id, err := NewStrongID()
	if err != nil {
		panic(err)
	}
	return id
}

// DeriveStrongID deterministically derives a StrongID using SHA-384 and domain separation.
func DeriveStrongID(namespace string, data []byte) StrongID {
	h := sha512.New384()
	writeFrame(h, []byte("github.com/MeViksry/quuid/derive-strong-id/v1"), []byte(namespace), data)
	var id StrongID
	copy(id[:], h.Sum(nil))
	return id
}

// DeriveStrongIDKeyed deterministically derives a StrongID using HMAC-SHA-512.
func DeriveStrongIDKeyed(secret []byte, namespace string, data []byte) (StrongID, error) {
	if len(secret) < 32 {
		return StrongID{}, ErrWeakSecret
	}
	h := hmac.New(sha512.New, secret)
	writeFrame(h, []byte("github.com/MeViksry/quuid/derive-strong-id-keyed/v1"), []byte(namespace), data)
	var id StrongID
	copy(id[:], h.Sum(nil)[:StrongIDSize])
	return id, nil
}

// ParseStrongID parses a StrongID from its supported text representations.
func ParseStrongID(s string) (StrongID, error) {
	raw, err := decodeText(s, strongIDPrefix, StrongIDSize)
	if err != nil {
		return StrongID{}, err
	}
	var id StrongID
	copy(id[:], raw)
	return id, nil
}

// MustParseStrongID is like ParseStrongID but panics on malformed input.
func MustParseStrongID(s string) StrongID {
	id, err := ParseStrongID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// String returns the canonical text representation.
func (id StrongID) String() string { return encodeText(strongIDPrefix, id[:]) }

// Hex returns the lowercase hexadecimal representation without a type prefix.
func (id StrongID) Hex() string { return fmt.Sprintf("%x", id[:]) }

// Bytes returns a defensive copy of the binary representation.
func (id StrongID) Bytes() []byte { return append([]byte(nil), id[:]...) }

// IsZero reports whether every byte is zero.
func (id StrongID) IsZero() bool { return allZero(id[:]) }

// Equal reports whether two values contain identical bytes.
func (id StrongID) Equal(other StrongID) bool {
	return subtle.ConstantTimeCompare(id[:], other[:]) == 1
}

// Compare compares two values lexicographically and returns -1, 0, or 1.
func (id StrongID) Compare(other StrongID) int {
	for i := range id {
		if id[i] < other[i] {
			return -1
		}
		if id[i] > other[i] {
			return 1
		}
	}
	return 0
}

// MarshalText implements encoding.TextMarshaler.
func (id StrongID) MarshalText() ([]byte, error) { return id.AppendText(nil) }

// AppendText appends the canonical text representation to dst.
func (id StrongID) AppendText(dst []byte) ([]byte, error) {
	return appendEncodedText(dst, strongIDPrefix, id[:]), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (id *StrongID) UnmarshalText(text []byte) error {
	if id == nil {
		return fmt.Errorf("quuid: UnmarshalText on nil *StrongID")
	}
	parsed, err := ParseStrongID(string(text))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// MarshalJSON implements json.Marshaler.
func (id StrongID) MarshalJSON() ([]byte, error) { return json.Marshal(id.String()) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *StrongID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return ErrNilValue
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("quuid: decode StrongID JSON: %w", err)
	}
	return id.UnmarshalText([]byte(value))
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (id StrongID) MarshalBinary() ([]byte, error) { return id.AppendBinary(nil) }

// AppendBinary appends the binary representation to dst.
func (id StrongID) AppendBinary(dst []byte) ([]byte, error) {
	return append(dst, id[:]...), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (id *StrongID) UnmarshalBinary(data []byte) error {
	if id == nil {
		return fmt.Errorf("quuid: UnmarshalBinary on nil *StrongID")
	}
	if len(data) != StrongIDSize {
		return ErrInvalidLength
	}
	copy(id[:], data)
	return nil
}

// Scan implements database/sql.Scanner.
func (id *StrongID) Scan(src any) error {
	if id == nil {
		return fmt.Errorf("quuid: Scan on nil *StrongID")
	}
	var parsed StrongID
	switch value := src.(type) {
	case nil:
		return ErrNilValue
	case string:
		var err error
		parsed, err = ParseStrongID(value)
		if err != nil {
			return err
		}
	case []byte:
		if len(value) == StrongIDSize {
			copy(parsed[:], value)
		} else {
			var err error
			parsed, err = ParseStrongID(string(value))
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("%w: %T", ErrInvalidSource, src)
	}
	*id = parsed
	return nil
}

// Value implements database/sql/driver.Valuer.
func (id StrongID) Value() (driver.Value, error) { return id.String(), nil }
