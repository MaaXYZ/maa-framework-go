package maa

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fixedShellOutputController struct {
	BlankController
	output string
}

func (c *fixedShellOutputController) Shell(string, int64) (string, bool) {
	return c.output, true
}

func shellOutputTestCases() []struct {
	name   string
	output string
} {
	return []struct {
		name   string
		output string
	}{
		{name: "empty", output: ""},
		{name: "plain", output: "shell output"},
		{name: "embedded NUL", output: "before\x00after"},
		{name: "leading NUL", output: "\x00after"},
		{name: "trailing NUL", output: "before\x00"},
		{name: "multiple NUL", output: "before\x00\x00middle\x00after"},
		{name: "UTF-8 with NUL", output: "中文\x00输出"},
	}
}

func newShellOutputTestController(t *testing.T, output string) *Controller {
	t.Helper()
	ctrl, err := NewCustomController(&fixedShellOutputController{output: output})
	require.NoError(t, err)
	t.Cleanup(func() { destroyEventually(t, ctrl) })

	job, err := ctrl.PostConnect()
	require.NoError(t, err)
	require.True(t, job.Wait().Success())
	return ctrl
}

func TestController_GetShellOutput_PreservesBytes(t *testing.T) {
	for _, tc := range shellOutputTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := newShellOutputTestController(t, tc.output)
			job, err := ctrl.PostShell("test command", time.Second)
			require.NoError(t, err)
			require.True(t, job.Wait().Success())

			output, err := ctrl.GetShellOutput()
			require.NoError(t, err)
			require.Equal(t, tc.output, output)
		})
	}
}

func TestTasker_ShellActionOutput_PreservesBytes(t *testing.T) {
	for _, tc := range shellOutputTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := newShellOutputTestController(t, tc.output)
			res := createResource(t)
			t.Cleanup(func() { destroyEventually(t, res) })
			require.NoError(t, res.Clear())

			tasker := createTasker(t)
			t.Cleanup(func() { destroyEventually(t, tasker) })
			taskerBind(t, tasker, ctrl, res)

			job, err := tasker.PostAction(ActionTypeShell, ShellParam{Cmd: "test command"}, Rect{}, nil)
			require.NoError(t, err)
			require.True(t, job.Wait().Success())
			detail, err := job.GetDetail()
			require.NoError(t, err)
			require.NotNil(t, detail)
			require.Len(t, detail.Nodes, 1)

			node, err := detail.Nodes[0].GetDetail()
			require.NoError(t, err)
			require.NotNil(t, node)
			require.NotNil(t, node.Action)
			require.True(t, node.Action.Success)

			var raw struct {
				Output string `json:"output"`
			}
			require.NoError(t, json.Unmarshal([]byte(node.Action.DetailJson), &raw))
			require.Equal(t, tc.output, raw.Output)

			require.NotNil(t, node.Action.Result)
			result, ok := node.Action.Result.AsShell()
			require.True(t, ok)
			require.NotNil(t, result)
			require.True(t, result.Success)
			require.Equal(t, tc.output, result.Output)

			output, err := ctrl.GetShellOutput()
			require.NoError(t, err)
			require.Equal(t, tc.output, output)
		})
	}
}
