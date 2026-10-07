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

func NewBool(b bool) Target {
	return Target{
		tp:  targetBool,
		val: b,
	}
}

func NewString(s string) Target {
	return Target{
		tp:  targetString,
		val: s,
	}
}

func NewRect(r rect.Rect) Target {
	return Target{
		tp:  targetRect,
		val: r,
	}
}

func (t Target) IsZero() bool   { return t.tp == targetNone }
func (t Target) IsBool() bool   { return t.tp == targetBool }
func (t Target) IsString() bool { return t.tp == targetString }
func (t Target) IsRect() bool   { return t.tp == targetRect }

func (t Target) AsBool() (bool, error) {
	if !t.IsBool() {
		return false, errors.New("target is not a boolean")
	}
	return t.val.(bool), nil
}

func (t Target) AsString() (string, error) {
	if !t.IsString() {
		return "", errors.New("target is not a string")
	}
	return t.val.(string), nil
}

func (t Target) AsRect() (rect.Rect, error) {
	if !t.IsRect() {
		return rect.Rect{}, errors.New("target is not a rect")
	}
	return t.val.(rect.Rect), nil
}

func (t Target) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
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
