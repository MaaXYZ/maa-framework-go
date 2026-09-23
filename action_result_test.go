package maa

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testActionDetailFromActionAct struct {
	results chan []testActionDetailResult
}

type testActionDetailResult struct {
	name   string
	detail *ActionDetail
	err    error
	assert func(*testing.T, *ActionDetail)
}

func (a *testActionDetailFromActionAct) Run(ctx *Context, arg *CustomActionArg) bool {
	type testCase struct {
		name       string
		actionType ActionType
		param      ActionParam
		assert     func(t *testing.T, detail *ActionDetail)
	}

	testCases := []testCase{
		{
			name:       "click",
			actionType: ActionTypeClick,
			param: &ClickParam{
				Target:  NewTargetRect(Rect{100, 100, 10, 10}),
				Contact: 1,
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				click, ok := detail.Result.AsClick()
				require.True(t, ok)
				require.NotNil(t, click)
			},
		},
		{
			name:       "long_press",
			actionType: ActionTypeLongPress,
			param: &LongPressParam{
				Target:   NewTargetRect(Rect{120, 110, 10, 10}),
				Duration: 1500 * time.Millisecond,
				Contact:  2,
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				longPress, ok := detail.Result.AsLongPress()
				require.True(t, ok)
				require.NotNil(t, longPress)
			},
		},
		{
			name:       "swipe",
			actionType: ActionTypeSwipe,
			param: &SwipeParam{
				Begin:     NewTargetRect(Rect{100, 100, 10, 10}),
				End:       []Target{NewTargetRect(Rect{200, 200, 10, 10})},
				Duration:  []time.Duration{300 * time.Millisecond},
				EndHold:   []time.Duration{50 * time.Millisecond},
				OnlyHover: true,
				Contact:   1,
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				swipe, ok := detail.Result.AsSwipe()
				require.True(t, ok)
				require.NotNil(t, swipe)
			},
		},
		{
			name:       "multi_swipe",
			actionType: ActionTypeMultiSwipe,
			param: &MultiSwipeParam{
				Swipes: []MultiSwipeItem{
					{
						Starting:  0,
						Begin:     NewTargetRect(Rect{300, 300, 10, 10}),
						End:       []Target{NewTargetRect(Rect{320, 320, 10, 10})},
						Duration:  []time.Duration{200 * time.Millisecond},
						EndHold:   []time.Duration{20 * time.Millisecond},
						OnlyHover: true,
						Contact:   0,
					},
					{
						Starting:  100 * time.Millisecond,
						Begin:     NewTargetRect(Rect{400, 400, 10, 10}),
						End:       []Target{NewTargetRect(Rect{420, 420, 10, 10})},
						Duration:  []time.Duration{300 * time.Millisecond},
						EndHold:   []time.Duration{40 * time.Millisecond},
						OnlyHover: false,
						Contact:   1,
					},
				},
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				multiSwipe, ok := detail.Result.AsMultiSwipe()
				require.True(t, ok)
				require.NotNil(t, multiSwipe)
			},
		},
		{
			name:       "click_key",
			actionType: ActionTypeClickKey,
			param: &ClickKeyParam{
				Key: []int{4, 5},
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				clickKey, ok := detail.Result.AsClickKey()
				require.True(t, ok)
				require.NotNil(t, clickKey)
			},
		},
		{
			name:       "key_down",
			actionType: ActionTypeKeyDown,
			param: &KeyDownParam{
				Key: 4,
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				keyDown, ok := detail.Result.AsClickKey()
				require.True(t, ok)
				require.NotNil(t, keyDown)
			},
		},
		{
			name:       "key_up",
			actionType: ActionTypeKeyUp,
			param: &KeyUpParam{
				Key: 4,
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				keyUp, ok := detail.Result.AsClickKey()
				require.True(t, ok)
				require.NotNil(t, keyUp)
			},
		},
		{
			name:       "long_press_key",
			actionType: ActionTypeLongPressKey,
			param: &LongPressKeyParam{
				Key:      []int{24},
				Duration: 800 * time.Millisecond,
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				longPressKey, ok := detail.Result.AsLongPressKey()
				require.True(t, ok)
				require.NotNil(t, longPressKey)
			},
		},
		{
			name:       "input_text",
			actionType: ActionTypeInputText,
			param: &InputTextParam{
				InputText: "Hello",
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				inputText, ok := detail.Result.AsInputText()
				require.True(t, ok)
				require.NotNil(t, inputText)
			},
		},
		{
			name:       "start_app",
			actionType: ActionTypeStartApp,
			param: &StartAppParam{
				Package: "com.android.settings",
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				startApp, ok := detail.Result.AsApp()
				require.True(t, ok)
				require.NotNil(t, startApp)
			},
		},
		{
			name:       "stop_app",
			actionType: ActionTypeStopApp,
			param: &StopAppParam{
				Package: "com.android.settings",
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				stopApp, ok := detail.Result.AsApp()
				require.True(t, ok)
				require.NotNil(t, stopApp)
			},
		},
		{
			name:       "scroll",
			actionType: ActionTypeScroll,
			param: &ScrollParam{
				Target: NewTargetRect(Rect{100, 100, 10, 10}),
				Dx:     120,
				Dy:     -240,
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				scroll, ok := detail.Result.AsScroll()
				require.True(t, ok)
				require.NotNil(t, scroll)
			},
		},
		{
			name:       "screencap",
			actionType: ActionTypeScreencap,
			param: &ScreencapParam{
				Filename: "test_capture",
				Format:   "jpg",
				Quality:  85,
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				screencap, ok := detail.Result.AsScreencap()
				require.True(t, ok)
				require.NotNil(t, screencap)
				require.NotEmpty(t, screencap.Filepath)
				require.Equal(t, "jpg", screencap.Format)
				require.Equal(t, 85, screencap.Quality)
				require.True(t, screencap.Success)
			},
		},
		{
			name:       "touch_down",
			actionType: ActionTypeTouchDown,
			param: &TouchDownParam{
				Target:   NewTargetRect(Rect{50, 60, 10, 10}),
				Pressure: 500,
				Contact:  0,
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				touchDown, ok := detail.Result.AsTouch()
				require.True(t, ok)
				require.NotNil(t, touchDown)
			},
		},
		{
			name:       "touch_move",
			actionType: ActionTypeTouchMove,
			param: &TouchMoveParam{
				Target:   NewTargetRect(Rect{70, 80, 10, 10}),
				Pressure: 700,
				Contact:  0,
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				touchMove, ok := detail.Result.AsTouch()
				require.True(t, ok)
				require.NotNil(t, touchMove)
			},
		},
		{
			name:       "touch_up",
			actionType: ActionTypeTouchUp,
			param: &TouchUpParam{
				Contact: 0,
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				touchUp, ok := detail.Result.AsTouch()
				require.True(t, ok)
				require.NotNil(t, touchUp)
			},
		},
	}

	results := make([]testActionDetailResult, 0, len(testCases))
	for _, tc := range testCases {
		detail, err := ctx.RunActionDirect(tc.actionType, tc.param, arg.Box, arg.RecognitionDetail)
		results = append(results, testActionDetailResult{tc.name, detail, err, tc.assert})
	}
	a.results <- results

	return true
}

func requireActionResultMatchesRaw(t *testing.T, detail *ActionDetail) {
	t.Helper()
	require.NotNil(t, detail)
	require.NotNil(t, detail.Result)

	resultJSON, err := json.Marshal(detail.Result.Value())
	require.NoError(t, err)
	resultMap := map[string]any{}
	require.NoError(t, json.Unmarshal(resultJSON, &resultMap))

	rawDetail := map[string]any{}
	require.NoError(t, json.Unmarshal([]byte(detail.DetailJson), &rawDetail))

	for key, rawVal := range rawDetail {
		resultVal, ok := resultMap[key]
		require.True(t, ok, "result missing key: %s", key)
		require.Equal(t, rawVal, resultVal)
	}
}

func TestActionDetail_ResultMatchesRaw(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	require.True(t, ctrl.PostScreencap().Wait().Success())

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	resultsCh := make(chan []testActionDetailResult, 1)
	err := res.RegisterCustomAction("TestActionDetail_ResultMatchesRawAct", &testActionDetailFromActionAct{resultsCh})
	require.NoError(t, err)

	pipeline := NewPipeline()
	testNode := NewNode("TestActionDetail_ResultMatchesRaw").
		SetAction(ActCustom(CustomActionParam{CustomAction: "TestActionDetail_ResultMatchesRawAct"}))
	pipeline.AddNode(testNode)

	got := tasker.PostTask(testNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)

	select {
	case results := <-resultsCh:
		for _, result := range results {
			t.Run(result.name, func(t *testing.T) {
				require.NoError(t, result.err)
				require.NotNil(t, result.detail)
				require.NotEmpty(t, result.detail.DetailJson)
				require.NotNil(t, result.detail.Result)
				requireActionResultMatchesRaw(t, result.detail)
				result.assert(t, result.detail)
			})
		}
	default:
		t.Fatal("custom action callback was not called")
	}
}

func TestParseActionResultNullDetail(t *testing.T) {
	result, err := parseActionResult(string(ActionTypeSwipe), "null")
	require.NoError(t, err)
	require.Nil(t, result)
}
