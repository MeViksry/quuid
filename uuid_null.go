package quuid

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// NullUUID represents a UUID that may be SQL NULL or JSON null.
type NullUUID struct {
	UUID  UUID
	Valid bool
}

// NewNullUUID marks id as valid even when it is NilUUID.
func NewNullUUID(id UUID) NullUUID { return NullUUID{UUID: id, Valid: true} }

// ToNullUUID returns an invalid value for NilUUID and a valid value otherwise.
func ToNullUUID(id UUID) NullUUID { return NullUUID{UUID: id, Valid: !id.IsZero()} }

// Scan implements database/sql.Scanner.
func (n *NullUUID) Scan(src any) error {
	if n == nil {
		return fmt.Errorf("quuid: Scan on nil *NullUUID")
	}
	if isNullDatabaseValue(src) {
		n.UUID = NilUUID
		n.Valid = false
		return nil
	}
	id, err := scanUUIDSource(src)
	if err != nil {
		n.UUID = NilUUID
		n.Valid = false
		return err
	}
	n.UUID = id
	n.Valid = true
	return nil
}

// Value implements database/sql/driver.Valuer.
func (n NullUUID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.UUID.Value()
}

// MarshalJSON implements json.Marshaler.
func (n NullUUID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return n.UUID.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *NullUUID) UnmarshalJSON(data []byte) error {
	if n == nil {
		return fmt.Errorf("quuid: UnmarshalJSON on nil *NullUUID")
	}
	if string(data) == "null" {
		n.UUID = NilUUID
		n.Valid = false
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		n.UUID = NilUUID
		n.Valid = false
		return fmt.Errorf("quuid: decode NullUUID JSON: %w", err)
	}
	if value == "" {
		n.UUID = NilUUID
		n.Valid = false
		return nil
	}
	id, err := ParseUUID(value)
	if err != nil {
		n.UUID = NilUUID
		n.Valid = false
		return err
	}
	n.UUID = id
	n.Valid = true
	return nil
}

// NullBinaryUUID is the nullable counterpart of BinaryUUID. Valid values are
// written as exactly 16 bytes; invalid values are written as SQL NULL.
type NullBinaryUUID struct {
	UUID  UUID
	Valid bool
}

// NewNullBinaryUUID marks id as a valid binary UUID value even when it is NilUUID.
func NewNullBinaryUUID(id UUID) NullBinaryUUID {
	return NullBinaryUUID{UUID: id, Valid: true}
}

// ToNullBinaryUUID returns an invalid value for NilUUID and a valid value otherwise.
func ToNullBinaryUUID(id UUID) NullBinaryUUID {
	return NullBinaryUUID{UUID: id, Valid: !id.IsZero()}
}

// Scan implements database/sql.Scanner.
func (n *NullBinaryUUID) Scan(src any) error {
	if n == nil {
		return fmt.Errorf("quuid: Scan on nil *NullBinaryUUID")
	}
	if isNullDatabaseValue(src) {
		n.UUID = NilUUID
		n.Valid = false
		return nil
	}
	id, err := scanUUIDSource(src)
	if err != nil {
		n.UUID = NilUUID
		n.Valid = false
		return err
	}
	n.UUID = id
	n.Valid = true
	return nil
}

// Value implements database/sql/driver.Valuer.
func (n NullBinaryUUID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.UUID.Binary().Value()
}

// MarshalJSON implements json.Marshaler.
func (n NullBinaryUUID) MarshalJSON() ([]byte, error) {
	return NullUUID{UUID: n.UUID, Valid: n.Valid}.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *NullBinaryUUID) UnmarshalJSON(data []byte) error {
	if n == nil {
		return fmt.Errorf("quuid: UnmarshalJSON on nil *NullBinaryUUID")
	}
	var text NullUUID
	if err := text.UnmarshalJSON(data); err != nil {
		n.UUID = NilUUID
		n.Valid = false
		return err
	}
	n.UUID = text.UUID
	n.Valid = text.Valid
	return nil
}

func isNullDatabaseValue(src any) bool {
	switch value := src.(type) {
	case nil:
		return true
	case string:
		return value == ""
	case []byte:
		return len(value) == 0
	default:
		return false
	}
}
