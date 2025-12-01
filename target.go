package maa

import "github.com/MaaXYZ/maa-framework-go/v3/internal/target"

type Target = target.Target

func NewTargetBool(val bool) Target {
	return target.NewBool(val)
}

func NewTargetString(val string) Target {
	return target.NewString(val)
}

func NewTargetRect(val Rect) Target {
	return target.NewRect(val)
}
