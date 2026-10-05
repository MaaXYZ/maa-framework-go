package rect

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/jsoncodec"
)

// Rect represents a 2D rectangle area
type Rect [4]int

func (r Rect) X() int {
	return r[0]
}

func (r Rect) Y() int {
	return r[1]
}

func (r Rect) Width() int {
	return r[2]
}

func (r Rect) Height() int {
	return r[3]
}

// IsZero reports whether both coordinates and both dimensions are zero.
// It lets the encoding/json "omitzero" option omit empty rectangles,
// so pipeline defaults and parent-node inheritance apply.
func (r Rect) IsZero() bool {
	return r == Rect{}
}

// UnmarshalJSON decodes a rectangle from a JSON array of integers,
// mirroring the upstream pipeline parser: [x, y] expands to a 1x1
// rectangle at (x, y), and [x, y, w, h] is used as-is. Any other
// length, null, or non-integer element is an error; on error the
// rectangle is left unchanged.
func (r *Rect) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("rect must not be null")
	}
	var elements []*int
	if err := jsoncodec.Unmarshal(data, &elements); err != nil {
		return errors.New("rect must be an array of integers")
	}
	for _, element := range elements {
		if element == nil {
			return errors.New("rect elements must be integers")
		}
	}
	switch len(elements) {
	case 2:
		*r = Rect{*elements[0], *elements[1], 1, 1}
	case 4:
		*r = Rect{*elements[0], *elements[1], *elements[2], *elements[3]}
	default:
		return fmt.Errorf("rect must have 2 or 4 elements, got %d", len(elements))
	}
	return nil
}
