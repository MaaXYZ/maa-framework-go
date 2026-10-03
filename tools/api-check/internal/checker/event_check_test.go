package checker

import (
	"path/filepath"
	"strings"
	"testing"
)

const eventFixtureGo = `package maa
type Event string
const EventNodeAction = Event("Node.Action")
func (c *callback) handleRaw(handle uintptr, msg string, detail []byte) {
 name, status := parseEvent(msg)
 switch Event(name) { case EventNodeAction: handleNodeAction(c.sink, handle, status, detail) }
}
`

const eventFixtureHeader = `#define MaaMsg_Node_Action_Starting ("Node.Action.Starting")
#define MaaMsg_Node_Action_Succeeded ("Node.Action.Succeeded")
#define MaaMsg_Node_Action_Failed ("Node.Action.Failed")
`

func TestEventCoverage(t *testing.T) {
	for _, tt := range []struct {
		name, source, header, want, wantErr string
	}{
		{"matching", eventFixtureGo, eventFixtureHeader, "", ""},
		{"new native event", eventFixtureGo, eventFixtureHeader + "#define MaaMsg_New (\"Node.New.Starting\")\n", "C event message missing in Go: Node.New.Starting", ""},
		{"removed native message", eventFixtureGo, strings.Replace(eventFixtureHeader, "#define MaaMsg_Node_Action_Failed (\"Node.Action.Failed\")", "", 1), "Go event message absent from C: Node.Action.Failed", ""},
		{"empty dispatch", strings.Replace(eventFixtureGo, "handleNodeAction(c.sink, handle, status, detail)", "", 1), eventFixtureHeader, "Go event missing dispatch: EventNodeAction", ""},
		{"wrong switch", strings.Replace(eventFixtureGo, "Event(name)", "Event(\"Node.Action\")", 1), eventFixtureHeader, "Go event missing dispatch: EventNodeAction", ""},
		{"wrong parsed input", strings.Replace(eventFixtureGo, "parseEvent(msg)", "parseEvent(\"Node.Action.Starting\")", 1), eventFixtureHeader, "Go event missing dispatch: EventNodeAction", ""},
		{"unused closure", strings.Replace(eventFixtureGo, "switch Event(name)", "_ = func() { switch Event(name)", 1) + "}\n", eventFixtureHeader, "Go event missing dispatch: EventNodeAction", ""},
		{"unsupported C expression", eventFixtureGo, "#define MaaMsg_New helper()\n", "", "unsupported event message"},
		{"duplicate C message", eventFixtureGo, eventFixtureHeader + "#define MaaMsg_Duplicate (\"Node.Action.Starting\")\n", "", "duplicate event message"},
		{"unsupported Go expression", strings.Replace(eventFixtureGo, "Event(\"Node.Action\")", "makeEvent()", 1), eventFixtureHeader, "", "unsupported event constant"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			headerDir := filepath.Join(root, "headers")
			writeFixtureFile(t, filepath.Join(root, "event.go"), tt.source)
			writeFixtureFile(t, filepath.Join(headerDir, "MaaFramework", "MaaMsg.h"), tt.header)
			issues, err := checkEventCoverage(root, headerDir)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.want == "" && len(issues) != 0 || tt.want != "" && !strings.Contains(issueText(issues), tt.want) {
				t.Fatalf("issues = %v, want %q", issues, tt.want)
			}
		})
	}
}
