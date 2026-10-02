package maa

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func intPointer(value int) *int {
	return &value
}

func durationPointer(value time.Duration) *time.Duration {
	return &value
}

func TestPipelineV2InlineSubRecognitionUsesNestedWrapper(t *testing.T) {
	and := RecAnd(Inline(RecDirectHit(), "direct"))
	encoded, err := json.Marshal(and)
	require.NoError(t, err)
	require.JSONEq(t,
		`{"type":"And","param":{"all_of":[{"sub_name":"direct","recognition":{"type":"DirectHit","param":{"roi_offset":[0,0,0,0]}}}]}}`,
		string(encoded),
	)

	var decoded Recognition
	require.NoError(t, unmarshalJSON(encoded, &decoded))
	andParam, ok := decoded.Param.(*AndRecognitionParam)
	require.True(t, ok)
	require.Len(t, andParam.AllOf, 1)
	require.NotNil(t, andParam.AllOf[0].Inline)
	require.Equal(t, "direct", andParam.AllOf[0].Inline.SubName)
	require.Equal(t, RecognitionTypeDirectHit, andParam.AllOf[0].Inline.Type)

	or := RecOr(Inline(RecDirectHit(), "direct"))
	orEncoded, err := json.Marshal(or)
	require.NoError(t, err)
	require.JSONEq(t,
		`{"type":"Or","param":{"any_of":[{"sub_name":"direct","recognition":{"type":"DirectHit","param":{"roi_offset":[0,0,0,0]}}}]}}`,
		string(orEncoded),
	)
}

func TestPipelineV2DirectHitROIJSON(t *testing.T) {
	rect := Rect{10, 20, 30, 40}
	recognition := &Recognition{
		Type: RecognitionTypeDirectHit,
		Param: &DirectHitParam{
			ROI:       NewTargetRect(rect),
			ROIOffset: Rect{1, 2, 3, 4},
		},
	}

	encoded, err := marshalJSON(recognition)
	require.NoError(t, err)
	require.JSONEq(t, `{"type":"DirectHit","param":{"roi":[10,20,30,40],"roi_offset":[1,2,3,4]}}`, string(encoded))

	var decoded Recognition
	require.NoError(t, unmarshalJSON(encoded, &decoded))
	param, ok := decoded.Param.(*DirectHitParam)
	require.True(t, ok)
	gotROI, err := param.ROI.AsRect()
	require.NoError(t, err)
	require.Equal(t, rect, gotROI)
	require.Equal(t, Rect{1, 2, 3, 4}, param.ROIOffset)
}

func TestPipelineV2NeuralNetworkDetectThresholdJSON(t *testing.T) {
	threshold := []float64{0.25, 0}
	recognition := RecNeuralNetworkDetect(NeuralNetworkDetectParam{
		Model:     "detector.onnx",
		Threshold: threshold,
	})
	threshold[0] = 0.9
	param := recognition.Param.(*NeuralNetworkDetectParam)
	require.Equal(t, []float64{0.25, 0}, param.Threshold, "constructor must clone threshold slices")

	encoded, err := marshalJSON(recognition)
	require.NoError(t, err)
	require.JSONEq(t,
		`{"type":"NeuralNetworkDetect","param":{"model":"detector.onnx","roi_offset":[0,0,0,0],"threshold":[0.25,0]}}`,
		string(encoded),
	)

	var decoded Recognition
	require.NoError(t, unmarshalJSON(encoded, &decoded))
	decodedParam, ok := decoded.Param.(*NeuralNetworkDetectParam)
	require.True(t, ok)
	require.Equal(t, []float64{0.25, 0}, decodedParam.Threshold)
}

func TestPipelineV2NativePressureRoundTrip(t *testing.T) {
	tests := []struct {
		name               string
		build              func(*int) *Action
		get                func(ActionParam) *int
		inheritUsesDefault bool
	}{
		{
			name: "Click",
			build: func(pressure *int) *Action {
				return ActClick(ClickParam{Target: NewTargetRect(Rect{1, 2, 3, 4}), Pressure: pressure})
			},
			get: func(param ActionParam) *int {
				return param.(*ClickParam).Pressure
			},
		},
		{
			name: "LongPress",
			build: func(pressure *int) *Action {
				return ActLongPress(LongPressParam{
					Target:   NewTargetRect(Rect{1, 2, 3, 4}),
					Duration: 10 * time.Millisecond,
					Pressure: pressure,
				})
			},
			get: func(param ActionParam) *int {
				return param.(*LongPressParam).Pressure
			},
		},
		{
			name: "Swipe",
			build: func(pressure *int) *Action {
				return ActSwipe(SwipeParam{
					Begin:    NewTargetRect(Rect{1, 2, 3, 4}),
					End:      []Target{NewTargetRect(Rect{5, 6, 7, 8})},
					Duration: []time.Duration{10 * time.Millisecond},
					Pressure: pressure,
				})
			},
			get: func(param ActionParam) *int {
				return param.(*SwipeParam).Pressure
			},
		},
		{
			name: "MultiSwipe",
			build: func(pressure *int) *Action {
				return ActMultiSwipe(MultiSwipeItem{
					Begin:    NewTargetRect(Rect{1, 2, 3, 4}),
					End:      []Target{NewTargetRect(Rect{5, 6, 7, 8})},
					Duration: []time.Duration{10 * time.Millisecond},
					Pressure: pressure,
				})
			},
			get: func(param ActionParam) *int {
				return param.(*MultiSwipeParam).Swipes[0].Pressure
			},
			inheritUsesDefault: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := createResource(t)
			t.Cleanup(func() { require.NoError(t, res.Destroy()) })

			steps := []struct {
				name     string
				pressure *int
				want     int
			}{
				{name: "default", pressure: nil, want: 1},
				{name: "explicit zero", pressure: intPointer(0), want: 0},
				{name: "inherit zero", pressure: nil, want: 0},
				{name: "explicit nonzero", pressure: intPointer(37), want: 37},
				{name: "inherit nonzero", pressure: nil, want: 37},
			}
			for _, step := range steps {
				t.Run(step.name, func(t *testing.T) {
					require.NoError(t, res.OverridePipeline(map[string]*Node{
						"Pressure": NewNode("Pressure").SetAction(tc.build(step.pressure)),
					}))
					node, err := res.GetNode("Pressure")
					require.NoError(t, err)
					require.NotNil(t, node.Action)
					pressure := tc.get(node.Action.Param)
					require.NotNil(t, pressure)
					want := step.want
					if tc.inheritUsesDefault && step.pressure == nil && step.name != "default" {
						want = 1
					}
					require.Equal(t, want, *pressure)
				})
			}
		})
	}
}

func TestPipelineV2NativeShellTimeoutRoundTrip(t *testing.T) {
	res := createResource(t)
	t.Cleanup(func() { require.NoError(t, res.Destroy()) })

	steps := []struct {
		name    string
		timeout *time.Duration
		want    time.Duration
	}{
		{name: "default", timeout: nil, want: 20 * time.Second},
		{name: "explicit zero", timeout: durationPointer(0), want: 0},
		{name: "inherit zero", timeout: nil, want: 0},
		{name: "explicit finite", timeout: durationPointer(1500 * time.Millisecond), want: 1500 * time.Millisecond},
		{name: "inherit finite", timeout: nil, want: 1500 * time.Millisecond},
		{name: "infinite", timeout: durationPointer(-time.Millisecond), want: -time.Millisecond},
	}
	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			require.NoError(t, res.OverridePipeline(map[string]*Node{
				"ShellTimeout": NewNode("ShellTimeout").SetAction(&Action{
					Type: ActionTypeShell,
					Param: &ShellParam{
						Cmd:          "echo pipeline-v2",
						ShellTimeout: step.timeout,
					},
				}),
			}))
			node, err := res.GetNode("ShellTimeout")
			require.NoError(t, err)
			param, ok := node.Action.Param.(*ShellParam)
			require.True(t, ok)
			require.NotNil(t, param.ShellTimeout)
			require.Equal(t, step.want, *param.ShellTimeout)
		})
	}
}

func TestPipelineV2NativeAndOrInlineRoundTrip(t *testing.T) {
	res := createResource(t)
	t.Cleanup(func() { require.NoError(t, res.Destroy()) })

	pipeline := map[string]*Node{
		"AndInline": NewNode("AndInline").SetRecognition(
			RecAnd(Inline(&Recognition{
				Type: RecognitionTypeDirectHit,
				Param: &DirectHitParam{
					ROI:       NewTargetRect(Rect{10, 20, 30, 40}),
					ROIOffset: Rect{1, 2, 3, 4},
				},
			}, "direct")),
		),
		"OrInline": NewNode("OrInline").SetRecognition(
			RecOr(Inline(&Recognition{
				Type: RecognitionTypeDirectHit,
				Param: &DirectHitParam{
					ROI: NewTargetRect(Rect{50, 60, 70, 80}),
				},
			}, "direct")),
		),
	}
	require.NoError(t, res.OverridePipeline(pipeline))

	for _, name := range []string{"AndInline", "OrInline"} {
		t.Run(name, func(t *testing.T) {
			node, err := res.GetNode(name)
			require.NoError(t, err)
			require.NotNil(t, node.Recognition)
			switch node.Recognition.Type {
			case RecognitionTypeAnd:
				param, ok := node.Recognition.Param.(*AndRecognitionParam)
				require.True(t, ok)
				require.Len(t, param.AllOf, 1)
				inline := param.AllOf[0].Inline
				require.NotNil(t, inline)
				require.Equal(t, "direct", inline.SubName)
				direct, ok := inline.Param.(*DirectHitParam)
				require.True(t, ok)
				gotROI, err := direct.ROI.AsRect()
				require.NoError(t, err)
				require.Equal(t, Rect{10, 20, 30, 40}, gotROI)
				require.Equal(t, Rect{1, 2, 3, 4}, direct.ROIOffset)
			case RecognitionTypeOr:
				param, ok := node.Recognition.Param.(*OrRecognitionParam)
				require.True(t, ok)
				require.Len(t, param.AnyOf, 1)
				inline := param.AnyOf[0].Inline
				require.NotNil(t, inline)
				require.Equal(t, "direct", inline.SubName)
				direct, ok := inline.Param.(*DirectHitParam)
				require.True(t, ok)
				gotROI, err := direct.ROI.AsRect()
				require.NoError(t, err)
				require.Equal(t, Rect{50, 60, 70, 80}, gotROI)
			default:
				t.Fatalf("unexpected recognition type %q", node.Recognition.Type)
			}
		})
	}
}
