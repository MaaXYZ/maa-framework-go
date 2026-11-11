package maa

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomRecognitionParam_IndexJSON(t *testing.T) {
	for _, index := range []int{0, 2, -1, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			var param CustomRecognitionParam
			input := fmt.Sprintf(`{"custom_recognition":"CustomIndex","index":%d}`, index)
			require.NoError(t, json.Unmarshal([]byte(input), &param))
			encoded, err := json.Marshal(RecCustom(param))
			require.NoError(t, err)
			want := `{"type":"Custom","param":{"custom_recognition":"CustomIndex"}}`
			if index != 0 {
				want = fmt.Sprintf(`{"type":"Custom","param":%s}`, input)
			}
			require.JSONEq(t, want, string(encoded))
		})
	}
}

func TestCustomRecognitionParam_IndexRejectsNonInteger(t *testing.T) {
	for _, value := range []string{`1.5`, `"1"`} {
		t.Run(value, func(t *testing.T) {
			recognition := *RecCustom(CustomRecognitionParam{CustomRecognition: "Keep"})
			before := recognition
			input := fmt.Sprintf(`{"type":"Custom","param":{"custom_recognition":"Replace","index":%s}}`, value)
			require.Error(t, json.Unmarshal([]byte(input), &recognition))
			require.Equal(t, before, recognition)
		})
	}
}

func TestResource_CustomRecognitionIndexRoundTrip(t *testing.T) {
	resource := createResource(t)
	defer destroyEventually(t, resource)
	for _, index := range []int{0, 2, -1, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			input := fmt.Sprintf(`{"custom_recognition":"CustomIndex","index":%d}`, index)
			var param CustomRecognitionParam
			require.NoError(t, json.Unmarshal([]byte(input), &param))
			node := NewNode(fmt.Sprintf("CustomIndex%d", index)).SetRecognition(RecCustom(param))
			require.NoError(t, resource.OverridePipeline(map[string]*Node{node.Name: node}))
			got, err := resource.GetNode(node.Name)
			require.NoError(t, err)
			gotParam, ok := got.Recognition.Param.(*CustomRecognitionParam)
			require.True(t, ok)
			require.Equal(t, index, gotParam.Index)
		})
	}
}

func TestResource_CustomRecognitionZeroIndexInherits(t *testing.T) {
	resource := createResource(t)
	defer destroyEventually(t, resource)
	require.NoError(t, resource.OverridePipeline(`{"CustomIndex":{"recognition":{"type":"Custom","param":{"custom_recognition":"CustomIndex","index":1}}}}`))
	node := NewNode("CustomIndex").SetRecognition(RecCustom(CustomRecognitionParam{CustomRecognition: "CustomIndex"}))
	require.NoError(t, resource.OverridePipeline(map[string]*Node{node.Name: node}))
	got, err := resource.GetNode(node.Name)
	require.NoError(t, err)
	require.Equal(t, 1, got.Recognition.Param.(*CustomRecognitionParam).Index)

	require.NoError(t, resource.OverridePipeline(`{"CustomIndex":{"recognition":{"type":"Custom","param":{"index":0}}}}`))
	got, err = resource.GetNode(node.Name)
	require.NoError(t, err)
	require.Zero(t, got.Recognition.Param.(*CustomRecognitionParam).Index)
}

func TestResource_CustomRecognitionIndexRejectsOutsideInt32(t *testing.T) {
	for _, index := range []string{"2147483648", "-2147483649"} {
		t.Run(index, func(t *testing.T) {
			var param CustomRecognitionParam
			input := fmt.Sprintf(`{"custom_recognition":"CustomIndex","index":%s}`, index)
			err := json.Unmarshal([]byte(input), &param)
			if strconv.IntSize == 32 {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			resource := createResource(t)
			defer destroyEventually(t, resource)
			node := NewNode("CustomIndex").SetRecognition(RecCustom(param))
			require.Error(t, resource.OverridePipeline(map[string]*Node{node.Name: node}))
		})
	}
}

type customIndexRecognitionResult struct {
	detail *RecognitionDetail
	err    error
}

func TestCustomRecognition_IndexSelectsFilteredResult(t *testing.T) {
	cases := []struct {
		index int
		box   Rect
		value string
		hit   bool
	}{
		{index: 0, box: Rect{30, 30, 10, 10}, value: "last", hit: true},
		{index: 1, box: Rect{10, 10, 10, 10}, value: "first", hit: true},
		{index: -1, box: Rect{10, 10, 10, 10}, value: "first", hit: true},
		{index: -2, box: Rect{30, 30, 10, 10}, value: "last", hit: true},
		{index: 2},
		{index: -3},
		{index: math.MinInt32},
		{index: math.MaxInt32},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.index), func(t *testing.T) {
			var param CustomRecognitionParam
			input := fmt.Sprintf(`{"custom_recognition":"CustomIndex","index":%d}`, tc.index)
			require.NoError(t, json.Unmarshal([]byte(input), &param))
			detail := runCustomIndexRecognition(t, param)
			require.Equal(t, tc.hit, detail.Hit)
			require.NotNil(t, detail.Results)
			require.Len(t, detail.Results.All, 3)
			require.Len(t, detail.Results.Filtered, 2)
			if !tc.hit {
				require.Nil(t, detail.Results.Best)
				return
			}
			require.NotNil(t, detail.Results.Best)
			best, ok := detail.Results.Best.AsCustom()
			require.True(t, ok)
			require.Equal(t, tc.box, best.Box)
			require.Equal(t, tc.value, best.Detail)
		})
	}
}

func runCustomIndexRecognition(t *testing.T, param CustomRecognitionParam) *RecognitionDetail {
	t.Helper()
	const results = `{"$all":[{"box":[10,10,10,10],"detail":"first"},{"box":[20,20,10,10],"detail":"middle"},{"box":[30,30,10,10],"detail":"last"}],"$filtered":[{"box":[30,30,10,10],"detail":"last"},{"box":[10,10,10,10],"detail":"first"}]}`
	controller := createBlankController(t)
	defer destroyEventually(t, controller)
	connect, err := controller.PostConnect()
	require.NoError(t, err)
	require.True(t, connect.Wait().Success())
	resource := createResource(t)
	defer destroyEventually(t, resource)
	require.NoError(t, resource.RegisterCustomRecognition("CustomIndex", CustomRecognitionFunc(
		func(_ *Context, _ *CustomRecognitionArg) (*CustomRecognitionResult, bool) {
			return &CustomRecognitionResult{Box: Rect{1, 2, 3, 4}, Detail: results}, true
		},
	)))
	got := make(chan customIndexRecognitionResult, 1)
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	require.NoError(t, resource.RegisterCustomAction("RunCustomIndex", CustomActionFunc(
		func(ctx *Context, _ *CustomActionArg) bool {
			detail, err := ctx.RunRecognitionDirect(RecognitionTypeCustom, &param, img)
			got <- customIndexRecognitionResult{detail: detail, err: err}
			return err == nil
		},
	)))
	tasker := createTasker(t)
	defer destroyEventually(t, tasker)
	taskerBind(t, tasker, controller, resource)
	node := NewNode("RunCustomIndex").SetAction(ActCustom(CustomActionParam{CustomAction: "RunCustomIndex"}))
	job, err := tasker.PostTask(node.Name, map[string]*Node{node.Name: node})
	require.NoError(t, err)
	require.True(t, job.Wait().Success())
	select {
	case result := <-got:
		require.NoError(t, result.err)
		require.NotNil(t, result.detail)
		return result.detail
	default:
		t.Fatal("custom action callback was not called")
		return nil
	}
}
