package target

import (
	"encoding/json"
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/rect"
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

func TestTarget_UnmarshalJSON(t *testing.T) {
	type Case struct {
		Name   string
		JSON   string
		Expect Target
	}

	cases := []Case{
		{
			Name:   "Bool",
			JSON:   "true",
			Expect: NewBool(true),
		},
		{
			Name:   "String",
			JSON:   "\"test\"",
			Expect: NewString("test"),
		},
		{
			Name:   "Rect",
			JSON:   "[100, 100, 100, 100]",
			Expect: NewRect(rect.Rect{100, 100, 100, 100}),
		},
		{
			Name:   "Unknown",
			JSON:   "null",
			Expect: Target{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var got Target
			err := json.Unmarshal([]byte(tc.JSON), &got)
			require.NoError(t, err)
			require.Equal(t, tc.Expect, got)
		})
	}
}
