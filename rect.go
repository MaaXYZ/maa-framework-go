package maa

type Rect struct {
	X, Y, W, H int32
}

func (r Rect) ToInts() [4]int32 {
	return [4]int32{r.X, r.Y, r.W, r.H}
}
