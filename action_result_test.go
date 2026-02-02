package maa

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testActionDetailFromActionAct struct {
	t *testing.T
}

func (a *testActionDetailFromActionAct) Run(ctx *Context, arg *CustomActionArg) bool {
	runAction := func(actionType NodeActionType, param NodeActionParam) *ActionDetail {
		detail, err := ctx.RunActionDirect(actionType, param, arg.Box, arg.RecognitionDetail)
		require.NoError(a.t, err)
		require.NotNil(a.t, detail)
		require.NotEmpty(a.t, detail.DetailJson)
		require.NotNil(a.t, detail.Result)

		requireActionResultMatchesRaw(a.t, detail)
		return detail
	}

	type testCase struct {
		name       string
		actionType NodeActionType
		param      NodeActionParam
		assert     func(t *testing.T, detail *ActionDetail)
	}

	testCases := []testCase{
		{
			name:       "click",
			actionType: NodeActionTypeClick,
			param: &NodeClickParam{
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
			actionType: NodeActionTypeLongPress,
			param: &NodeLongPressParam{
				Target:   NewTargetRect(Rect{120, 110, 10, 10}),
				Duration: (1500 * time.Millisecond).Milliseconds(),
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
			actionType: NodeActionTypeSwipe,
			param: &NodeSwipeParam{
				Begin:     NewTargetRect(Rect{100, 100, 10, 10}),
				End:       []Target{NewTargetRect(Rect{200, 200, 10, 10})},
				Duration:  []int64{300},
				EndHold:   []int64{50},
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
			actionType: NodeActionTypeMultiSwipe,
			param: &NodeMultiSwipeParam{
				Swipes: []NodeMultiSwipeItem{
					NewMultiSwipeItem(
						WithMultiSwipeItemStarting(0),
						WithMultiSwipeItemBegin(NewTargetRect(Rect{300, 300, 10, 10})),
						WithMultiSwipeItemEnd([]Target{NewTargetRect(Rect{320, 320, 10, 10})}),
						WithMultiSwipeItemDuration([]time.Duration{200 * time.Millisecond}),
						WithMultiSwipeItemEndHold([]time.Duration{20 * time.Millisecond}),
						WithMultiSwipeItemOnlyHover(true),
						WithMultiSwipeItemContact(0),
					),
					NewMultiSwipeItem(
						WithMultiSwipeItemStarting(100*time.Millisecond),
						WithMultiSwipeItemBegin(NewTargetRect(Rect{400, 400, 10, 10})),
						WithMultiSwipeItemEnd([]Target{NewTargetRect(Rect{420, 420, 10, 10})}),
						WithMultiSwipeItemDuration([]time.Duration{300 * time.Millisecond}),
						WithMultiSwipeItemEndHold([]time.Duration{40 * time.Millisecond}),
						WithMultiSwipeItemOnlyHover(false),
						WithMultiSwipeItemContact(1),
					),
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
			actionType: NodeActionTypeClickKey,
			param: &NodeClickKeyParam{
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
			actionType: NodeActionTypeKeyDown,
			param: &NodeKeyDownParam{
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
			actionType: NodeActionTypeKeyUp,
			param: &NodeKeyUpParam{
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
			actionType: NodeActionTypeLongPressKey,
			param: &NodeLongPressKeyParam{
				Key:      []int{24},
				Duration: (800 * time.Millisecond).Milliseconds(),
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				longPressKey, ok := detail.Result.AsLongPressKey()
				require.True(t, ok)
				require.NotNil(t, longPressKey)
			},
		},
		{
			name:       "input_text",
			actionType: NodeActionTypeInputText,
			param: &NodeInputTextParam{
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
			actionType: NodeActionTypeStartApp,
			param: &NodeStartAppParam{
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
			actionType: NodeActionTypeStopApp,
			param: &NodeStopAppParam{
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
			actionType: NodeActionTypeScroll,
			param: &NodeScrollParam{
				Dx: 120,
				Dy: -240,
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				scroll, ok := detail.Result.AsScroll()
				require.True(t, ok)
				require.NotNil(t, scroll)
			},
		},
		{
			name:       "touch_down",
			actionType: NodeActionTypeTouchDown,
			param: &NodeTouchDownParam{
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
			actionType: NodeActionTypeTouchMove,
			param: &NodeTouchMoveParam{
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
			actionType: NodeActionTypeTouchUp,
			param: &NodeTouchUpParam{
				Contact: 0,
			},
			assert: func(t *testing.T, detail *ActionDetail) {
				touchUp, ok := detail.Result.AsTouch()
				require.True(t, ok)
				require.NotNil(t, touchUp)
			},
		},
	}

	for _, tc := range testCases {
		a.t.Run(tc.name, func(t *testing.T) {
			detail := runAction(tc.actionType, tc.param)
			tc.assert(t, detail)
		})
	}

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
	ctrl := createCarouselImageController(t)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	defer res.Destroy()

	tasker := createTasker(t)
	defer tasker.Destroy()
	taskerBind(t, tasker, ctrl, res)

	err := res.RegisterCustomAction("TestActionDetail_ResultMatchesRawAct", &testActionDetailFromActionAct{t})
	require.NoError(t, err)

	pipeline := NewPipeline()
	testNode := NewNode("TestActionDetail_ResultMatchesRaw",
		WithAction(ActCustom("TestActionDetail_ResultMatchesRawAct")),
	)
	pipeline.AddNode(testNode)

	got := tasker.PostTask(testNode.Name, pipeline).
		Wait().Success()
	require.True(t, got)
}
