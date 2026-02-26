package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Int64OrString is a type that can hold an int64 or a string.
type Int64OrString struct {
	value int64
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (i *Int64OrString) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v := v.(type) {
	case float64:
		i.value = int64(v)
	case string:
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		i.value = val
	default:
		return fmt.Errorf("invalid type for int64OrString: %T", v)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (i Int64OrString) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.value)
}

// String returns the s(tring )representation of the int64 value
func (i Int64OrString) String() string {
	return strconv.FormatInt(i.value, 10)
}
func (i Int64OrString) Value() int64 {
	return i.value
}
