package checker

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Decoded envelope discriminant fixtures. actionWire is a named envelope DTO so
// tests can rewrite the whole envelope or take its address.
const (
	discriminatorWireType    = "type actionWire struct { Type ActionType `json:\"type\"`; Param json.RawMessage `json:\"param\"` }\n"
	discriminatorRawDecl     = "var raw struct { Type ActionType `json:\"type\"`; Param json.RawMessage `json:\"param\"` }"
	discriminatorWireDecl    = "var raw actionWire"
	discriminatorPointerDecl = "var raw = &actionWire{}"
)

const (
	discriminatorActionSig      = "func (a *Action) UnmarshalJSON(data []byte) error {"
	discriminatorRecognitionSig = "func (r *Recognition) UnmarshalJSON(data []byte) error {"
	discriminatorHelperSig      = "func decodeActionParam(t ActionType, data json.RawMessage) (ActionParam, error) {"
)

func withActionWire(source string) string {
	return strings.Replace(source, "type RawActionParam struct", discriminatorWireType+"type RawActionParam struct", 1)
}

func withDoNothing(source string) string {
	return strings.Replace(source, `const ActionTypeClick ActionType = "Click"`, "const ActionTypeDoNothing ActionType = \"DoNothing\"\nconst ActionTypeClick ActionType = \"Click\"", 1)
}

// TestPipelineDiscriminatorRewrites locks in that a decoded Type selector stops
// establishing coverage as soon as the envelope or its Type field can be
// rewritten, aliased, or escaped.
func TestPipelineDiscriminatorRewrites(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string // required extraction error substring
	}{
		{
			name: "field rewrite after decode",
			source: withDoNothing(replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorRawDecl+`
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 raw.Type = ActionTypeDoNothing
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`)),
			want: "rewrite of decoded envelope raw.Type after its JSON decode",
		},
		{
			name: "whole envelope written after decode",
			source: withActionWire(replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorWireDecl+`
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 (raw) = actionWire{Type: ActionTypeClick}
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`)),
			want: "rewrite of decoded envelope raw after its JSON decode",
		},
		{
			name: "pointer field rewrite after decode",
			source: withActionWire(replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorPointerDecl+`
 if err := unmarshalJSON(data, raw); err != nil { return err }
 (*raw).Type = ActionTypeClick
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`)),
			want: "rewrite of decoded envelope raw.Type after its JSON decode",
		},
		{
			name: "pointer whole envelope written after decode",
			source: withActionWire(replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorPointerDecl+`
 if err := unmarshalJSON(data, raw); err != nil { return err }
 *raw = actionWire{Type: ActionTypeClick}
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`)),
			want: "rewrite of decoded envelope raw after its JSON decode",
		},
		{
			name: "field rewrite inside a closure",
			source: replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorRawDecl+`
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 func() { raw.Type = ActionTypeClick }()
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`),
			want: "rewrite of decoded envelope raw.Type inside a closure",
		},
		{
			name: "uninvoked closure rewrite before the decode",
			source: replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorRawDecl+`
 _ = func() { raw.Type = ActionTypeClick }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`),
			want: "rewrite of decoded envelope raw.Type inside a closure",
		},
		{
			name: "whole envelope written inside a closure",
			source: withActionWire(replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorWireDecl+`
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 func() { raw = actionWire{Type: ActionTypeClick} }()
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`)),
			want: "rewrite of decoded envelope raw inside a closure",
		},
		{
			name: "envelope alias through a short assignment",
			source: replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorRawDecl+`
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 p := &raw
 _ = p
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`),
			want: "alias of decoded envelope raw",
		},
		{
			name: "envelope alias through a variable declaration",
			source: replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorRawDecl+`
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 var p = &raw
 _ = p
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`),
			want: "alias of decoded envelope raw",
		},
		{
			name: "envelope alias through a typed variable declaration",
			source: withActionWire(replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorWireDecl+`
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 var p *actionWire = &raw
 _ = p
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`)),
			want: "alias of decoded envelope raw",
		},
		{
			name: "discriminant address alias",
			source: replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorRawDecl+`
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 p := &raw.Type
 _ = p
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`),
			want: "alias of decoded envelope raw.Type",
		},
		{
			name: "envelope escape through a call",
			source: withActionWire(replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorWireDecl+`
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 escapeEnvelope(&raw)
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}
func escapeEnvelope(v *actionWire) {}
`)),
			want: "escape of decoded envelope raw",
		},
		{
			name: "pointer method call on the envelope",
			source: withActionWire(replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorWireDecl+`
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 raw.reset()
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}
func (w *actionWire) reset() {}
`)),
			want: "method call on decoded envelope raw",
		},
		{
			name: "helper reassigns its enum parameter",
			source: replaceFixtureFunc(pipelineFixtureGo, discriminatorHelperSig, `func decodeActionParam(t ActionType, data json.RawMessage) (ActionParam, error) {
 t = ActionTypeClick
 var param ActionParam
 switch t { case ActionTypeClick: param = &ClickParam{} }
 if err := unmarshalJSON(data, param); err != nil { return nil, err }
 return param, nil
}`),
			want: "reassigns its ActionType parameter t",
		},
		{
			name: "helper shadows its enum parameter with var",
			source: replaceFixtureFunc(pipelineFixtureGo, discriminatorHelperSig, `func decodeActionParam(t ActionType, data json.RawMessage) (ActionParam, error) {
 var param ActionParam
 {
  var t ActionType = ActionTypeClick
  switch t { case ActionTypeClick: param = &ClickParam{} }
 }
 if err := unmarshalJSON(data, param); err != nil { return nil, err }
 return param, nil
}`),
			want: "shadows its ActionType parameter t",
		},
		{
			name: "helper shadows its enum parameter with a short declaration",
			source: replaceFixtureFunc(pipelineFixtureGo, discriminatorHelperSig, `func decodeActionParam(t ActionType, data json.RawMessage) (ActionParam, error) {
 { t := ActionTypeClick; _ = t }
 var param ActionParam
 switch t { case ActionTypeClick: param = &ClickParam{} }
 if err := unmarshalJSON(data, param); err != nil { return nil, err }
 return param, nil
}`),
			want: "shadows its ActionType parameter t",
		},
		{
			name: "helper shadows its enum parameter with a range variable",
			source: replaceFixtureFunc(pipelineFixtureGo, discriminatorHelperSig, `func decodeActionParam(t ActionType, data json.RawMessage) (ActionParam, error) {
 for t := range []ActionType{ActionTypeClick} { _ = t }
 var param ActionParam
 switch t { case ActionTypeClick: param = &ClickParam{} }
 if err := unmarshalJSON(data, param); err != nil { return nil, err }
 return param, nil
}`),
			want: "shadows its ActionType parameter t",
		},
		{
			name: "helper writes its enum parameter inside a closure",
			source: replaceFixtureFunc(pipelineFixtureGo, discriminatorHelperSig, `func decodeActionParam(t ActionType, data json.RawMessage) (ActionParam, error) {
 func() { t = ActionTypeClick }()
 var param ActionParam
 switch t { case ActionTypeClick: param = &ClickParam{} }
 if err := unmarshalJSON(data, param); err != nil { return nil, err }
 return param, nil
}`),
			want: "writes its ActionType parameter t inside a closure",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.source == pipelineFixtureGo {
				t.Fatalf("fixture mutation did not apply")
			}
			dir, path := writePipelineFixture(t, pipelineFixtureSchema(), tt.source)
			if _, _, err := checkPipelineCoverage(dir, path, nil); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v; want %q", err, tt.want)
			}
		})
	}
}

func TestPipelineDiscriminatorNestedWritesAndEscapes(t *testing.T) {
	for _, tt := range []struct{ name, body, want string }{
		{"field address LHS", "*(&raw.Type) = ActionTypeClick", "rewrite of decoded envelope raw.Type after its JSON decode"},
		{"tuple second LHS", "_, raw.Type = replacementType()", "rewrite of decoded envelope raw.Type after its JSON decode"},
		{"range field LHS", "for _, raw.Type = range []ActionType{ActionTypeClick} {}", "rewrite of decoded envelope raw.Type after its JSON decode"},
		{"range field address LHS", "for _, *(&raw.Type) = range []ActionType{ActionTypeClick} {}", "rewrite of decoded envelope raw.Type after its JSON decode"},
		{"closure range field LHS", "func() { for _, raw.Type = range []ActionType{ActionTypeClick} {} }()", "rewrite of decoded envelope raw.Type inside a closure"},
		{"range pointer alias", "for _, p := range []*actionWire{&raw} { p.Type = ActionTypeClick }", "escape of decoded envelope raw through a range"},
		{"channel envelope alias", "ch := make(chan *actionWire, 1); ch <- &raw", "escape of decoded envelope raw through a channel"},
		{"channel field alias", "ch := make(chan *ActionType, 1); ch <- &raw.Type", "escape of decoded envelope raw.Type through a channel"},
		{"map key envelope alias", "p := map[*actionWire]int{&raw: 1}; _ = p", "alias of decoded envelope raw"},
		{"map key field alias", "p := map[*ActionType]int{&raw.Type: 1}; _ = p", "alias of decoded envelope raw.Type"},
		{"field pointer method", "raw.Type.reset()", "method call on decoded envelope raw.Type"},
		{"nested composite envelope alias", "p := struct { P *actionWire }{P: &raw}; p.P.Type = ActionTypeClick", "alias of decoded envelope raw"},
		{"nested composite field alias", "p := &struct { P *ActionType }{P: &raw.Type}; *p.P = ActionTypeClick", "alias of decoded envelope raw.Type"},
		{"closure returns envelope pointer", "p := func() *actionWire { return &raw }(); p.Type = ActionTypeClick", "escape of decoded envelope raw through a return"},
		{"closure returns field pointer", "p := func() *ActionType { return &raw.Type }(); *p = ActionTypeClick", "escape of decoded envelope raw.Type through a return"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := withActionWire(replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw actionWire
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 `+tt.body+`
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}
func replacementType() (int, ActionType) { return 0, ActionTypeClick }
func (t *ActionType) reset() { *t = ActionTypeClick }
`))
			dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
			if _, _, err := checkPipelineCoverage(dir, path, nil); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v; want %q", err, tt.want)
			}
		})
	}
}

func TestPipelineHelperDiscriminatorPointers(t *testing.T) {
	for _, tt := range []struct{ name, body, want string }{
		{"address LHS", "*(&t) = ActionTypeClick", "reassigns its ActionType parameter t"},
		{"range address LHS", "for _, *(&t) = range []ActionType{ActionTypeClick} {}", "reassigns its ActionType parameter t"},
		{"address alias", "p := &t; *p = ActionTypeClick", "takes the address of its ActionType parameter t"},
		{"nested address alias", "p := struct { P *ActionType }{P: &t}; *p.P = ActionTypeClick", "takes the address of its ActionType parameter t"},
		{"pointer method", "t.reset()", "calls a method on its ActionType parameter t"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := replaceFixtureFunc(pipelineFixtureGo, discriminatorHelperSig, `func decodeActionParam(t ActionType, data json.RawMessage) (ActionParam, error) {
 `+tt.body+`
 var param ActionParam
 switch t { case ActionTypeClick: param = &ClickParam{} }
 if err := unmarshalJSON(data, param); err != nil { return nil, err }
 return param, nil
}
func (t *ActionType) reset() { *t = ActionTypeClick }
`)
			dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
			if _, _, err := checkPipelineCoverage(dir, path, nil); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v; want %q", err, tt.want)
			}
		})
	}
}

func TestPipelineDiscriminatorCLI(t *testing.T) {
	schema := pipelineFixtureSchema()
	defs := fixtureDefs(schema)
	defs["ActionEnum"].(schemaObject)["enum"] = []any{"Click", "DoNothing"}
	defs["DoNothingV2"] = schemaObject{"properties": schemaObject{
		"type":  schemaObject{"const": "DoNothing"},
		"param": schemaObject{"properties": schemaObject{}},
	}}
	action := fixtureProps(schema, "ActionV2")["action"].(schemaObject)
	action["anyOf"] = append(action["anyOf"].([]any), schemaObject{"$ref": "#/$defs/DoNothingV2"})
	source := strings.Replace(withDoNothing(pipelineFixtureGo), "case ActionTypeClick: param = &ClickParam{}", "case ActionTypeClick: param = &ClickParam{}\n case ActionTypeDoNothing: param = &DoNothingParam{}", 1)
	source += "\ntype DoNothingParam struct{}\n"
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ name, old, new, want string }{
		{"normal discriminant", "", "", "PASS: no inconsistencies found."},
		{"envelope rewrite", "param, err := decodeActionParam(raw.Type, raw.Param)", "raw.Type = ActionTypeDoNothing\n param, err := decodeActionParam(raw.Type, raw.Param)", "rewrite of decoded envelope raw.Type after its JSON decode"},
		{"helper rewrite", "switch t {", "t = ActionTypeDoNothing\n switch t {", "reassigns its ActionType parameter t"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mutated := source
			wantCode := 0
			if tt.old != "" {
				mutated = strings.Replace(source, tt.old, tt.new, 1)
				wantCode = 2
			}
			dir, path := writePipelineFixture(t, schema, mutated)
			writeFixtureFiles(t, dir, repoFixtureFiles())
			config := filepath.Join(dir, "enabled.yaml")
			writeFixtureFile(t, config, "header_dir: deps/include\npipeline_schema: "+filepath.Base(path)+"\n")
			args, err := json.Marshal([]string{"--config", config})
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(binary, "-test.run=^TestPipelineRunHelper$")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "MAA_PIPELINE_RUN_HELPER=1", "MAA_PIPELINE_RUN_ARGS="+string(args))
			output, err := cmd.CombinedOutput()
			code := 0
			if err != nil {
				exit, ok := err.(*exec.ExitError)
				if !ok {
					t.Fatal(err)
				}
				code = exit.ExitCode()
			}
			if code != wantCode || !strings.Contains(string(output), tt.want) {
				t.Fatalf("got exit %d, want %d; output:\n%s", code, wantCode, output)
			}
		})
	}
}

// TestPipelineDiscriminatorTrueCoverage verifies that ordinary initialization
// before the decode and the normal receiver data flow keep establishing
// coverage after the rewrite checks were added.
func TestPipelineDiscriminatorTrueCoverage(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "pre-decode field initialization",
			source: replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorRawDecl+`
 raw.Type = ActionTypeClick
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`),
		},
		{
			name: "pre-decode field address initialization",
			source: replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorRawDecl+`
 *(&raw.Type) = ActionTypeClick
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`),
		},
		{
			name: "pre-decode envelope initialization",
			source: withActionWire(replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw = actionWire{Type: ActionTypeClick}
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`)),
		},
		{
			name: "normal receiver type and param flow",
			source: replaceFixtureFunc(pipelineFixtureGo, discriminatorRecognitionSig, `func (r *Recognition) UnmarshalJSON(data []byte) error {
 var raw struct { Type RecognitionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 r.Type = raw.Type
 switch raw.Type {
 case RecognitionTypeDirectHit: r.Param = &DirectHitParam{}
 }
 return nil
}`),
		},
		{
			name: "nested scalar field reads",
			source: replaceFixtureFunc(pipelineFixtureGo, discriminatorActionSig, `func (a *Action) UnmarshalJSON(data []byte) error {
 `+discriminatorRawDecl+`
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 snapshot := struct { Type ActionType; Param json.RawMessage }{Type: raw.Type, Param: raw.Param}
 _ = snapshot
 _ = func() ActionType { return raw.Type }
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.source == pipelineFixtureGo {
				t.Fatalf("fixture mutation did not apply")
			}
			dir, path := writePipelineFixture(t, pipelineFixtureSchema(), tt.source)
			issues, _, err := checkPipelineCoverage(dir, path, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if text := issueText(issues); text != "" {
				t.Fatalf("unexpected issues: %s", text)
			}
		})
	}
}
