package quuid

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Scan accepts canonical UUID text or exactly 16 binary bytes. SQL NULL is
// rejected; use NullUUID or NullBinaryUUID for nullable columns.
func (id *UUID) Scan(src any) error {
	if id == nil {
		return fmt.Errorf("quuid: Scan on nil *UUID")
	}
	parsed, err := scanUUIDSource(src)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value returns canonical UUID text. This is the safest portable default for
// PostgreSQL uuid columns and text-oriented drivers. Use UUID.Binary() for
// MySQL BINARY(16), SQLite BLOB, PostgreSQL bytea, or another binary column.
func (id UUID) Value() (driver.Value, error) { return id.String(), nil }

// BinaryUUID is an explicit database representation wrapper. It has the same
// logical UUID value but its driver.Value is []byte with exactly 16 bytes.
// There is no package-global SQL mode, so text and binary columns may coexist.
type BinaryUUID UUID

// Binary returns an explicit binary database representation of id.
func (id UUID) Binary() BinaryUUID { return BinaryUUID(id) }

// UUID returns the logical UUID represented by id.
func (id BinaryUUID) UUID() UUID { return UUID(id) }

// String returns the canonical text representation.
func (id BinaryUUID) String() string { return UUID(id).String() }

// Bytes returns a defensive copy of the binary representation.
func (id BinaryUUID) Bytes() []byte { return UUID(id).Bytes() }

// IsZero reports whether every byte is zero.
func (id BinaryUUID) IsZero() bool { return UUID(id).IsZero() }

// Equal reports whether two values contain identical bytes.
func (id BinaryUUID) Equal(other BinaryUUID) bool { return UUID(id).Equal(UUID(other)) }

// Compare compares two values lexicographically and returns -1, 0, or 1.
func (id BinaryUUID) Compare(other BinaryUUID) int { return UUID(id).Compare(UUID(other)) }

// Version returns the UUID version field.
func (id BinaryUUID) Version() Version { return UUID(id).Version() }

// Variant returns the UUID variant field.
func (id BinaryUUID) Variant() Variant { return UUID(id).Variant() }

// MarshalText implements encoding.TextMarshaler.
func (id BinaryUUID) MarshalText() ([]byte, error) { return UUID(id).MarshalText() }

// AppendText appends the canonical text representation to dst.
func (id BinaryUUID) AppendText(dst []byte) ([]byte, error) {
	return UUID(id).AppendText(dst)
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (id BinaryUUID) MarshalBinary() ([]byte, error) { return UUID(id).MarshalBinary() }

// AppendBinary appends the binary representation to dst.
func (id BinaryUUID) AppendBinary(dst []byte) ([]byte, error) {
	return UUID(id).AppendBinary(dst)
}

// MarshalJSON implements json.Marshaler.
func (id BinaryUUID) MarshalJSON() ([]byte, error) { return UUID(id).MarshalJSON() }

// UnmarshalText implements encoding.TextUnmarshaler.
func (id *BinaryUUID) UnmarshalText(text []byte) error {
	if id == nil {
		return fmt.Errorf("quuid: UnmarshalText on nil *BinaryUUID")
	}
	parsed, err := ParseUUID(string(text))
	if err != nil {
		return err
	}
	*id = BinaryUUID(parsed)
	return nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (id *BinaryUUID) UnmarshalBinary(data []byte) error {
	if id == nil {
		return fmt.Errorf("quuid: UnmarshalBinary on nil *BinaryUUID")
	}
	if len(data) != UUIDSize {
		return ErrInvalidLength
	}
	copy((*id)[:], data)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (id *BinaryUUID) UnmarshalJSON(data []byte) error {
	if id == nil {
		return fmt.Errorf("quuid: UnmarshalJSON on nil *BinaryUUID")
	}
	if string(data) == "null" {
		return ErrNilValue
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("quuid: decode BinaryUUID JSON: %w", err)
	}
	return id.UnmarshalText([]byte(value))
}

// Scan implements database/sql.Scanner.
func (id *BinaryUUID) Scan(src any) error {
	if id == nil {
		return fmt.Errorf("quuid: Scan on nil *BinaryUUID")
	}
	parsed, err := scanUUIDSource(src)
	if err != nil {
		return err
	}
	*id = BinaryUUID(parsed)
	return nil
}

// Value implements database/sql/driver.Valuer.
func (id BinaryUUID) Value() (driver.Value, error) {
	return append([]byte(nil), id[:]...), nil
}

func scanUUIDSource(src any) (UUID, error) {
	switch value := src.(type) {
	case nil:
		return NilUUID, ErrNilValue
	case string:
		return ParseUUID(value)
	case []byte:
		if len(value) == UUIDSize {
			var id UUID
			copy(id[:], value)
			return id, nil
		}
		return ParseUUID(string(value))
	default:
		return NilUUID, fmt.Errorf("%w: %T", ErrInvalidSource, src)
	}
}
