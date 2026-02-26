package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Float64OrString is a type that can hold a float64 or a string.
type Float64OrString struct {
	value float64
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (f *Float64OrString) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v := v.(type) {
	case float64:
		f.value = v
	case string:
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		f.value = val
	default:
		return fmt.Errorf("invalid type for float64OrString: %T", v)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (f Float64OrString) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.value)
}

// String returns the string representation of the float64 value
func (f Float64OrString) String() string {
	return strconv.FormatFloat(f.value, 'f', -1, 64)
}

// Value returns the float64 value
func (f Float64OrString) Value() float64 {
	return f.value
}