package vergeos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

	return fmt.Errorf("vergeos: cannot unmarshal %s into FlexInt", string(data))
}

// MarshalJSON implements json.Marshaler for FlexInt.
func (f FlexInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(f))
}

// Int returns the FlexInt as an int.
func (f FlexInt) Int() int {
	return int(f)
}

// FlexText is a string that also accepts a JSON number or null.
// VergeOS uses it for values that are an id on one release and a name on another.
type FlexText string

// UnmarshalJSON implements json.Unmarshaler for FlexText.
func (f *FlexText) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*f = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexText(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexText(n.String())
		return nil
	}
	return fmt.Errorf("vergeos: cannot unmarshal %s into FlexText", data)
}

// MarshalJSON implements json.Marshaler for FlexText.
func (f FlexText) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(f))
}

// String returns the text.
func (f FlexText) String() string {
	return string(f)
}

// StringList is a list of strings. VergeOS returns some list columns as a
// JSON array, some as one delimited string, and some as null.
type StringList []string

// UnmarshalJSON implements json.Unmarshaler for StringList.
func (s *StringList) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*s = nil
		return nil
	}
	switch data[0] {
	case '[':
		var list []string
		if err := json.Unmarshal(data, &list); err != nil {
			return err
		}
		*s = list
		return nil
	case '"':
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		*s = splitAPIList(text)
		return nil
	default:
		var n json.Number
		if err := json.Unmarshal(data, &n); err != nil {
			return fmt.Errorf("vergeos: cannot unmarshal %s into StringList", data)
		}
		if n.String() == "" {
			*s = nil
			return nil
		}
		*s = StringList{n.String()}
		return nil
	}
}

// MarshalJSON implements json.Marshaler for StringList.
func (s StringList) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	return json.Marshal([]string(s))
}

// splitAPIList splits a delimited list. Spaces stay inside an entry.
func splitAPIList(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
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

	return fmt.Errorf("vergeos: cannot unmarshal %s into FlexFK", string(data))
}

// MarshalJSON implements json.Marshaler for FlexFK.
func (f FlexFK) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(f))
}

// Int returns the FlexFK as an int.
func (f FlexFK) Int() int {
	return int(f)
}

// FlexString is a string that can be unmarshaled from a JSON string, number, or null.
//
// Reference columns such as owner and creator arrive as a "table/key" string
// or as a bare key. A bare key is kept as its decimal text.
type FlexString string

// UnmarshalJSON implements json.Unmarshaler for FlexString.
func (f *FlexString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*f = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexString(s)
		return nil
	}
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexString(strconv.FormatInt(n, 10))
		return nil
	}
	return fmt.Errorf("vergeos: cannot unmarshal %s into FlexString", data)
}

// MarshalJSON implements json.Marshaler for FlexString.
func (f FlexString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(f))
}

// String returns the reference text.
func (f FlexString) String() string {
	return string(f)
}

// redactedSecret is what formatting prints in place of a client secret.
const redactedSecret = "[redacted]"

// clientSecretField is the JSON name of an OAuth client secret.
// Auth source settings and OIDC applications both use it. Responses drop
// the value, and formatting never prints it.
const clientSecretField = "client_secret"

// WriteOnlySecret is a client secret.
//
// JSON encoding writes the secret so it can be sent to the API. JSON
// decoding discards the payload, so a response cannot fill the field.
// String, GoString, and Format print "[redacted]" when a secret is set
// and print nothing when it is empty. Call Value to read the secret.
// Logging the WriteOnlySecret itself does not log the secret.
type WriteOnlySecret struct {
	v string
}

// NewWriteOnlySecret returns a write-only secret holding value.
func NewWriteOnlySecret(value string) WriteOnlySecret {
	return WriteOnlySecret{v: value}
}

// Value returns the secret. Printing the WriteOnlySecret does not.
func (s WriteOnlySecret) Value() string {
	return s.v
}

// String returns "[redacted]" when a secret is set.
func (s WriteOnlySecret) String() string {
	if s.v == "" {
		return ""
	}
	return redactedSecret
}

// GoString returns the same redacted text as String.
func (s WriteOnlySecret) GoString() string {
	return s.String()
}

// Format prints the redacted form for every verb except %T and %p,
// which fmt handles before calling Format.
func (s WriteOnlySecret) Format(f fmt.State, _ rune) {
	fmt.Fprint(f, s.String())
}

// formatRedacted writes v with the caller's fmt flags. v must not be a
// value whose Format or String method calls formatRedacted on itself.
func formatRedacted(f fmt.State, verb rune, v any) {
	format := "%"
	if f.Flag('+') {
		format += "+"
	}
	if f.Flag('#') {
		format += "#"
	}
	if w, ok := f.Width(); ok {
		format += fmt.Sprintf("%d", w)
	}
	if p, ok := f.Precision(); ok {
		format += fmt.Sprintf(".%d", p)
	}
	format += string(verb)
	fmt.Fprintf(f, format, v)
}

// MarshalJSON writes the secret so an API request can carry it.
func (s WriteOnlySecret) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}

// UnmarshalJSON discards the payload. A client secret read back from
// the API is not stored.
func (s *WriteOnlySecret) UnmarshalJSON(_ []byte) error {
	if s != nil {
		*s = WriteOnlySecret{}
	}
	return nil
}
