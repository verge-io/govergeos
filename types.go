package vergeos

import (
	"encoding/json"
	"strconv"
)

// FlexInt is an int that can be unmarshaled from either a JSON number or string.
// The VergeOS API sometimes returns IDs as strings when specific fields are requested.
type FlexInt int

// UnmarshalJSON implements json.Unmarshaler for FlexInt.
func (f *FlexInt) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as int first
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*f = FlexInt(i)
		return nil
	}

	// Try to unmarshal as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "" {
			*f = 0
			return nil
		}
		i, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		*f = FlexInt(i)
		return nil
	}

	// Default to 0
	*f = 0
	return nil
}

// MarshalJSON implements json.Marshaler for FlexInt.
func (f FlexInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(f))
}

// Int returns the FlexInt as an int.
func (f FlexInt) Int() int {
	return int(f)
}

// FlexFK is an int that can be unmarshaled from a JSON number, string, or
// an object with a "$key" or "id" field. This handles VergeOS API responses
// where foreign key fields may be returned as expanded objects depending on
// the requested field selection.
type FlexFK int

// UnmarshalJSON implements json.Unmarshaler for FlexFK.
func (f *FlexFK) UnmarshalJSON(data []byte) error {
	// Try int first
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*f = FlexFK(i)
		return nil
	}

	// Try string (some fields return IDs as strings)
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "" {
			*f = 0
			return nil
		}
		i, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		*f = FlexFK(i)
		return nil
	}

	// Try object with $key or id (expanded FK)
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err == nil {
		if raw, ok := obj["$key"]; ok {
			var key int
			if err := json.Unmarshal(raw, &key); err == nil {
				*f = FlexFK(key)
				return nil
			}
		}
		if raw, ok := obj["id"]; ok {
			var id int
			if err := json.Unmarshal(raw, &id); err == nil {
				*f = FlexFK(id)
				return nil
			}
		}
	}

	*f = 0
	return nil
}

// MarshalJSON implements json.Marshaler for FlexFK.
func (f FlexFK) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(f))
}

// Int returns the FlexFK as an int.
func (f FlexFK) Int() int {
	return int(f)
}
