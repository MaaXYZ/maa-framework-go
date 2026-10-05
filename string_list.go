package maa

import (
	"bytes"
	"errors"
	"fmt"
)

// StringList is an ordered list of strings.
// JSON input may be a single string or an array of strings; a scalar
// normalizes to a one-element list. Non-nil lists marshal as arrays.
// In pipeline parameters, nil is omitted to inherit the existing/default
// value; an empty non-nil list clears it.
type StringList []string

// UnmarshalJSON normalizes a scalar string to a one-element list.
// Invalid input, including null, leaves the receiver unchanged.
func (s *StringList) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		return errors.New("string list: expected a string or array of strings")
	}
	if data[0] == '[' {
		var items []*string
		if err := unmarshalJSON(data, &items); err != nil {
			return fmt.Errorf("string list: %w", err)
		}
		list := make(StringList, len(items))
		for i, item := range items {
			if item == nil {
				return errors.New("string list: array items must all be strings")
			}
			list[i] = *item
		}
		*s = list
		return nil
	}
	var single string
	if err := unmarshalJSON(data, &single); err != nil {
		return fmt.Errorf("string list: expected a string or array of strings: %w", err)
	}
	*s = StringList{single}
	return nil
}
