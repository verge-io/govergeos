package vergeos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// jsonString reads a JSON string, number, null, or an object that carries
// name, $key, or id. Missing and null become an empty string.
func jsonString(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return "", nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, nil
	}
	var n json.Number
	if err := json.Unmarshal(raw, &n); err == nil {
		return n.String(), nil
	}
	if raw[0] == '{' {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(raw, &obj); err != nil {
			return "", err
		}
		for _, key := range []string{"name", "$key", "id"} {
			if nested, ok := obj[key]; ok {
				return jsonString(nested)
			}
		}
		return "", nil
	}
	return "", fmt.Errorf("vergeos: cannot read %s as a string", raw)
}

// jsonKey reads a foreign key. An expanded object yields $key or id, which
// is the row id, rather than the object's display name.
func jsonKey(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return "", nil
	}
	if raw[0] == '{' {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(raw, &obj); err != nil {
			return "", err
		}
		for _, key := range []string{"$key", "id"} {
			if nested, ok := obj[key]; ok {
				return jsonString(nested)
			}
		}
		return "", nil
	}
	return jsonString(raw)
}

// jsonBool reads a JSON bool, 0/1, null, or the words true/false/yes/no/on/off.
func jsonBool(raw json.RawMessage) (bool, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return false, nil
	}
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		return b, nil
	}
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n != 0, nil
	}
	s, err := jsonString(raw)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true, nil
	case "", "0", "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("vergeos: cannot read %s as a bool", raw)
	}
}

// jsonInt64 reads a JSON number, numeric string, or null.
func jsonInt64(raw json.RawMessage) (int64, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return 0, nil
	}
	var n int64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n, nil
	}
	s, err := jsonString(raw)
	if err != nil {
		return 0, err
	}
	if s == "" {
		return 0, nil
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("vergeos: cannot read %s as an integer", raw)
	}
	return int64(f), nil
}

// jsonFlexIntPtr reads an int foreign key. Null and a missing value are nil.
// An expanded object is read from $key or id.
func jsonFlexIntPtr(raw json.RawMessage) (*FlexInt, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}
	var fk FlexFK
	if err := json.Unmarshal(raw, &fk); err != nil {
		return nil, err
	}
	v := FlexInt(fk)
	return &v, nil
}

func fieldErr(name string, err error) error {
	return fmt.Errorf("vergeos: field %s: %w", name, err)
}
