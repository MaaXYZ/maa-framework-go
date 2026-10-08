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
		`{"type":"And","param":{"all_of":[{"sub_name":"direct","recognition":{"type":"DirectHit","param":{}}}]}}`,
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
		`{"type":"Or","param":{"any_of":[{"sub_name":"direct","recognition":{"type":"DirectHit","param":{}}}]}}`,
		string(orEncoded),
	)
}

func TestPipelineV2DirectHitROIJSON(t *testing.T) {
	rect := Rect{10, 20, 30, 40}
	recognition := &Recognition{
		Type: RecognitionTypeDirectHit,
		Param: &DirectHitParam{
			ROI:       NewTargetRect(rect),
			ROIOffset: &Rect{1, 2, 3, 4},
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
	require.Equal(t, &Rect{1, 2, 3, 4}, param.ROIOffset)
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
		`{"type":"NeuralNetworkDetect","param":{"model":"detector.onnx","threshold":[0.25,0]}}`,
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
					Duration: durationPointer(10 * time.Millisecond),
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
					ROIOffset: &Rect{1, 2, 3, 4},
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
				require.Equal(t, &Rect{1, 2, 3, 4}, direct.ROIOffset)
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

func TestPipelineUnmarshalAnchorShorthand(t *testing.T) {
	t.Run("string and list forms resolve against the node name", func(t *testing.T) {
		var pipeline Pipeline
		require.NoError(t, json.Unmarshal([]byte(`{
			"A": {"anchor": "Anchor1"},
			"B": {"anchor": ["Anchor1", "Anchor2"]}
		}`), &pipeline))
		node, ok := pipeline.GetNode("A")
		require.True(t, ok)
		require.Equal(t, "A", node.Name)
		require.Equal(t, map[string]string{"Anchor1": "A"}, node.Anchor)
		node, ok = pipeline.GetNode("B")
		require.True(t, ok)
		require.Equal(t, "B", node.Name)
		require.Equal(t, map[string]string{"Anchor1": "B", "Anchor2": "B"}, node.Anchor)
	})

	t.Run("nodes without anchors still take their name from the key", func(t *testing.T) {
		var pipeline Pipeline
		require.NoError(t, json.Unmarshal([]byte(`{"Plain": {}}`), &pipeline))
		node, ok := pipeline.GetNode("Plain")
		require.True(t, ok)
		require.Equal(t, "Plain", node.Name)
		require.Nil(t, node.Anchor)
	})

	t.Run("object form preserves absent targets and empty target clears", func(t *testing.T) {
		var pipeline Pipeline
		require.NoError(t, json.Unmarshal([]byte(`{"N": {"anchor": {"A": "M", "B": ""}}}`), &pipeline))
		node, ok := pipeline.GetNode("N")
		require.True(t, ok)
		require.Equal(t, "N", node.Name)
		require.Equal(t, map[string]string{"A": "M", "B": ""}, node.Anchor)
		require.False(t, pipeline.HasNode("M"), "target existence is not validated during decoding")
	})

	t.Run("empty names lists and duplicates match the native parser", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			anchor string
			want   map[string]string
		}{
			{name: "empty string", anchor: `""`, want: map[string]string{"": "N"}},
			{name: "empty list", anchor: `[]`, want: map[string]string{}},
			{name: "duplicates and empty string", anchor: `["A", "", "A"]`, want: map[string]string{"A": "N", "": "N"}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				var pipeline Pipeline
				require.NoError(t, json.Unmarshal([]byte(`{"N": {"anchor": `+tc.anchor+`}}`), &pipeline))
				node, ok := pipeline.GetNode("N")
				require.True(t, ok)
				require.Equal(t, tc.want, node.Anchor)
			})
		}
	})

	t.Run("v1 flat nodes resolve anchors the same way and re-encode as v2", func(t *testing.T) {
		var pipeline Pipeline
		require.NoError(t, json.Unmarshal([]byte(`{
			"N": {"recognition": "DirectHit", "action": "Command", "exec": "sh", "args": "-c", "anchor": "A"}
		}`), &pipeline))
		node, ok := pipeline.GetNode("N")
		require.True(t, ok)
		require.Equal(t, map[string]string{"A": "N"}, node.Anchor)
		cmd, ok := node.Action.Param.(*CommandParam)
		require.True(t, ok)
		require.Equal(t, []string{"-c"}, cmd.Args)
		encoded, err := json.Marshal(&pipeline)
		require.NoError(t, err)
		require.JSONEq(t, `{"N": {
			"recognition": {"type": "DirectHit", "param": {}},
			"action": {"type": "Command", "param": {"exec": "sh", "args": ["-c"]}},
			"anchor": {"A": "N"}
		}}`, string(encoded), "re-encoding emits the object form")
	})

	t.Run("standalone Node still rejects the shorthand forms", func(t *testing.T) {
		for _, payload := range []string{`{"anchor": "A"}`, `{"anchor": ["A", "B"]}`} {
			node := Node{Name: "keep", Anchor: map[string]string{"A": "M"}}
			require.Error(t, json.Unmarshal([]byte(payload), &node))
			require.Equal(t, Node{Name: "keep", Anchor: map[string]string{"A": "M"}}, node)
		}
	})

	t.Run("null and malformed anchors are rejected with invariance", func(t *testing.T) {
		seedPipeline := func() *Pipeline {
			return NewPipeline().AddNode(NewNode("Keep").
				SetAnchor(map[string]string{"A": "M"}).
				SetRecognition(RecDirectHit()).
				SetAction(ActDoNothing()).
				SetNext([]NextItem{{Name: "Next", Anchor: true}}).
				SetEnabled(false).
				SetRateLimit(10 * time.Millisecond).
				SetFocus("keep focus").
				SetAttach(map[string]any{"keep": "value"}))
		}
		for name, payload := range map[string]string{
			"null anchor":   `{"N": {"anchor": null}}`,
			"number anchor": `{"N": {"anchor": 1}}`,
			"bool anchor":   `{"N": {"anchor": true}}`,
			"null entry":    `{"N": {"anchor": ["A", null]}}`,
			"number entry":  `{"N": {"anchor": ["A", 1]}}`,
			"bool entry":    `{"N": {"anchor": ["A", true]}}`,
			"object entry":  `{"N": {"anchor": ["A", {}]}}`,
			"array entry":   `{"N": {"anchor": ["A", []]}}`,
			"bad node":      `{"N": {"anchor": "A", "next": "X", "rate_limit": "fast"}}`,
			"not an object": `{"N": 5}`,
			"mixed nodes":   `{"Good": {"anchor": "A"}, "N": {"anchor": ["A", true]}}`,
		} {
			t.Run(name, func(t *testing.T) {
				seeded := seedPipeline()
				before := seedPipeline()
				originalNode, ok := seeded.GetNode("Keep")
				require.True(t, ok)
				require.Error(t, json.Unmarshal([]byte(payload), seeded), payload)
				require.Equal(t, before, seeded, "all state, including Name, must be preserved")
				node, ok := seeded.GetNode("Keep")
				require.True(t, ok)
				require.Same(t, originalNode, node)
			})
		}
	})

	t.Run("successful decoding replaces existing nodes", func(t *testing.T) {
		pipeline := NewPipeline().AddNode(NewNode("Keep").SetAnchor(map[string]string{"Old": "Keep"}))
		require.NoError(t, json.Unmarshal([]byte(`{"N": {"anchor": "A"}}`), pipeline))
		require.False(t, pipeline.HasNode("Keep"))
		require.Equal(t, 1, pipeline.Len())
		node, ok := pipeline.GetNode("N")
		require.True(t, ok)
		require.Equal(t, &Node{Name: "N", Anchor: map[string]string{"A": "N"}}, node)

		require.NoError(t, json.Unmarshal([]byte(`{"N": {"anchor": ["B", "C"]}, "Plain": {}}`), pipeline))
		require.Equal(t, 2, pipeline.Len())
		node, ok = pipeline.GetNode("N")
		require.True(t, ok)
		require.Equal(t, &Node{Name: "N", Anchor: map[string]string{"B": "N", "C": "N"}}, node)
		node, ok = pipeline.GetNode("Plain")
		require.True(t, ok)
		require.Equal(t, &Node{Name: "Plain"}, node)
	})

	t.Run("null decodes to an empty pipeline", func(t *testing.T) {
		var pipeline Pipeline
		require.NoError(t, json.Unmarshal([]byte(`null`), &pipeline))
		require.Equal(t, 0, pipeline.Len())
	})
}

func TestPipelineV2NativeAnchorShorthandRoundTrip(t *testing.T) {
	const payload = `{
		"String": {"anchor": "A"},
		"List": {"anchor": ["A", "B"]},
		"Object": {"anchor": {"A": "Missing", "B": ""}},
		"EmptyName": {"anchor": ""},
		"EmptyList": {"anchor": []},
		"Duplicate": {"anchor": ["A", "", "A"]},
		"Plain": {}
	}`
	var pipeline Pipeline
	require.NoError(t, json.Unmarshal([]byte(payload), &pipeline))

	rawResource := createResource(t)
	t.Cleanup(func() { require.NoError(t, rawResource.Destroy()) })
	typedResource := createResource(t)
	t.Cleanup(func() { require.NoError(t, typedResource.Destroy()) })
	require.NoError(t, rawResource.OverridePipeline(payload))
	require.NoError(t, typedResource.OverridePipeline(&pipeline))

	for _, name := range []string{"String", "List", "Object", "EmptyName", "EmptyList", "Duplicate", "Plain"} {
		t.Run(name, func(t *testing.T) {
			rawNode, err := rawResource.GetNode(name)
			require.NoError(t, err)
			typedNode, err := typedResource.GetNode(name)
			require.NoError(t, err)
			require.Equal(t, rawNode, typedNode, "typed normalization must preserve native behavior")
			decoded, ok := pipeline.GetNode(name)
			require.True(t, ok)
			require.Equal(t, name, decoded.Name)
			if len(decoded.Anchor) == 0 {
				require.Empty(t, rawNode.Anchor)
			} else {
				require.Equal(t, decoded.Anchor, rawNode.Anchor)
			}
		})
	}
}
