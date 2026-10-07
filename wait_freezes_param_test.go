package maa

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWaitFreezesParam_MarshalJSON(t *testing.T) {
	param := WaitFreezesParam{
		Time:         1500 * time.Millisecond,
		Target:       NewTargetBool(true),
		TargetOffset: Rect{1, 2, 3, 4},
		Threshold:    0.9,
		Method:       3,
		RateLimit:    500 * time.Millisecond,
		Timeout:      20 * time.Second,
	}
	data, err := json.Marshal(param)
	require.NoError(t, err)
	require.JSONEq(t,
		`{"time":1500,"target":true,"target_offset":[1,2,3,4],"threshold":0.9,"method":3,"rate_limit":500,"timeout":20000}`,
		string(data))
}

func TestWaitFreezesParam_ZeroMarshalsEmpty(t *testing.T) {
	data, err := json.Marshal(WaitFreezesParam{})
	require.NoError(t, err)
	require.JSONEq(t, `{}`, string(data))
}

func TestWaitFreezesParam_UnmarshalJSON(t *testing.T) {
	var param WaitFreezesParam
	require.NoError(t, unmarshalJSON([]byte(
		`{"time":1500,"target":true,"target_offset":[1,2,3,4],"threshold":0.9,"method":3,"rate_limit":500,"timeout":20000}`),
		&param))
	require.Equal(t, 1500*time.Millisecond, param.Time)
	require.Equal(t, NewTargetBool(true), param.Target)
	require.Equal(t, Rect{1, 2, 3, 4}, param.TargetOffset)
	require.Equal(t, 0.9, param.Threshold)
	require.Equal(t, 3, param.Method)
	require.Equal(t, 500*time.Millisecond, param.RateLimit)
	require.Equal(t, 20*time.Second, param.Timeout)
}

func TestWaitFreezesParam_UnmarshalJSON_FailureLeavesTargetUnchanged(t *testing.T) {
	seeded := WaitFreezesParam{
		Time:         time.Second,
		Target:       NewTargetString("NodeA"),
		TargetOffset: Rect{1, 2, 3, 4},
		Threshold:    0.9,
		Method:       3,
		RateLimit:    500 * time.Millisecond,
		Timeout:      20 * time.Second,
	}
	before := seeded
	for name, payload := range map[string]string{
		"threshold type":     `{"threshold":"bad"}`,
		"bad target offset":  `{"target_offset":[1,2,3]}`,
		"non-numeric scalar": `"fast"`,
		"explicit null":      `null`,
		"invalid json":       `{`,
	} {
		t.Run(name, func(t *testing.T) {
			require.Error(t, json.Unmarshal([]byte(payload), &seeded))
			require.Equal(t, before, seeded)
		})
	}
}

func TestWaitFreezesParam_NumericShorthand(t *testing.T) {
	var param WaitFreezesParam
	require.NoError(t, json.Unmarshal([]byte(`500`), &param))
	require.Equal(t, WaitFreezesParam{Time: 500 * time.Millisecond}, param)
	encoded, err := json.Marshal(&param)
	require.NoError(t, err)
	require.JSONEq(t, `{"time":500}`, string(encoded))
}

func TestWaitFreezesParam_JSONRoundTrip(t *testing.T) {
	param := WaitFreezesParam{
		Time:      2 * time.Second,
		Target:    NewTargetRect(Rect{0, 0, 10, 10}),
		Threshold: 0.95,
		Method:    5,
		RateLimit: 1000 * time.Millisecond,
		Timeout:   20 * time.Second,
	}
	data, err := json.Marshal(param)
	require.NoError(t, err)
	var decoded WaitFreezesParam
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, param, decoded)
}
