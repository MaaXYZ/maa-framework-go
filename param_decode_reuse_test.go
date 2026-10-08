package maa

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParamDecodeReuse_PreservesOmittedFields(t *testing.T) {
	boolPtr := func(value bool) *bool { return &value }
	cases := []struct {
		name     string
		newParam func() any
		input    string
		want     any
	}{
		{
			name:     "ClickKey",
			newParam: func() any { return &ClickKeyParam{Key: []int{65, 66}} },
			input:    `{"key":0}`,
			want:     &ClickKeyParam{Key: []int{0}},
		},
		{
			name:     "KeyDown",
			newParam: func() any { return &KeyDownParam{Key: 65, AutoUp: boolPtr(true)} },
			input:    `{"auto_up":false}`,
			want:     &KeyDownParam{Key: 65, AutoUp: boolPtr(false)},
		},
		{
			name:     "KeyUp",
			newParam: func() any { return &KeyUpParam{Key: 65, AutoUp: boolPtr(true)} },
			input:    `{"key":0}`,
			want:     &KeyUpParam{Key: 0, AutoUp: boolPtr(true)},
		},
		{
			name: "Command",
			newParam: func() any {
				return &CommandParam{Exec: "python3", Args: []string{"old.py"}, Detach: true}
			},
			input: `{"args":"new.py"}`,
			want:  &CommandParam{Exec: "python3", Args: []string{"new.py"}, Detach: true},
		},
		{
			name: "CustomAction",
			newParam: func() any {
				return &CustomActionParam{
					Target: NewTargetString("NodeA"), TargetOffset: Rect{1, 2, 3, 4},
					CustomAction: "old", CustomActionParam: map[string]any{"id": json.Number("9007199254740993")},
				}
			},
			input: `{"custom_action":"new"}`,
			want: &CustomActionParam{
				Target: NewTargetString("NodeA"), TargetOffset: Rect{1, 2, 3, 4},
				CustomAction: "new", CustomActionParam: map[string]any{"id": json.Number("9007199254740993")},
			},
		},
		{
			name: "ColorMatch",
			newParam: func() any {
				return &ColorMatchParam{
					ROI: NewTargetString("NodeA"), ROIOffset: Rect{1, 2, 3, 4}, Method: ColorMatchMethodHSV,
					Lower: [][]int{{10, 20, 30}}, Upper: [][]int{{40, 50, 60}}, Count: 7,
					OrderBy: ColorMatchOrderByScore, Index: 2, Connected: true,
				}
			},
			input: `{"count":0}`,
			want: &ColorMatchParam{
				ROI: NewTargetString("NodeA"), ROIOffset: Rect{1, 2, 3, 4}, Method: ColorMatchMethodHSV,
				Lower: [][]int{{10, 20, 30}}, Upper: [][]int{{40, 50, 60}}, Count: 0,
				OrderBy: ColorMatchOrderByScore, Index: 2, Connected: true,
			},
		},
		{
			name: "OCR",
			newParam: func() any {
				return &OCRParam{
					ROI: NewTargetString("NodeA"), ROIOffset: Rect{1, 2, 3, 4}, Expected: StringList{"old"},
					Threshold: 0.7, Replace: [][2]string{{"old", "fixed"}}, OrderBy: OCROrderByExpected,
					Index: 2, OnlyRec: true, Model: "custom", ColorFilter: "ColorNode",
				}
			},
			input: `{"threshold":0}`,
			want: &OCRParam{
				ROI: NewTargetString("NodeA"), ROIOffset: Rect{1, 2, 3, 4}, Expected: StringList{"old"},
				Threshold: 0, Replace: [][2]string{{"old", "fixed"}}, OrderBy: OCROrderByExpected,
				Index: 2, OnlyRec: true, Model: "custom", ColorFilter: "ColorNode",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, input := range []string{`{}`, `null`} {
				param := tc.newParam()
				require.NoError(t, json.Unmarshal([]byte(input), param))
				require.Equal(t, tc.newParam(), param, "input %s must preserve every existing field", input)
			}

			param := tc.newParam()
			require.NoError(t, json.Unmarshal([]byte(tc.input), param))
			require.Equal(t, tc.want, param, "provided fields must replace existing values while omitted fields are preserved")
		})
	}
}

func TestParamDecodeReuse_Aliases(t *testing.T) {
	for _, newParam := range []struct {
		name string
		new  func() any
		key  func(any) int
	}{
		{"KeyDown", func() any { return &KeyDownParam{Key: 65} }, func(p any) int { return p.(*KeyDownParam).Key }},
		{"KeyUp", func() any { return &KeyUpParam{Key: 65} }, func(p any) int { return p.(*KeyUpParam).Key }},
	} {
		for _, tc := range []struct {
			name  string
			input string
			want  int
		}{
			{"alias replaces old key", `{"key_code":66}`, 66},
			{"canonical key wins", `{"key_code":66,"key":0}`, 0},
		} {
			t.Run(newParam.name+"/"+tc.name, func(t *testing.T) {
				param := newParam.new()
				require.NoError(t, json.Unmarshal([]byte(tc.input), param))
				require.Equal(t, tc.want, newParam.key(param))
			})
		}
	}

	for _, tc := range []struct {
		name  string
		input string
		want  StringList
	}{
		{"text replaces seeded expected", `{"text":"new"}`, StringList{"new"}},
		{"expected wins over text", `{"text":"alias","expected":"canonical"}`, StringList{"canonical"}},
		{"empty expected wins over text", `{"text":"alias","expected":[]}`, StringList{}},
		{"canonical expected ignores invalid text", `{"expected":"canonical","text":7}`, StringList{"canonical"}},
		{"text null clears seeded expected", `{"text":null}`, nil},
	} {
		t.Run("OCR/"+tc.name, func(t *testing.T) {
			param := OCRParam{Expected: StringList{"old"}, Model: "custom"}
			require.NoError(t, json.Unmarshal([]byte(tc.input), &param))
			require.Equal(t, tc.want, param.Expected)
			require.Equal(t, "custom", param.Model)
		})
	}
}

func TestParamDecodeReuse_ErrorPreservesReceiverAndSharedValues(t *testing.T) {
	cases := []struct {
		name  string
		new   func() (param, shared any)
		input string
	}{
		{
			name: "ClickKey",
			new: func() (any, any) {
				keys := []int{65, 66}
				return &ClickKeyParam{Key: keys}, keys
			},
			input: `{"key":[67,"bad"]}`,
		},
		{
			name: "KeyDown",
			new: func() (any, any) {
				autoUp := true
				return &KeyDownParam{Key: 65, AutoUp: &autoUp}, &autoUp
			},
			input: `{"auto_up":false,"key":"bad"}`,
		},
		{
			name: "KeyUp",
			new: func() (any, any) {
				autoUp := true
				return &KeyUpParam{Key: 65, AutoUp: &autoUp}, &autoUp
			},
			input: `{"auto_up":false,"key_code":"bad"}`,
		},
		{
			name: "Command",
			new: func() (any, any) {
				args := []string{"old.py"}
				return &CommandParam{Exec: "python3", Args: args, Detach: true}, args
			},
			input: `{"args":["new.py"],"detach":false,"exec":7}`,
		},
		{
			name: "CustomAction",
			new: func() (any, any) {
				custom := map[string]any{"id": json.Number("9007199254740993")}
				return &CustomActionParam{
					Target: NewTargetString("NodeA"), TargetOffset: Rect{1, 2, 3, 4},
					CustomAction: "old", CustomActionParam: custom,
				}, custom
			},
			input: `{"custom_action":"new","target":true,"custom_action_param":{"id":1},"target_offset":[1,2,3]}`,
		},
		{
			name: "ColorMatch",
			new: func() (any, any) {
				lower, upper := [][]int{{10, 20, 30}}, [][]int{{40, 50, 60}}
				return &ColorMatchParam{
					ROI: NewTargetString("NodeA"), ROIOffset: Rect{1, 2, 3, 4}, Method: ColorMatchMethodHSV,
					Lower: lower, Upper: upper, Count: 7, OrderBy: ColorMatchOrderByScore, Index: 2, Connected: true,
				}, [][][]int{lower, upper}
			},
			input: `{"roi":true,"lower":[1,2,3],"count":0,"upper":[[4,"bad",6]]}`,
		},
		{
			name: "OCR",
			new: func() (any, any) {
				expected, replace := StringList{"old"}, [][2]string{{"old", "fixed"}}
				return &OCRParam{
					ROI: NewTargetString("NodeA"), ROIOffset: Rect{1, 2, 3, 4}, Expected: expected,
					Threshold: 0.7, Replace: replace, OrderBy: OCROrderByExpected,
					Index: 2, OnlyRec: true, Model: "custom", ColorFilter: "ColorNode",
				}, []any{expected, replace}
			},
			input: `{"roi":true,"expected":"new","replace":["new","fixed"],"model":"new","threshold":"bad"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			param, shared := tc.new()
			wantParam, wantShared := tc.new()
			require.Error(t, json.Unmarshal([]byte(tc.input), param))
			require.Equal(t, wantParam, param, "a decode error must preserve all existing fields")
			require.Equal(t, wantShared, shared, "a decode error must preserve data shared with the caller")
		})
	}
}

func TestParamDecodeReuse_NullSemantics(t *testing.T) {
	for _, tc := range []struct {
		name string
		new  func() any
		want any
	}{
		{"KeyDown AutoUp", func() any { autoUp := true; return &KeyDownParam{Key: 65, AutoUp: &autoUp} }, &KeyDownParam{Key: 65}},
		{"KeyUp AutoUp", func() any { autoUp := true; return &KeyUpParam{Key: 65, AutoUp: &autoUp} }, &KeyUpParam{Key: 65}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			param := tc.new()
			require.NoError(t, json.Unmarshal([]byte(`{"auto_up":null}`), param))
			require.Equal(t, tc.want, param)
		})
	}
	t.Run("omitted AutoUp stays nil", func(t *testing.T) {
		param := KeyDownParam{Key: 65}
		require.NoError(t, json.Unmarshal([]byte(`{"key_code":66}`), &param))
		require.Equal(t, KeyDownParam{Key: 66}, param)
	})
	t.Run("expected null stays invalid", func(t *testing.T) {
		param := OCRParam{Expected: StringList{"old"}, Model: "custom"}
		require.Error(t, json.Unmarshal([]byte(`{"model":"new","expected":null,"text":"alias"}`), &param))
		require.Equal(t, OCRParam{Expected: StringList{"old"}, Model: "custom"}, param)
	})
	t.Run("custom parameter null clears only that field", func(t *testing.T) {
		param := CustomActionParam{CustomAction: "old", CustomActionParam: map[string]any{"id": json.Number("1")}}
		require.NoError(t, json.Unmarshal([]byte(`{"custom_action_param":null}`), &param))
		require.Equal(t, CustomActionParam{CustomAction: "old"}, param)
	})
}

func TestParamDecodeReuse_CustomActionParameterReplacement(t *testing.T) {
	shared := map[string]any{"old": json.Number("1")}
	param := CustomActionParam{CustomAction: "handler", CustomActionParam: shared}
	require.NoError(t, json.Unmarshal([]byte(`{"custom_action_param":{"id":9007199254740993}}`), &param))
	require.Equal(t, CustomActionParam{
		CustomAction: "handler", CustomActionParam: map[string]any{"id": json.Number("9007199254740993")},
	}, param)
	require.Equal(t, map[string]any{"old": json.Number("1")}, shared)
}

func TestParamDecodeReuse_InvalidOCRAliasPreservesReceiver(t *testing.T) {
	param := OCRParam{Expected: StringList{"old"}, Threshold: 0.7, Model: "custom"}
	require.Error(t, json.Unmarshal([]byte(`{"text":7,"threshold":0.2,"model":"new"}`), &param))
	require.Equal(t, OCRParam{Expected: StringList{"old"}, Threshold: 0.7, Model: "custom"}, param)
}
