package quuid

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// SortableIDSize is the binary size of a SortableID in bytes.
const (
	SortableIDSize   = 40
	sortableIDPrefix = "qsort_"
)

// SortableID contains an 8-byte Unix millisecond timestamp followed by
// 32 bytes of cryptographic entropy. Canonical text values sort by timestamp.
// The timestamp is intentionally visible; do not use this type when creation
// time must remain private.
type SortableID [SortableIDSize]byte

// NewSortableID returns a new sortable identifier using the generator clock and entropy source.
func NewSortableID() (SortableID, error) {
	return NewSortableIDAt(time.Now(), rand.Reader)
}

// NewSortableIDAt returns a SortableID for t using r as its entropy source.
func NewSortableIDAt(t time.Time, r io.Reader) (SortableID, error) {
	var id SortableID
	ms := t.UnixMilli()
	if ms < 0 {
		return id, ErrTimeOutOfRange
	}
	if r == nil {
		return id, fmt.Errorf("quuid: nil entropy reader")
	}
	binary.BigEndian.PutUint64(id[:8], uint64(ms))
	if _, err := io.ReadFull(r, id[8:]); err != nil {
		return SortableID{}, fmt.Errorf("quuid: read entropy: %w", err)
	}
	if allZero(id[8:]) {
		return SortableID{}, ErrZeroEntropy
	}
	return id, nil
}

// MustNewSortableID is like NewSortableID but panics when generation fails.
func MustNewSortableID() SortableID {
	id, err := NewSortableID()
	if err != nil {
		panic(err)
	}
	return id
}

// ParseSortableID parses a SortableID from its supported text representations.
func ParseSortableID(s string) (SortableID, error) {
	raw, err := decodeText(s, sortableIDPrefix, SortableIDSize)
	if err != nil {
		return SortableID{}, err
	}
	var id SortableID
	copy(id[:], raw)
	return id, nil
}

// MustParseSortableID is like ParseSortableID but panics on malformed input.
func MustParseSortableID(s string) SortableID {
	id, err := ParseSortableID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// String returns the canonical text representation.
func (id SortableID) String() string { return encodeText(sortableIDPrefix, id[:]) }

// Hex returns the lowercase hexadecimal representation without a type prefix.
func (id SortableID) Hex() string { return fmt.Sprintf("%x", id[:]) }

// Bytes returns a defensive copy of the binary representation.
func (id SortableID) Bytes() []byte { return append([]byte(nil), id[:]...) }

// IsZero reports whether every byte is zero.
func (id SortableID) IsZero() bool { return allZero(id[:]) }

// Timestamp returns the UTC timestamp embedded in a SortableID.
func (id SortableID) Timestamp() time.Time {
	ms := binary.BigEndian.Uint64(id[:8])
	if ms > uint64(^uint64(0)>>1) {
		return time.Time{}
	}
	return time.UnixMilli(int64(ms)).UTC()
}

// Entropy returns the 256-bit random portion of a SortableID.
func (id SortableID) Entropy() ID {
	var entropy ID
	copy(entropy[:], id[8:])
	return entropy
}

// Equal reports whether two values contain identical bytes.
func (id SortableID) Equal(other SortableID) bool {
	return subtle.ConstantTimeCompare(id[:], other[:]) == 1
}

// Compare compares two values lexicographically and returns -1, 0, or 1.
func (id SortableID) Compare(other SortableID) int {
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
func (id SortableID) MarshalText() ([]byte, error) { return id.AppendText(nil) }

// AppendText appends the canonical text representation to dst.
func (id SortableID) AppendText(dst []byte) ([]byte, error) {
	return appendEncodedText(dst, sortableIDPrefix, id[:]), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (id *SortableID) UnmarshalText(text []byte) error {
	if id == nil {
		return fmt.Errorf("quuid: UnmarshalText on nil *SortableID")
	}
	parsed, err := ParseSortableID(string(text))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// MarshalJSON implements json.Marshaler.
func (id SortableID) MarshalJSON() ([]byte, error) { return json.Marshal(id.String()) }

// UnmarshalJSON implements json.Unmarshaler.
func (id *SortableID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return ErrNilValue
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("quuid: decode SortableID JSON: %w", err)
	}
	return id.UnmarshalText([]byte(value))
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (id SortableID) MarshalBinary() ([]byte, error) { return id.AppendBinary(nil) }

// AppendBinary appends the binary representation to dst.
func (id SortableID) AppendBinary(dst []byte) ([]byte, error) {
	return append(dst, id[:]...), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (id *SortableID) UnmarshalBinary(data []byte) error {
	if id == nil {
		return fmt.Errorf("quuid: UnmarshalBinary on nil *SortableID")
	}
	if len(data) != SortableIDSize {
		return ErrInvalidLength
	}
	copy(id[:], data)
	return nil
}

// Scan implements database/sql.Scanner.
func (id *SortableID) Scan(src any) error {
	if id == nil {
		return fmt.Errorf("quuid: Scan on nil *SortableID")
	}
	var parsed SortableID
	switch value := src.(type) {
	case nil:
		return ErrNilValue
	case string:
		var err error
		parsed, err = ParseSortableID(value)
		if err != nil {
			return err
		}
	case []byte:
		if len(value) == SortableIDSize {
			copy(parsed[:], value)
		} else {
			var err error
			parsed, err = ParseSortableID(string(value))
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
func (id SortableID) Value() (driver.Value, error) { return id.String(), nil }
