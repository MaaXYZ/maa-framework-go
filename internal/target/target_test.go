package target

import (
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/rect"
	"github.com/stretchr/testify/require"
)

func TestTarget_IsZero(t *testing.T) {
	type Case struct {
		Name   string
		Target Target
		Expect bool
	}

	cases := []Case{
		{
			Name:   "IsZero",
			Target: Target{},
			Expect: true,
		},
		{
			Name:   "IsNotZero",
			Target: NewBool(true),
			Expect: false,
		},
		{
			Name:   "IsNotZero",
			Target: NewString("test"),
			Expect: false,
		},
		{
			Name:   "IsNotZero",
			Target: NewRect(rect.Rect{100, 100, 100, 100}),
			Expect: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			got := tc.Target.IsZero()
			require.Equal(t, tc.Expect, got)
		})
	}

}
