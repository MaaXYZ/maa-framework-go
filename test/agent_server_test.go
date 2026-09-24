package test

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/buffer"
	"github.com/stretchr/testify/require"
)

const (
	agentTestRecognition = "AgentRoundTripRecognition"
	agentTestAction      = "AgentRoundTripAction"
	agentTestNode        = "AgentRoundTrip"
)

var agentTestBox = maa.Rect{10, 20, 30, 40}

type agentCallbackReport struct {
	RecognitionCalls int      `json:"recognition_calls"`
	ActionCalls      int      `json:"action_calls"`
	RecognitionName  string   `json:"recognition_name"`
	ActionName       string   `json:"action_name"`
	RecognitionParam string   `json:"recognition_param"`
	ActionParam      string   `json:"action_param"`
	RecognitionROI   maa.Rect `json:"recognition_roi"`
	ActionBox        maa.Rect `json:"action_box"`
	ImageWidth       int      `json:"image_width"`
	ImageHeight      int      `json:"image_height"`
	ImagePixel       [4]uint8 `json:"image_pixel"`
	CachedImage      bool     `json:"cached_image"`
	BufferRoundTrip  bool     `json:"buffer_round_trip"`
	NodeRead         bool     `json:"node_read"`
	TaskRead         bool     `json:"task_read"`
	RecognitionRead  bool     `json:"recognition_read"`
	Error            string   `json:"error"`
}

func TestAgentServer_CallbackRoundTrip(t *testing.T) {
	if os.Getenv("MAA_AGENT_TEST_SERVER") == "1" {
		runAgentServerHelper(t)
		return
	}

	dir := t.TempDir()
	identifierBytes := make([]byte, 8)
	_, err := rand.Read(identifierBytes)
	require.NoError(t, err)
	identifier := "go-test-" + hex.EncodeToString(identifierBytes)
	readyPath := filepath.Join(dir, "ready")
	reportPath := filepath.Join(dir, "report.json")
	logPath := filepath.Join(dir, "server.log")
	logFile, err := os.Create(logPath)
	require.NoError(t, err)
	defer logFile.Close()

	cmd := exec.Command(os.Args[0], "-test.run=^TestAgentServer_CallbackRoundTrip$")
	cmd.Env = append(os.Environ(),
		"MAA_AGENT_TEST_SERVER=1",
		"MAA_AGENT_TEST_ID="+identifier,
		"MAA_AGENT_TEST_READY="+readyPath,
		"MAA_AGENT_TEST_REPORT="+reportPath,
	)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	require.NoError(t, cmd.Start())
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	serverExited := false
	t.Cleanup(func() {
		if serverExited {
			return
		}
		_ = cmd.Process.Kill()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("agent server process did not exit after kill")
		}
	})

	readyDeadline := time.NewTimer(10 * time.Second)
	defer readyDeadline.Stop()
	readyTick := time.NewTicker(20 * time.Millisecond)
	defer readyTick.Stop()
	for {
		if _, err := os.Stat(readyPath); err == nil {
			break
		}
		select {
		case err := <-done:
			serverExited = true
			t.Fatalf("agent server exited before ready: %v\n%s", err, readAgentLog(logPath))
		case <-readyDeadline.C:
			t.Fatalf("agent server did not become ready:\n%s", readAgentLog(logPath))
		case <-readyTick.C:
		}
	}

	client, err := maa.NewAgentClient(maa.WithIdentifier(identifier))
	require.NoError(t, err)
	connected := false
	var res *maa.Resource
	var ctrl *maa.Controller
	var tasker *maa.Tasker
	t.Cleanup(func() {
		if connected {
			_ = client.Disconnect()
		}
		if tasker != nil {
			tasker.Destroy()
		}
		if ctrl != nil {
			ctrl.Destroy()
		}
		if res != nil {
			res.Destroy()
		}
		client.Destroy()
	})
	require.NoError(t, client.SetTimeout(5*time.Second))
	res, err = maa.NewResource()
	require.NoError(t, err)
	require.NoError(t, client.BindResource(res))
	require.NoError(t, client.Connect(), readAgentLog(logPath))
	connected = true
	require.True(t, client.Connected())
	recNames, err := client.GetCustomRecognitionList()
	require.NoError(t, err)
	require.Contains(t, recNames, agentTestRecognition)
	actionNames, err := client.GetCustomActionList()
	require.NoError(t, err)
	require.Contains(t, actionNames, agentTestAction)

	ctrl, err = maa.NewBlankController()
	require.NoError(t, err)
	require.True(t, ctrl.PostConnect().Wait().Success())
	tasker, err = maa.NewTasker()
	require.NoError(t, err)
	require.NoError(t, tasker.BindResource(res))
	require.NoError(t, tasker.BindController(ctrl))
	require.True(t, tasker.Initialized())

	node := maa.NewNode(agentTestNode).
		SetRecognition(maa.RecCustom(maa.CustomRecognitionParam{
			ROI:                    maa.NewTargetRect(maa.Rect{2, 3, 80, 90}),
			CustomRecognition:      agentTestRecognition,
			CustomRecognitionParam: map[string]any{"token": "recognition"},
		})).
		SetAction(maa.ActCustom(maa.CustomActionParam{
			CustomAction:      agentTestAction,
			CustomActionParam: map[string]any{"token": "action"},
		})).
		SetTimeout(6 * time.Second).
		SetPreDelay(0).
		SetPostDelay(0)
	pipeline := maa.NewPipeline().AddNode(node)
	job := tasker.PostTask(agentTestNode, pipeline).Wait()
	require.NoError(t, job.Error())
	require.True(t, job.Success(), "task failed; server log:\n%s", readAgentLog(logPath))
	detail, err := job.GetDetail()
	require.NoError(t, err)
	require.Len(t, detail.Nodes, 1)
	nodeDetail, err := detail.Nodes[0].GetDetail()
	require.NoError(t, err)
	require.NotNil(t, nodeDetail.Recognition)
	require.True(t, nodeDetail.Recognition.Hit)
	require.Equal(t, agentTestBox, nodeDetail.Recognition.Box)
	require.NotNil(t, nodeDetail.Recognition.Results)
	require.NotNil(t, nodeDetail.Recognition.Results.Best)
	result, ok := nodeDetail.Recognition.Results.Best.AsCustom()
	require.True(t, ok)
	require.Equal(t, "recognized by agent", result.Detail)
	require.NotNil(t, nodeDetail.Action)
	require.True(t, nodeDetail.Action.Success)
	require.Equal(t, agentTestBox, nodeDetail.Action.Box)

	require.NoError(t, client.Disconnect())
	connected = false
	select {
	case err := <-done:
		serverExited = true
		require.NoError(t, err, readAgentLog(logPath))
	case <-time.After(5 * time.Second):
		t.Fatalf("agent server did not exit after disconnect:\n%s", readAgentLog(logPath))
	}
	reportBytes, err := os.ReadFile(reportPath)
	require.NoError(t, err, readAgentLog(logPath))
	var report agentCallbackReport
	require.NoError(t, json.Unmarshal(reportBytes, &report))
	require.Empty(t, report.Error, readAgentLog(logPath))
	require.Equal(t, 1, report.RecognitionCalls)
	require.Equal(t, 1, report.ActionCalls)
	require.Equal(t, agentTestRecognition, report.RecognitionName)
	require.Equal(t, agentTestAction, report.ActionName)
	require.JSONEq(t, `{"token":"recognition"}`, report.RecognitionParam)
	require.JSONEq(t, `{"token":"action"}`, report.ActionParam)
	require.Equal(t, maa.Rect{2, 3, 80, 90}, report.RecognitionROI)
	require.Equal(t, agentTestBox, report.ActionBox)
	require.Equal(t, 1280, report.ImageWidth)
	require.Equal(t, 720, report.ImageHeight)
	require.Equal(t, [4]uint8{0, 0, 0, 255}, report.ImagePixel)
	require.True(t, report.CachedImage)
	require.True(t, report.BufferRoundTrip)
	require.True(t, report.NodeRead)
	require.True(t, report.TaskRead)
	require.True(t, report.RecognitionRead)
}

func runAgentServerHelper(t *testing.T) {
	var report agentCallbackReport
	err := maa.AgentServerRegisterCustomRecognition(agentTestRecognition, maa.CustomRecognitionFunc(
		func(_ *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
			report.RecognitionCalls++
			report.RecognitionName = arg.CustomRecognitionName
			report.RecognitionParam = arg.CustomRecognitionParam
			report.RecognitionROI = arg.Roi
			if arg.TaskID == 0 || arg.CurrentTaskName != agentTestNode || arg.Img == nil {
				report.Error = fmt.Sprintf("invalid recognition argument: task=%d name=%q image=%v", arg.TaskID, arg.CurrentTaskName, arg.Img)
				return nil, false
			}
			bounds := arg.Img.Bounds()
			if bounds.Empty() {
				report.Error = "recognition image is empty"
				return nil, false
			}
			report.ImageWidth, report.ImageHeight = bounds.Dx(), bounds.Dy()
			r, g, b, a := arg.Img.At(bounds.Min.X, bounds.Min.Y).RGBA()
			report.ImagePixel = [4]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
			return &maa.CustomRecognitionResult{Box: agentTestBox, Detail: "recognized by agent"}, true
		},
	))
	require.NoError(t, err)
	err = maa.AgentServerRegisterCustomAction(agentTestAction, maa.CustomActionFunc(
		func(ctx *maa.Context, arg *maa.CustomActionArg) bool {
			report.ActionCalls++
			report.ActionName = arg.CustomActionName
			report.ActionParam = arg.CustomActionParam
			report.ActionBox = arg.Box
			if arg.TaskID == 0 || arg.CurrentTaskName != agentTestNode || arg.RecognitionDetail == nil {
				report.Error = fmt.Sprintf("invalid action argument: task=%d name=%q detail=%v", arg.TaskID, arg.CurrentTaskName, arg.RecognitionDetail)
				return false
			}
			if !arg.RecognitionDetail.Hit || arg.RecognitionDetail.Box != agentTestBox {
				report.Error = fmt.Sprintf("wrong recognition detail: %+v", arg.RecognitionDetail)
				return false
			}
			tasker := ctx.GetTasker()
			if tasker == nil {
				report.Error = "context returned nil tasker"
				return false
			}
			img, err := tasker.GetController().CacheImage()
			report.CachedImage = err == nil && img != nil && img.Bounds() == image.Rect(0, 0, 1280, 720)
			if report.CachedImage {
				imgBuffer := buffer.NewImageBuffer()
				if imgBuffer != nil {
					if imgBuffer.Set(img) {
						copy := imgBuffer.Get()
						report.BufferRoundTrip = copy != nil && copy.Bounds() == img.Bounds()
					}
					imgBuffer.Destroy()
				}
			}
			nodeJSON, err := ctx.GetNodeJSON(agentTestNode)
			report.NodeRead = err == nil && strings.Contains(nodeJSON, agentTestAction)
			taskDetail, err := tasker.GetTaskDetail(arg.TaskID)
			report.TaskRead = err == nil && taskDetail != nil && taskDetail.ID == arg.TaskID
			recDetail, err := tasker.GetRecognitionDetail(arg.RecognitionDetail.ID)
			report.RecognitionRead = err == nil && recDetail != nil && recDetail.Box == agentTestBox
			if !report.CachedImage || !report.BufferRoundTrip || !report.NodeRead || !report.TaskRead || !report.RecognitionRead {
				report.Error = fmt.Sprintf("remote context calls failed: cached=%t buffer=%t node=%t task=%t recognition=%t", report.CachedImage, report.BufferRoundTrip, report.NodeRead, report.TaskRead, report.RecognitionRead)
				return false
			}
			return true
		},
	))
	require.NoError(t, err)
	require.NoError(t, maa.AgentServerStartUp(os.Getenv("MAA_AGENT_TEST_ID")))
	require.NoError(t, os.WriteFile(os.Getenv("MAA_AGENT_TEST_READY"), []byte("ready"), 0600))
	maa.AgentServerJoin()
	maa.AgentServerShutDown()
	data, err := json.Marshal(report)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(os.Getenv("MAA_AGENT_TEST_REPORT"), data, 0600))
}

func readAgentLog(path string) string {
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err.Error()
	}
	return string(data)
}
