// Package target implements the pipeline v2 target wire type: a variant that
// is either the boolean true, a node or [Anchor] name, or a rectangle.
package target

import (
	"bytes"
	"errors"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/jsoncodec"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/rect"
)

type targetType int

const (
	targetNone targetType = iota
	targetBool
	targetString
	targetRect
)

// Target is a type-safe variant that can hold one of three value types:
// bool, string, or rect.Rect. It provides methods for type checking,
// safe value retrieval, and JSON serialization/deserialization.
type Target struct {
	tp  targetType
	val any
}

// NewBool creates the boolean-true target when b is true; false is not a
// valid pipeline target and fails to encode.
func NewBool(b bool) Target {
	return Target{
		tp:  targetBool,
		val: b,
	}
}

// NewString creates a target that refers to a node name or an [Anchor]-prefixed
// anchor name. The string is passed through as-is; syntax is validated when
// the pipeline is parsed.
func NewString(s string) Target {
	return Target{
		tp:  targetString,
		val: s,
	}
}

// NewRect creates a target for an explicit rectangle.
func NewRect(r rect.Rect) Target {
	return Target{
		tp:  targetRect,
		val: r,
	}
}

// IsZero reports whether the target is the zero Target, which carries no
// value: it is omitted by the encoding/json "omitzero" option and fails to
// encode standalone.
func (t Target) IsZero() bool { return t.tp == targetNone }

// IsBool reports whether the target holds the boolean true.
func (t Target) IsBool() bool { return t.tp == targetBool }

// IsString reports whether the target holds a node or anchor name.
func (t Target) IsString() bool { return t.tp == targetString }

// IsRect reports whether the target holds a rectangle.
func (t Target) IsRect() bool { return t.tp == targetRect }

// AsBool returns the boolean value; it fails if the target does not hold one.
func (t Target) AsBool() (bool, error) {
	if !t.IsBool() {
		return false, errors.New("target is not a boolean")
	}
	return t.val.(bool), nil
}

// AsString returns the node or anchor name; it fails if the target does not
// hold a string.
func (t Target) AsString() (string, error) {
	if !t.IsString() {
		return "", errors.New("target is not a string")
	}
	return t.val.(string), nil
}

// AsRect returns the rectangle; it fails if the target does not hold one.
func (t Target) AsRect() (rect.Rect, error) {
	if !t.IsRect() {
		return rect.Rect{}, errors.New("target is not a rect")
	}
	return t.val.(rect.Rect), nil
}

// MarshalJSON encodes the target: the boolean true as JSON true, a string as
// a JSON string, and a rectangle as a four-element array. A zero target and a
// wrapped false are not valid pipeline targets and fail to encode.
func (t Target) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return nil, errors.New("cannot encode a zero pipeline target")
	}

	switch t.tp {
	case targetBool:
		if !t.val.(bool) {
			return nil, errors.New("pipeline target must be true, a string, or a point/rectangle")
		}
		return jsoncodec.Marshal(true)
	case targetString:
		return jsoncodec.Marshal(t.val.(string))
	case targetRect:
		return jsoncodec.Marshal(t.val.(rect.Rect))
	default:
		return nil, errors.New("unknown target type")
	}
}

// UnmarshalJSON decodes the JSON boolean true, a string, or an integer array
// of two elements (a point, normalized to a 1x1 rectangle) or four elements
// (a rectangle). null, false, numbers, objects, other lengths, and
// non-integer or null elements are errors; on error the target is left
// unchanged.
func (t *Target) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if string(data) == "null" {
		return errors.New("pipeline target must not be null")
	}

	if string(data) == "true" {
		t.tp = targetBool
		t.val = true
		return nil
	}

	var s string
	if err := jsoncodec.Unmarshal(data, &s); err == nil {
		t.tp = targetString
		t.val = s
		return nil
	}

	var coordinates []*int
	if err := jsoncodec.Unmarshal(data, &coordinates); err == nil {
		if len(coordinates) != 2 && len(coordinates) != 4 {
			return errors.New("pipeline target requires exactly 2 or 4 integer coordinates")
		}
		for _, coordinate := range coordinates {
			if coordinate == nil {
				return errors.New("pipeline target coordinates must be integers")
			}
		}
		r := rect.Rect{*coordinates[0], *coordinates[1], 1, 1}
		if len(coordinates) == 4 {
			r[2], r[3] = *coordinates[2], *coordinates[3]
		}
		*t = NewRect(r)
		return nil
	}

	return errors.New("pipeline target must be true, a string, or a point/rectangle")
}
