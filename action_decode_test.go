package maa

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAction_UnmarshalJSON_KnownParamErrors(t *testing.T) {
	cases := []struct {
		name string
		data string
	}{
		{"action type must be a string", `{"type":5,"param":{}}`},
		{"swipe duration must be an array", `{"type":"Swipe","param":{"duration":"bad"}}`},
		{"multi swipe starting must be a number", `{"type":"MultiSwipe","param":{"swipes":[{"starting":"bad"}]}}`},
		{"long press key duration must be a number", `{"type":"LongPressKey","param":{"duration":"bad"}}`},
		{"shell timeout must be a number", `{"type":"Shell","param":{"shell_timeout":"bad"}}`},
		{"offset must be an array", `{"type":"Click","param":{"target_offset":"bad"}}`},
		{"offset must have 2 or 4 elements", `{"type":"Click","param":{"target_offset":[1,2,3]}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			action := Action{Type: ActionTypeClick, Param: &ClickParam{Contact: 7}}
			require.Error(t, unmarshalJSON([]byte(tc.data), &action))
			require.Equal(t, ActionTypeClick, action.Type)
			param, ok := action.Param.(*ClickParam)
			require.True(t, ok, "malformed known params must not fall back to raw, got %T", action.Param)
			require.Equal(t, 7, param.Contact)
		})
	}
}

func TestActionParam_OffsetJSONOmission(t *testing.T) {
	target := NewTargetRect(Rect{100, 200, 10, 10})
	params := []struct {
		name      string
		offsetKey string
		zero      any
		nonZero   any
	}{
		{"ClickParam", "target_offset", ClickParam{Target: target}, ClickParam{Target: target, TargetOffset: Rect{1, 2, 0, 0}}},
		{"LongPressParam", "target_offset", LongPressParam{Target: target}, LongPressParam{Target: target, TargetOffset: Rect{1, 2, 0, 0}}},
		{"SwipeParam", "begin_offset", SwipeParam{Begin: target}, SwipeParam{Begin: target, BeginOffset: Rect{1, 2, 0, 0}}},
		{"MultiSwipeItem", "begin_offset", MultiSwipeItem{Begin: target}, MultiSwipeItem{Begin: target, BeginOffset: Rect{1, 2, 0, 0}}},
		{"TouchDownParam", "target_offset", TouchDownParam{Target: target}, TouchDownParam{Target: target, TargetOffset: Rect{1, 2, 0, 0}}},
		{"TouchMoveParam", "target_offset", TouchMoveParam{Target: target}, TouchMoveParam{Target: target, TargetOffset: Rect{1, 2, 0, 0}}},
		{"ScrollParam", "target_offset", ScrollParam{Target: target}, ScrollParam{Target: target, TargetOffset: Rect{1, 2, 0, 0}}},
		{"CustomActionParam", "target_offset", CustomActionParam{Target: target, CustomAction: "act"}, CustomActionParam{Target: target, TargetOffset: Rect{1, 2, 0, 0}, CustomAction: "act"}},
	}
	for _, p := range params {
		t.Run(p.name, func(t *testing.T) {
			encoded, err := marshalJSON(p.zero)
			require.NoError(t, err)
			var wire map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(encoded, &wire))
			_, ok := wire[p.offsetKey]
			require.False(t, ok, "zero offset must omit %s, got %s", p.offsetKey, encoded)

			encoded, err = marshalJSON(p.nonZero)
			require.NoError(t, err)
			wire = nil
			require.NoError(t, json.Unmarshal(encoded, &wire))
			require.Contains(t, wire, p.offsetKey)
			var elements []int
			require.NoError(t, json.Unmarshal(wire[p.offsetKey], &elements))
			require.Len(t, elements, 4)
			require.Equal(t, []int{1, 2, 0, 0}, elements)
		})
	}
}

func TestAction_OffsetDecodedPointReencodesAsRect(t *testing.T) {
	var action Action
	require.NoError(t, unmarshalJSON([]byte(`{"type":"Click","param":{"target_offset":[5,10]}}`), &action))
	param, ok := action.Param.(*ClickParam)
	require.True(t, ok)
	require.Equal(t, Rect{5, 10, 1, 1}, param.TargetOffset)

	encoded, err := marshalJSON(action)
	require.NoError(t, err)
	require.JSONEq(t, `{"type":"Click","param":{"target_offset":[5,10,1,1]}}`, string(encoded))
}

func TestKeyParam_UnmarshalJSON_KeyCodeAlias(t *testing.T) {
	cases := []struct {
		name    string
		param   string
		wantKey int
	}{
		{"key still decodes", `{"key":32}`, 32},
		{"key_code alias decodes", `{"key_code":65}`, 65},
		{"key wins over key_code", `{"key":1,"key_code":2}`, 1},
		{"key wins regardless of document order", `{"key_code":2,"key":1}`, 1},
		{"neither present stays zero", `{}`, 0},
	}
	for _, actionType := range []ActionType{ActionTypeKeyDown, ActionTypeKeyUp} {
		for _, tc := range cases {
			t.Run(string(actionType)+"/"+tc.name, func(t *testing.T) {
				var action Action
				input := fmt.Sprintf(`{"type":%q,"param":%s}`, actionType, tc.param)
				require.NoError(t, unmarshalJSON([]byte(input), &action))

				key := -1
				switch p := action.Param.(type) {
				case *KeyDownParam:
					key = p.Key
				case *KeyUpParam:
					key = p.Key
				default:
					t.Fatalf("unexpected parameter type %T", action.Param)
				}
				require.Equal(t, tc.wantKey, key)

				wantParam := fmt.Sprintf(`{"key":%d}`, tc.wantKey)
				if tc.wantKey == 0 {
					wantParam = `{}`
				}
				encoded, err := marshalJSON(action)
				require.NoError(t, err)
				require.JSONEq(t, fmt.Sprintf(`{"type":%q,"param":%s}`, actionType, wantParam), string(encoded))
			})
		}
	}
}

func TestKeyParam_UnmarshalJSON_AutoUpWithAlias(t *testing.T) {
	var action Action
	require.NoError(t, unmarshalJSON([]byte(`{"type":"KeyDown","param":{"key_code":65,"auto_up":true}}`), &action))
	param, ok := action.Param.(*KeyDownParam)
	require.True(t, ok)
	require.Equal(t, 65, param.Key)
	require.NotNil(t, param.AutoUp)
	require.True(t, *param.AutoUp)
}

func TestKeyParam_UnmarshalJSON_InvalidKeys(t *testing.T) {
	for _, input := range []string{
		`{"key":"bad"}`, `{"key_code":"bad"}`, `{"key":null}`, `{"key_code":null}`, `{"key_code":[65]}`,
	} {
		keyDown := KeyDownParam{Key: 7}
		got := keyDown
		require.Error(t, unmarshalJSON([]byte(input), &got), "input %s", input)
		require.Equal(t, keyDown, got)

		keyUp := KeyUpParam{Key: 7}
		gotUp := keyUp
		require.Error(t, unmarshalJSON([]byte(input), &gotUp), "input %s", input)
		require.Equal(t, keyUp, gotUp)
	}
}

func TestSwipeActionResult_EndRoundTrip(t *testing.T) {
	encodedEnd := func(t *testing.T, result *SwipeActionResult) string {
		t.Helper()
		encoded, err := json.Marshal(result)
		require.NoError(t, err)
		var wire swipeActionResultWire
		require.NoError(t, json.Unmarshal(encoded, &wire))
		return string(wire.End)
	}

	t.Run("decoded end re-encodes verbatim", func(t *testing.T) {
		var result SwipeActionResult
		require.NoError(t, json.Unmarshal([]byte(`{"begin":[1,2],"end":[[3,4],[5,6]]}`), &result))
		require.Equal(t, []Point{{3, 4}, {5, 6}}, result.End)
		require.JSONEq(t, "[[3,4],[5,6]]", encodedEnd(t, &result))
	})

	t.Run("single point end form is preserved verbatim", func(t *testing.T) {
		var result SwipeActionResult
		require.NoError(t, json.Unmarshal([]byte(`{"end":[3,4]}`), &result))
		require.Equal(t, []Point{{3, 4}}, result.End)
		require.JSONEq(t, "[3,4]", encodedEnd(t, &result))
	})

	t.Run("constructed result encodes end as a point list", func(t *testing.T) {
		result := SwipeActionResult{End: []Point{{3, 4}, {5, 6}}}
		require.JSONEq(t, "[[3,4],[5,6]]", encodedEnd(t, &result))
	})

	t.Run("malformed end is rejected", func(t *testing.T) {
		var result SwipeActionResult
		require.Error(t, json.Unmarshal([]byte(`{"end":[["bad"]]}`), &result))
		require.Nil(t, result.End)
	})
}

func TestActionResult_AsAccessorMatrix(t *testing.T) {
	details := []struct {
		action ActionType
		detail string
	}{
		{ActionTypeClick, `{"point":[1,2],"contact":0,"pressure":0}`},
		{ActionTypeLongPress, `{"point":[1,2],"duration":1000,"contact":0,"pressure":0}`},
		{ActionTypeSwipe, `{"begin":[1,2],"end":[[3,4]],"duration":[200]}`},
		{ActionTypeMultiSwipe, `{"swipes":[{"begin":[1,2],"end":[[3,4]],"duration":[200]}]}`},
		{ActionTypeClickKey, `{"keycode":[27],"auto_up":false}`},
		{ActionTypeKeyDown, `{"keycode":[27],"auto_up":false}`},
		{ActionTypeKeyUp, `{"keycode":[27],"auto_up":false}`},
		{ActionTypeLongPressKey, `{"keycode":[27],"duration":1000}`},
		{ActionTypeInputText, `{"text":"hi"}`},
		{ActionTypeStartApp, `{"package":"com.example.app"}`},
		{ActionTypeStopApp, `{"package":"com.example.app"}`},
		{ActionTypeScroll, `{"point":[100,200],"dx":3,"dy":-3}`},
		{ActionTypeTouchDown, `{"contact":0,"point":[1,2],"pressure":0,"auto_up":false}`},
		{ActionTypeTouchMove, `{"contact":0,"point":[1,2],"pressure":0,"auto_up":false}`},
		{ActionTypeTouchUp, `{"contact":0,"point":[1,2],"pressure":0,"auto_up":false}`},
		{ActionTypeShell, `{"cmd":"echo hi","shell_timeout":20000,"success":true,"output":"hi"}`},
		{ActionTypeScreencap, `{"filepath":"x.png","format":"png","quality":100,"success":true}`},
	}

	results := map[ActionType]*ActionResult{}
	for _, d := range details {
		result, err := parseActionResult(string(d.action), d.detail)
		require.NoError(t, err, "action %s", d.action)
		require.NotNil(t, result)
		require.Equal(t, d.action, result.Type())
		results[d.action] = result
	}

	accessors := []struct {
		name   string
		call   func(*ActionResult) (any, bool)
		accept []ActionType
	}{
		{"AsClick", func(r *ActionResult) (any, bool) { return r.AsClick() }, []ActionType{ActionTypeClick}},
		{"AsLongPress", func(r *ActionResult) (any, bool) { return r.AsLongPress() }, []ActionType{ActionTypeLongPress}},
		{"AsSwipe", func(r *ActionResult) (any, bool) { return r.AsSwipe() }, []ActionType{ActionTypeSwipe}},
		{"AsMultiSwipe", func(r *ActionResult) (any, bool) { return r.AsMultiSwipe() }, []ActionType{ActionTypeMultiSwipe}},
		{"AsClickKey", func(r *ActionResult) (any, bool) { return r.AsClickKey() }, []ActionType{ActionTypeClickKey, ActionTypeKeyDown, ActionTypeKeyUp}},
		{"AsLongPressKey", func(r *ActionResult) (any, bool) { return r.AsLongPressKey() }, []ActionType{ActionTypeLongPressKey}},
		{"AsInputText", func(r *ActionResult) (any, bool) { return r.AsInputText() }, []ActionType{ActionTypeInputText}},
		{"AsApp", func(r *ActionResult) (any, bool) { return r.AsApp() }, []ActionType{ActionTypeStartApp, ActionTypeStopApp}},
		{"AsScroll", func(r *ActionResult) (any, bool) { return r.AsScroll() }, []ActionType{ActionTypeScroll}},
		{"AsTouch", func(r *ActionResult) (any, bool) { return r.AsTouch() }, []ActionType{ActionTypeTouchDown, ActionTypeTouchMove, ActionTypeTouchUp}},
		{"AsShell", func(r *ActionResult) (any, bool) { return r.AsShell() }, []ActionType{ActionTypeShell}},
		{"AsScreencap", func(r *ActionResult) (any, bool) { return r.AsScreencap() }, []ActionType{ActionTypeScreencap}},
	}

	for _, acc := range accessors {
		t.Run(acc.name, func(t *testing.T) {
			for _, d := range details {
				val, ok := acc.call(results[d.action])
				if slices.Contains(acc.accept, d.action) {
					require.True(t, ok, "%s should accept %s", acc.name, d.action)
					require.NotNil(t, val)
				} else {
					require.False(t, ok, "%s should reject %s", acc.name, d.action)
					require.Nil(t, val)
				}
			}
		})
	}
}

func TestParseActionResult_Errors(t *testing.T) {
	_, err := parseActionResult("FutureAction", `{"value":1}`)
	require.ErrorContains(t, err, "unknown action result type")

	_, err = parseActionResult(string(ActionTypeClick), `{"point":"bad"}`)
	require.Error(t, err, "malformed detail must be rejected")
}

func TestPoint_UnmarshalJSON_Forms(t *testing.T) {
	t.Run("array and string forms", func(t *testing.T) {
		var fromArray Point
		require.NoError(t, json.Unmarshal([]byte(`[1,2]`), &fromArray))
		require.Equal(t, Point{1, 2}, fromArray)
		require.Equal(t, 1, fromArray.X())
		require.Equal(t, 2, fromArray.Y())

		var fromString Point
		require.NoError(t, json.Unmarshal([]byte(`"[1, 2]"`), &fromString))
		require.Equal(t, Point{1, 2}, fromString)
	})

	t.Run("invalid forms are rejected", func(t *testing.T) {
		for _, input := range []string{
			`[1]`, `[1,2,3]`, `["a","b"]`, `5`, `"abc"`, `"[1]"`, `"[1,2,3]"`,
		} {
			var p Point
			require.Error(t, json.Unmarshal([]byte(input), &p), "input %s", input)
		}
	})
}

func TestParseSwipeEndPoints_Forms(t *testing.T) {
	t.Run("accepted forms", func(t *testing.T) {
		cases := []struct {
			input string
			want  []Point
		}{
			{`[[1,2],[3,4]]`, []Point{{1, 2}, {3, 4}}},
			{`[1,2]`, []Point{{1, 2}}},
			{`"[1, 2]"`, []Point{{1, 2}}},
			{`"[[1,2],[3,4]]"`, []Point{{1, 2}, {3, 4}}},
			{`[]`, []Point{}},
		}
		for _, tc := range cases {
			points, err := parseSwipeEndPoints(json.RawMessage(tc.input))
			require.NoError(t, err, "input %s", tc.input)
			require.Equal(t, tc.want, points)
		}
	})

	t.Run("invalid forms are rejected", func(t *testing.T) {
		for _, input := range []string{
			`[1]`, `[1,2,3]`, `[[1,2],[3]]`, `[[1,"a"]]`, `["a","b"]`, `3`, `"abc"`, `"[1]"`,
		} {
			_, err := parseSwipeEndPoints(json.RawMessage(input))
			require.Error(t, err, "input %s", input)
		}
	})
}
