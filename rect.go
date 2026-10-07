package maa

import "github.com/MaaXYZ/maa-framework-go/v4/internal/rect"

// Rect is a [4]int rectangle: x, y, width, height. It is the Go type for
// pipeline ROI offsets, action target offsets, and result boxes.
//
// JSON decodes [x, y] as a 1x1 rectangle at (x, y) and [x, y, w, h] as-is,
// mirroring the upstream pipeline parser; null, any other length, or a
// non-integer element is an error and leaves the rectangle unchanged. The
// zero value is omitted by the encoding/json "omitzero" option, so pipeline
// defaults and parent-node inheritance apply.
type Rect = rect.Rect
