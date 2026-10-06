package maa

import (
	"bytes"
	"errors"
	"fmt"
	"math"
)

// ClassSelector selects a neural network class by index or label.
// Its zero value selects class index 0.
type ClassSelector struct {
	index   int
	label   string
	isLabel bool
}

// ClassIndex selects a class by its numeric index.
func ClassIndex(index int) ClassSelector { return ClassSelector{index: index} }

// ClassLabel selects a class by its label.
func ClassLabel(label string) ClassSelector {
	return ClassSelector{label: label, isLabel: true}
}

// IsIndex reports whether the selector selects by class index.
func (c ClassSelector) IsIndex() bool { return !c.isLabel }

// IsLabel reports whether the selector selects by class label.
func (c ClassSelector) IsLabel() bool { return c.isLabel }

// AsIndex returns the index, or an error if the selector holds a label.
func (c ClassSelector) AsIndex() (int, error) {
	if c.isLabel {
		return 0, errors.New("class selector is not an index")
	}
	return c.index, nil
}

// AsLabel returns the label, or an error if the selector holds an index.
func (c ClassSelector) AsLabel() (string, error) {
	if !c.isLabel {
		return "", errors.New("class selector is not a label")
	}
	return c.label, nil
}

// MarshalJSON encodes a selector as an integer or string.
// It returns an error for an index outside the int32 range, which the
// pipeline parser rejects.
func (c ClassSelector) MarshalJSON() ([]byte, error) {
	if c.isLabel {
		return marshalJSON(c.label)
	}
	if c.index < math.MinInt32 || c.index > math.MaxInt32 {
		return nil, fmt.Errorf("class selector: index %d outside the int32 range", c.index)
	}
	return marshalJSON(c.index)
}

// UnmarshalJSON accepts only an integer or string and leaves c unchanged on
// error. Indexes are limited to the int32 range the pipeline parser accepts.
func (c *ClassSelector) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		return errors.New("class selector: expected an integer or string")
	}
	if data[0] == '"' {
		var label string
		if err := unmarshalJSON(data, &label); err != nil {
			return fmt.Errorf("class selector: %w", err)
		}
		*c = ClassLabel(label)
		return nil
	}
	var index int
	if err := unmarshalJSON(data, &index); err != nil {
		return fmt.Errorf("class selector: expected an integer or string: %w", err)
	}
	if index < math.MinInt32 || index > math.MaxInt32 {
		return fmt.Errorf("class selector: index %d outside the int32 range", index)
	}
	*c = ClassIndex(index)
	return nil
}

// ClassSelectors is an ordered list of class indices and labels.
// JSON input may be a single integer or string, or an array of either.
// Non-nil lists marshal as arrays. In neural network parameters, nil is omitted
// to inherit the existing/default selection; an empty non-nil list clears the
// selection, matching all classes (detection results are still threshold-filtered).
type ClassSelectors []ClassSelector

// UnmarshalJSON normalizes a scalar to a one-element list. Invalid input,
// including null, leaves the receiver unchanged.
func (c *ClassSelectors) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) > 0 && data[0] == '[' {
		var selectors []ClassSelector
		if err := unmarshalJSON(data, &selectors); err != nil {
			return fmt.Errorf("class selectors: %w", err)
		}
		*c = selectors
		return nil
	}
	var selector ClassSelector
	if err := unmarshalJSON(data, &selector); err != nil {
		return err
	}
	*c = ClassSelectors{selector}
	return nil
}
