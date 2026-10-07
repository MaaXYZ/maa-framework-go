package maa

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTargetField_DecodeForms(t *testing.T) {
	type Case struct {
		Name   string
		Target string
		Expect Target
	}

	cases := []Case{
		{
			Name:   "Bool",
			Target: "true",
			Expect: NewTargetBool(true),
		},
		{
			Name:   "String",
			Target: `"NodeA"`,
			Expect: NewTargetString("NodeA"),
		},
		{
			Name:   "Point",
			Target: "[100, 200]",
			Expect: NewTargetRect(Rect{100, 200, 1, 1}),
		},
		{
			Name:   "Rect",
			Target: "[100, 200, 30, 40]",
			Expect: NewTargetRect(Rect{100, 200, 30, 40}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var param WaitFreezesParam
			require.NoError(t, json.Unmarshal([]byte(`{"target":`+tc.Target+`}`), &param))
			require.Equal(t, tc.Expect, param.Target)
		})
	}
}

func TestTargetField_DecodeFailureLeavesUnchanged(t *testing.T) {
	seed := func() WaitFreezesParam {
		return WaitFreezesParam{
			Time:         time.Second,
			Target:       NewTargetString("NodeA"),
			TargetOffset: Rect{1, 2, 3, 4},
			Threshold:    0.9,
			Method:       3,
			RateLimit:    500 * time.Millisecond,
			Timeout:      20 * time.Second,
		}
	}
	for name, payload := range map[string]string{
		"explicit null": `{"target":null}`,
		"false":         `{"target":false}`,
		"bad length":    `{"target":[1,2,3]}`,
		"number":        `{"target":123}`,
		"nested array":  `{"target":[[1,2],[3,4]]}`,
	} {
		t.Run(name, func(t *testing.T) {
			seeded := seed()
			require.Error(t, json.Unmarshal([]byte(payload), &seeded))
			require.Equal(t, seed(), seeded)
		})
	}
}
