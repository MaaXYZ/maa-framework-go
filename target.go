package maa

import "github.com/MaaXYZ/maa-framework-go/v4/internal/target"

// Target identifies a pipeline v2 ROI or action position: true for the current
// recognition box (or the full image for an ROI), a node or [Anchor] name, or a rectangle.
// JSON points [x, y] are normalized to [x, y, 1, 1]. A zero Target is unspecified.
type Target = target.Target

// NewTargetBool creates the current-result/full-image target when val is true.
// False is not a valid pipeline target and fails JSON encoding.
func NewTargetBool(val bool) Target {
	return target.NewBool(val)
}

// NewTargetString creates a target that refers to a node name or [Anchor] name.
func NewTargetString(val string) Target {
	return target.NewString(val)
}

// NewTargetRect creates a target for an explicit [x, y, width, height] rectangle.
func NewTargetRect(val Rect) Target {
	return target.NewRect(val)
}
