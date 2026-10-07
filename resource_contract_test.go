package maa

import (
	"image"
	"image/color"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

// stubCustomRecognition is a recognition runner that counts invocations.
type stubCustomRecognition struct {
	calls int32
}

func (s *stubCustomRecognition) Run(_ *Context, _ *CustomRecognitionArg) (*CustomRecognitionResult, bool) {
	atomic.AddInt32(&s.calls, 1)
	return &CustomRecognitionResult{
		Box:    Rect{1, 2, 3, 4},
		Detail: "stubCustomRecognition",
	}, true
}

// stubCustomAction is an action runner that counts invocations.
type stubCustomAction struct {
	calls int32
}

func (s *stubCustomAction) Run(_ *Context, _ *CustomActionArg) bool {
	atomic.AddInt32(&s.calls, 1)
	return true
}

// bindTaskerWithBlankControllerForTest builds a blank controller, connects it,
// and binds a fresh resource and tasker to it, mirroring the end-to-end
// registration tests in resource_test.go.
func bindTaskerWithBlankControllerForTest(t *testing.T) (*Resource, *Tasker) {
	t.Helper()
	ctrl := createBlankController(t)
	t.Cleanup(func() { require.NoError(t, ctrl.Destroy()) })
	connectJob, err := ctrl.PostConnect()
	require.NoError(t, err)
	isConnected := connectJob.Wait().Success()
	require.True(t, isConnected)

	res := createResource(t)
	t.Cleanup(func() { require.NoError(t, res.Destroy()) })

	tasker := createTasker(t)
	t.Cleanup(func() { require.NoError(t, tasker.Destroy()) })
	taskerBind(t, tasker, ctrl, res)
	return res, tasker
}

func TestResource_DuplicateCustomRegistrationRejected(t *testing.T) {
	t.Run("SameNameRecognition", func(t *testing.T) {
		res, tasker := bindTaskerWithBlankControllerForTest(t)
		name := "DupRec"

		first := &stubCustomRecognition{}
		require.NoError(t, res.RegisterCustomRecognition(name, first))
		second := &stubCustomRecognition{}
		require.Error(t, res.RegisterCustomRecognition(name, second))

		pipeline := NewPipeline()
		node := NewNode(t.Name()).
			SetRecognition(RecCustom(CustomRecognitionParam{CustomRecognition: name}))
		pipeline.AddNode(node)

		taskJob, err := tasker.PostTask(node.Name, pipeline)
		require.NoError(t, err)
		require.True(t, taskJob.Wait().Success())
		require.NotZero(t, atomic.LoadInt32(&first.calls))
		require.Zero(t, atomic.LoadInt32(&second.calls))
	})

	t.Run("SameNameAction", func(t *testing.T) {
		res, tasker := bindTaskerWithBlankControllerForTest(t)
		name := "DupAct"

		first := &stubCustomAction{}
		require.NoError(t, res.RegisterCustomAction(name, first))
		second := &stubCustomAction{}
		require.Error(t, res.RegisterCustomAction(name, second))

		pipeline := NewPipeline()
		node := NewNode(t.Name()).
			SetAction(ActCustom(CustomActionParam{CustomAction: name}))
		pipeline.AddNode(node)

		taskJob, err := tasker.PostTask(node.Name, pipeline)
		require.NoError(t, err)
		require.True(t, taskJob.Wait().Success())
		require.NotZero(t, atomic.LoadInt32(&first.calls))
		require.Zero(t, atomic.LoadInt32(&second.calls))
	})

	t.Run("CrossCategory", func(t *testing.T) {
		res, tasker := bindTaskerWithBlankControllerForTest(t)
		name := "CrossKind"

		first := &stubCustomRecognition{}
		require.NoError(t, res.RegisterCustomRecognition(name, first))
		second := &stubCustomAction{}
		require.Error(t, res.RegisterCustomAction(name, second))

		pipeline := NewPipeline()
		node := NewNode(t.Name()).
			SetRecognition(RecCustom(CustomRecognitionParam{CustomRecognition: name}))
		pipeline.AddNode(node)

		taskJob, err := tasker.PostTask(node.Name, pipeline)
		require.NoError(t, err)
		require.True(t, taskJob.Wait().Success())
		require.NotZero(t, atomic.LoadInt32(&first.calls))
		require.Zero(t, atomic.LoadInt32(&second.calls))
	})
}

func TestResource_PostOcrModel(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	job, err := res.PostOcrModel("./test/data_set/PipelineSmoking/resource/model/ocr")
	require.NoError(t, err)
	require.True(t, job.Wait().Success())
}

func TestResource_PostImage(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	job, err := res.PostImage("./test/data_set/PipelineSmoking/resource/image/Awards/DailyBadge.png")
	require.NoError(t, err)
	require.True(t, job.Wait().Success())
}

func TestResource_PostOcrModelMissingPathFails(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	job, err := res.PostOcrModel("./test/data_set/definitely-missing-ocr-model")
	require.NoError(t, err)
	require.NotNil(t, job)
	require.True(t, job.Wait().Failure())
}

func TestResource_PostImageMissingPathFails(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	job, err := res.PostImage("./test/data_set/definitely-missing-image.png")
	require.NoError(t, err)
	require.NotNil(t, job)
	require.True(t, job.Wait().Failure())
}

func TestResource_PostOcrModelNativeInvalidID(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()

	oldPost := native.MaaResourcePostOcrModel
	native.MaaResourcePostOcrModel = func(uintptr, string) int64 { return 0 }
	defer func() { native.MaaResourcePostOcrModel = oldPost }()

	job, err := res.PostOcrModel("whatever")
	require.Error(t, err)
	require.NotNil(t, job)
	require.True(t, job.Failure())
	require.True(t, job.Done())
	require.ErrorIs(t, job.Error(), err)
}

func TestResource_PostImageNativeInvalidID(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()

	oldPost := native.MaaResourcePostImage
	native.MaaResourcePostImage = func(uintptr, string) int64 { return 0 }
	defer func() { native.MaaResourcePostImage = oldPost }()

	job, err := res.PostImage("whatever")
	require.Error(t, err)
	require.NotNil(t, job)
	require.True(t, job.Failure())
	require.True(t, job.Done())
	require.ErrorIs(t, job.Error(), err)
}

func TestResource_OverrideImage(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()

	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	img.Set(0, 0, color.RGBA{R: 0xFF, A: 0xFF})
	img.Set(1, 1, color.RGBA{G: 0xFF, A: 0xFF})
	img.Set(2, 1, color.RGBA{B: 0xFF, A: 0xFF})

	require.NoError(t, res.OverrideImage("OverrideImageSuccess", img))

	// Inputs the image buffer refuses to encode: a nil image (including a
	// typed-nil pointer) and an image with non-positive bounds.
	require.Error(t, res.OverrideImage("OverrideImageNil", nil))
	require.Error(t, res.OverrideImage("OverrideImageEmpty", image.NewRGBA(image.Rect(0, 0, 0, 0))))
}

func TestResource_OverridePipelineInputForms(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()

	require.NoError(t, res.OverridePipeline(`{"OverridePipelineStringForm": {"timeout": 20000}}`))
	require.NoError(t, res.OverridePipeline([]byte(`{"OverridePipelineBytesForm": {"timeout": 20000}}`)))
	require.NoError(t, res.OverridePipeline("{}"))

	// Only a JSON object keyed by node name is accepted; arrays, scalars,
	// null, and an untyped nil must be rejected by the local pre-validation.
	for _, invalid := range []any{"[]", "null", "42", nil} {
		require.ErrorContains(t, res.OverridePipeline(invalid), "JSON object")
	}
}

type resourceLoadingEvent struct {
	status EventStatus
	detail ResourceLoadingDetail
}

type resourceLoadingRecorder struct {
	mu     sync.Mutex
	events []resourceLoadingEvent
}

func (r *resourceLoadingRecorder) onLoading(status EventStatus, detail ResourceLoadingDetail) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, resourceLoadingEvent{status: status, detail: detail})
}

func (r *resourceLoadingRecorder) snapshot() []resourceLoadingEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]resourceLoadingEvent(nil), r.events...)
}

func TestResource_ResourceLoadingSinkEvents(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()

	recorder := &resourceLoadingRecorder{}
	sinkId := res.OnResourceLoading(recorder.onLoading)
	require.NotZero(t, sinkId)

	resDir := "./test/data_set/PipelineSmoking/resource"
	bundleJob, err := res.PostBundle(resDir)
	require.NoError(t, err)
	require.True(t, bundleJob.Wait().Success())

	// The loader thread may deliver the terminal event just after the job
	// status turns terminal; await the events instead of assuming ordering.
	var startingDetail, succeededDetail ResourceLoadingDetail
	require.Eventually(t, func() bool {
		starting, succeeded := false, false
		for _, ev := range recorder.snapshot() {
			switch ev.status {
			case EventStatusStarting:
				starting, startingDetail = true, ev.detail
			case EventStatusSucceeded:
				succeeded, succeededDetail = true, ev.detail
			}
		}
		return starting && succeeded
	}, jobConcurrencyTimeout, 5*time.Millisecond)

	events := recorder.snapshot()
	require.Len(t, events, 2)
	for _, detail := range []ResourceLoadingDetail{startingDetail, succeededDetail} {
		require.Greater(t, detail.ResID, uint64(0))
		require.Equal(t, "Bundle", detail.Type)
		require.True(t, strings.HasSuffix(detail.Path, "test/data_set/PipelineSmoking/resource"),
			"path %q should end with the posted bundle directory", detail.Path)
	}

	// A sink removed before any load never sees later events; a fresh
	// resource avoids the loader's post-completion settle window.
	removed := &resourceLoadingRecorder{}
	res2 := createResource(t)
	defer res2.Destroy()
	removedSink := res2.OnResourceLoading(removed.onLoading)
	require.NotZero(t, removedSink)
	res2.RemoveSink(removedSink)

	removedJob, err := res2.PostBundle(resDir)
	require.NoError(t, err)
	require.True(t, removedJob.Wait().Success())
	require.Empty(t, removed.snapshot())
}

func TestResource_GetCustomRegistrationLists(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()

	recs, err := res.GetCustomRecognitionList()
	require.NoError(t, err)
	require.Empty(t, recs)
	acts, err := res.GetCustomActionList()
	require.NoError(t, err)
	require.Empty(t, acts)

	require.NoError(t, res.RegisterCustomRecognition("ListRec1", &stubCustomRecognition{}))
	require.NoError(t, res.RegisterCustomRecognition("ListRec2", &stubCustomRecognition{}))
	require.NoError(t, res.RegisterCustomAction("ListAct1", &stubCustomAction{}))
	require.NoError(t, res.RegisterCustomAction("ListAct2", &stubCustomAction{}))

	recs, err = res.GetCustomRecognitionList()
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"ListRec1", "ListRec2"}, recs)
	acts, err = res.GetCustomActionList()
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"ListAct1", "ListAct2"}, acts)

	require.NoError(t, res.UnregisterCustomRecognition("ListRec1"))
	recs, err = res.GetCustomRecognitionList()
	require.NoError(t, err)
	require.Equal(t, []string{"ListRec2"}, recs)
	require.NoError(t, res.ClearCustomRecognition())
	recs, err = res.GetCustomRecognitionList()
	require.NoError(t, err)
	require.Empty(t, recs)

	require.NoError(t, res.UnregisterCustomAction("ListAct1"))
	acts, err = res.GetCustomActionList()
	require.NoError(t, err)
	require.Equal(t, []string{"ListAct2"}, acts)
	require.NoError(t, res.ClearCustomAction())
	acts, err = res.GetCustomActionList()
	require.NoError(t, err)
	require.Empty(t, acts)
}

func TestResource_GetNodeJSONUnknownName(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	resDir := "./test/data_set/PipelineSmoking/resource"
	bundleJob, err := res.PostBundle(resDir)
	require.NoError(t, err)
	require.True(t, bundleJob.Wait().Success())

	_, err = res.GetNodeJSON("NoSuchNodeAnywhere")
	require.Error(t, err)
}

func TestResource_GetHashWithoutLoadedContent(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	_, err := res.GetHash()
	require.Error(t, err)
}

func TestResource_OverrideNextUnknownNode(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()

	name := "UnknownNodeCreatedByOverride"
	require.NoError(t, res.OverrideNext(name, []NextItem{{Name: "a"}}))

	nodes, err := res.GetNodeList()
	require.NoError(t, err)
	require.Contains(t, nodes, name)

	node, err := res.GetNode(name)
	require.NoError(t, err)
	require.Len(t, node.Next, 1)
	require.Equal(t, "a", node.Next[0].Name)

	// An empty next list clears the node's next list.
	require.NoError(t, res.OverrideNext(name, nil))
	node, err = res.GetNode(name)
	require.NoError(t, err)
	require.Empty(t, node.Next)
}

func TestResource_UnregisterUnknownCustomNameNoOp(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()
	require.NoError(t, res.UnregisterCustomAction("NeverRegisteredAction"))
	require.NoError(t, res.UnregisterCustomRecognition("NeverRegisteredRecognition"))
}

func TestResource_RegisterNilCustomRunnerRejected(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()

	require.ErrorContains(t, res.RegisterCustomAction("NilAction", nil), "nil")
	require.ErrorContains(t, res.RegisterCustomAction("NilAction", CustomActionFunc(nil)), "nil")
	require.ErrorContains(t, res.RegisterCustomRecognition("NilRecognition", nil), "nil")
	require.ErrorContains(t, res.RegisterCustomRecognition("NilRecognition", CustomRecognitionFunc(nil)), "nil")

	acts, err := res.GetCustomActionList()
	require.NoError(t, err)
	require.Empty(t, acts)
	recs, err := res.GetCustomRecognitionList()
	require.NoError(t, err)
	require.Empty(t, recs)
}
