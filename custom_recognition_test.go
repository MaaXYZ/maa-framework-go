package maa

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomRecognition_MissResults(t *testing.T) {
	testCases := []struct {
		name       string
		result     *CustomRecognitionResult
		matched    bool
		wantBox    Rect
		wantDetail string
	}{
		{
			name:       "result with false",
			result:     &CustomRecognitionResult{Box: Rect{10, 20, 30, 40}, Detail: "diagnostic detail"},
			matched:    false,
			wantBox:    Rect{10, 20, 30, 40},
			wantDetail: "diagnostic detail",
		},
		{name: "nil with false"},
		{name: "nil with true", matched: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := createBlankController(t)
			defer ctrl.Destroy()
			require.True(t, ctrl.PostConnect().Wait().Success())

			res := createResource(t)
			defer res.Destroy()
			var called atomic.Bool
			require.NoError(t, res.RegisterCustomRecognition("MissWithDetail", CustomRecognitionFunc(
				func(_ *Context, _ *CustomRecognitionArg) (*CustomRecognitionResult, bool) {
					called.Store(true)
					return tc.result, tc.matched
				},
			)))

			tasker := createTasker(t)
			defer tasker.Destroy()
			taskerBind(t, tasker, ctrl, res)

			events := make(chan uint64, 1)
			sinkID := tasker.OnNodeRecognitionInContext(func(_ *Context, status EventStatus, detail NodeRecognitionDetail) {
				if status == EventStatusFailed {
					events <- detail.RecognitionID
				}
			})
			defer tasker.RemoveContextSink(sinkID)

			pipeline := NewPipeline()
			node := NewNode("MissWithDetail").
				SetRecognition(RecCustom(CustomRecognitionParam{CustomRecognition: "MissWithDetail"})).
				SetTimeout(0)
			pipeline.AddNode(node)
			tasker.PostTask(node.Name, pipeline).Wait()
			require.True(t, called.Load(), "custom recognizer was not called")

			select {
			case id := <-events:
				detail, err := tasker.GetRecognitionDetail(int64(id))
				require.NoError(t, err)
				require.NotNil(t, detail)
				require.False(t, detail.Hit)
				require.NotNil(t, detail.Results)
				require.Nil(t, detail.Results.Best)
				require.Empty(t, detail.Results.Filtered)
				require.Len(t, detail.Results.All, 1)
				result, ok := detail.Results.All[0].AsCustom()
				require.True(t, ok)
				require.Equal(t, tc.wantBox, result.Box)
				require.Equal(t, tc.wantDetail, result.Detail)
			default:
				t.Fatal("recognition failure event was not emitted")
			}
		})
	}
}
