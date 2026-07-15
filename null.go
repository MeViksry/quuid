package quuid

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// NullID represents an ID that may be NULL in SQL or null in JSON.
type NullID struct {
	ID    ID
	Valid bool
}

// Scan implements database/sql.Scanner.
func (n *NullID) Scan(src any) error {
	if n == nil {
		return fmt.Errorf("quuid: Scan on nil *NullID")
	}
	if src == nil {
		n.ID = ID{}
		n.Valid = false
		return nil
	}
	id, err := scanIDSource(src)
	if err != nil {
		n.ID = ID{}
		n.Valid = false
		return err
	}
	n.ID = id
	n.Valid = true
	return nil
}

// Value implements database/sql/driver.Valuer.
func (n NullID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.ID.String(), nil
}

// MarshalJSON implements json.Marshaler.
func (n NullID) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *NullID) UnmarshalJSON(data []byte) error {
	if n == nil {
		return fmt.Errorf("quuid: UnmarshalJSON on nil *NullID")
	}
	if string(data) == "null" {
		n.ID = ID{}
		n.Valid = false
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	id, err := ParseID(value)
	if err != nil {
		return err
	}
	n.ID = id
	n.Valid = true
	return nil
}
