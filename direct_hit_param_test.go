package maa

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirectHitParam_ROIOffsetJSON(t *testing.T) {
	for _, tc := range []struct {
		name   string
		offset *Rect
		json   string
	}{
		{name: "unspecified", json: `{}`},
		{name: "explicit zero", offset: &Rect{}, json: `{"roi_offset":[0,0,0,0]}`},
		{name: "nonzero", offset: &Rect{1, 2, 3, 4}, json: `{"roi_offset":[1,2,3,4]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := json.Marshal(DirectHitParam{ROIOffset: tc.offset})
			require.NoError(t, err)
			require.JSONEq(t, tc.json, string(encoded))

			var decoded DirectHitParam
			require.NoError(t, json.Unmarshal([]byte(tc.json), &decoded))
			require.Equal(t, tc.offset, decoded.ROIOffset)
		})
	}
}

func TestDirectHitParam_NativeROIOffsetInheritance(t *testing.T) {
	res := createResource(t)
	t.Cleanup(func() { require.NoError(t, res.Destroy()) })

	const name = "DirectHitOffset"
	const initial = `{"DirectHitOffset":{"recognition":{"type":"DirectHit","param":{"roi":[10,20,30,40],"roi_offset":[1,2,3,4]}}}}`
	require.NoError(t, res.OverridePipeline(initial))

	for _, tc := range []struct {
		name   string
		offset *Rect
		want   Rect
	}{
		{name: "inherit nonzero", want: Rect{1, 2, 3, 4}},
		{name: "clear", offset: &Rect{}, want: Rect{}},
		{name: "inherit zero", want: Rect{}},
		{name: "set nonzero", offset: &Rect{5, 6, 7, 8}, want: Rect{5, 6, 7, 8}},
		{name: "inherit new offset", want: Rect{5, 6, 7, 8}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recognition := RecDirectHit()
			recognition.Param.(*DirectHitParam).ROIOffset = tc.offset
			require.NoError(t, res.OverridePipeline(map[string]*Node{
				name: NewNode(name).SetRecognition(recognition),
			}))
			node, err := res.GetNode(name)
			require.NoError(t, err)
			param := node.Recognition.Param.(*DirectHitParam)
			require.NotNil(t, param.ROIOffset)
			require.Equal(t, tc.want, *param.ROIOffset)
			roi, err := param.ROI.AsRect()
			require.NoError(t, err)
			require.Equal(t, Rect{10, 20, 30, 40}, roi)
		})
	}
}
