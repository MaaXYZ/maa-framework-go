package checker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const pipelineFixtureGo = `package maa
import "encoding/json"
type ActionType string
const ActionTypeClick ActionType = "Click"
type RecognitionType string
const RecognitionTypeDirectHit RecognitionType = "DirectHit"
type Action struct { Type ActionType ` + "`json:\"type\"`" + `; Param any ` + "`json:\"param\"`" + ` }
func (a *Action) UnmarshalJSON(data []byte) error { _ = actionParam(a.Type); return nil }
func actionParam(t ActionType) any { switch t { case ActionTypeClick: return &ClickParam{}; default: return &RawActionParam{} } }
type RawActionParam struct { Data json.RawMessage ` + "`json:\"-\"`" + ` }
type Recognition struct { Type RecognitionType ` + "`json:\"type\"`" + `; Param any ` + "`json:\"param\"`" + ` }
func (r *Recognition) UnmarshalJSON(data []byte) error { switch r.Type { case RecognitionTypeDirectHit: r.Param = &DirectHitParam{} }; return nil }
type Node struct { Action *Action ` + "`json:\"action\"`" + `; Recognition *Recognition ` + "`json:\"recognition\"`" + ` }
type ClickParam struct { Target string ` + "`json:\"target\"`" + ` }
type DirectHitParam struct{}
type MultiSwipeItem struct { Starting int ` + "`json:\"starting\"`" + ` }
type NextItem struct { Name string ` + "`json:\"name\"`" + ` }
type WaitFreezesParam struct { Time int ` + "`json:\"-\"`" + ` }
func (w WaitFreezesParam) MarshalJSON() ([]byte,error) {
 type NoMethod WaitFreezesParam
 type Unused struct { Ghost int ` + "`json:\"ghost\"`" + ` }
 return json.Marshal(struct { NoMethod; Time int ` + "`json:\"time\"`" + ` }{NoMethod:NoMethod(w),Time:w.Time})
}
func (w *WaitFreezesParam) UnmarshalJSON(data []byte) error {
 type NoMethod WaitFreezesParam
 var raw struct { NoMethod; Time int ` + "`json:\"time\"`" + ` }
 return json.Unmarshal(data, &raw)
}
`

func pipelineFixtureSchema() schemaObject {
	field := func() schemaObject { return schemaObject{"type": "string"} }
	obj := func(props schemaObject) schemaObject { return schemaObject{"type": "object", "properties": props} }
	branch := func(name, param string) schemaObject {
		return schemaObject{"properties": schemaObject{"type": schemaObject{"const": name}, "param": schemaObject{"allOf": []any{schemaObject{"$ref": "#/$defs/" + param}, schemaObject{"$ref": "#/$defs/jsonComments"}}}}}
	}
	wrapper := func(kind, name string) schemaObject {
		return schemaObject{"properties": schemaObject{kind: schemaObject{"anyOf": []any{schemaObject{"$ref": "#/$defs/" + name}}}}}
	}
	return schemaObject{"$defs": schemaObject{
		"Node":       obj(schemaObject{"action": field(), "recognition": field(), "interrupt": schemaObject{"deprecated": true}}),
		"ActionEnum": schemaObject{"enum": []any{"Click"}}, "RecognitionEnum": schemaObject{"enum": []any{"DirectHit"}},
		"ActionV2": wrapper("action", "ClickV2"), "RecognitionV2": wrapper("recognition", "DirectHitV2"),
		"ClickV2": branch("Click", "Click"), "Click": obj(schemaObject{"target": field()}),
		"DirectHitV2": branch("DirectHit", "DirectHit"), "DirectHit": obj(schemaObject{}),
		"SwipeListItem": obj(schemaObject{"starting": field()}), "WaitFreezes": obj(schemaObject{"time": field()}),
		"NodeAttr": obj(schemaObject{"name": field()}), "jsonComments": schemaObject{"patternProperties": schemaObject{}},
	}}
}

func writePipelineFixture(t *testing.T, schema schemaObject, source string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "pipeline.schema.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pipeline.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return dir, path
}

func fixtureDefs(schema schemaObject) schemaObject { return schema["$defs"].(schemaObject) }
func fixtureProps(schema schemaObject, name string) schemaObject {
	return fixtureDefs(schema)[name].(schemaObject)["properties"].(schemaObject)
}
func issueText(issues []issue) string {
	var out []string
	for _, issue := range issues {
		out = append(out, issue.message)
	}
	return strings.Join(out, "\n")
}

func TestPipelineCoverage(t *testing.T) {
	tests := []struct {
		name   string
		change func(schemaObject, string) (schemaObject, string)
		want   []string
	}{
		{name: "complete with codec DTO and unused tagged struct"},
		{name: "missing pressure", change: func(s schemaObject, g string) (schemaObject, string) {
			fixtureProps(s, "Click")["pressure"] = schemaObject{"type": "integer"}
			return s, g
		}, want: []string{"action.Click.param.pressure: missing Go field (ClickParam)"}},
		{name: "new type despite raw fallback", change: func(s schemaObject, g string) (schemaObject, string) {
			defs := fixtureDefs(s)
			defs["ActionEnum"].(schemaObject)["enum"] = []any{"Click", "Future"}
			defs["FutureV2"] = schemaObject{"properties": schemaObject{"type": schemaObject{"const": "Future"}, "param": schemaObject{"properties": schemaObject{}}}}
			wrapper := fixtureProps(s, "ActionV2")["action"].(schemaObject)
			wrapper["anyOf"] = append(wrapper["anyOf"].([]any), schemaObject{"$ref": "#/$defs/FutureV2"})
			return s, g
		}, want: []string{"action.Future.type: missing Go type constant", "action.Future.decoder: missing typed Go decoder case"}},
		{name: "enum without branch", change: func(s schemaObject, g string) (schemaObject, string) {
			fixtureDefs(s)["ActionEnum"].(schemaObject)["enum"] = []any{"Click", "Future"}
			return s, g
		}, want: []string{"action.Future.type", "action.Future.decoder", "action.Future.schema"}},
		{name: "missing decoder", change: func(s schemaObject, g string) (schemaObject, string) {
			return s, strings.Replace(g, "case ActionTypeClick: return &ClickParam{};", "", 1)
		}, want: []string{"action.Click.decoder: missing typed Go decoder case"}},
		{name: "nested missing field", change: func(s schemaObject, g string) (schemaObject, string) {
			fixtureProps(s, "SwipeListItem")["pressure"] = schemaObject{"type": "integer"}
			return s, g
		}, want: []string{"SwipeListItem.pressure: missing Go field (MultiSwipeItem)"}},
		{name: "extra Go field", change: func(s schemaObject, g string) (schemaObject, string) {
			return s, strings.Replace(g, "type ClickParam struct {", "type ClickParam struct { Obsolete int `json:\"obsolete\"`;", 1)
		}, want: []string{"action.Click.param.obsolete: Go field absent from schema (ClickParam,"}},
		{name: "tagged unexported field does not count", change: func(s schemaObject, g string) (schemaObject, string) {
			fixtureProps(s, "Click")["pressure"] = schemaObject{"type": "integer"}
			return s, strings.Replace(g, "type ClickParam struct {", "type ClickParam struct { pressure int `json:\"pressure\"`;", 1)
		}, want: []string{"action.Click.param.pressure: missing Go field (ClickParam)"}},
		{name: "codec DTO missing real field", change: func(s schemaObject, g string) (schemaObject, string) {
			fixtureProps(s, "WaitFreezes")["ghost"] = schemaObject{"type": "integer"}
			return s, g
		}, want: []string{"WaitFreezes.ghost: missing Go field"}},
		{name: "anonymous embedding and type alias", change: func(s schemaObject, g string) (schemaObject, string) {
			return s, strings.Replace(g, "type ClickParam struct { Target string `json:\"target\"` }", "type ClickBase struct { Target string `json:\"target\"` }; type ClickAlias = ClickBase; type ClickParam struct { ClickAlias }", 1)
		}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, g := pipelineFixtureSchema(), pipelineFixtureGo
			if tt.change != nil {
				s, g = tt.change(s, g)
			}
			dir, path := writePipelineFixture(t, s, g)
			issues, excluded, err := checkPipelineCoverage(dir, path, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(excluded) != 0 {
				t.Fatalf("unexpected exclusions: %v", excluded)
			}
			text := issueText(issues)
			if len(issues) != len(tt.want) {
				t.Fatalf("got %d issues; want %d:\n%s", len(issues), len(tt.want), text)
			}
			for _, want := range tt.want {
				if !strings.Contains(text, want) {
					t.Errorf("missing %q in %s", want, text)
				}
			}
		})
	}
}

func TestPipelineAmbiguousGoShapes(t *testing.T) {
	tests := []struct{ name, old, replacement, want string }{
		{"duplicate tags", "type ClickParam struct {", "type ClickParam struct { Other int `json:\"target\"`;", "conflicting JSON field target"},
		{"conflicting promoted fields", "type ClickParam struct { Target string `json:\"target\"` }", "type BaseA struct { Target string `json:\"target\"` }; type BaseB struct { Target string `json:\"target\"` }; type ClickParam struct { BaseA; BaseB }", "conflicting JSON field target"},
		{"promoted field shadow", "type ClickParam struct {", "type ClickBase struct { Target string `json:\"target\"` }; type ClickParam struct { ClickBase;", "conflicting JSON field target"},
		{"unrelated decoder literal", "case ActionTypeClick: return &ClickParam{};", "case ActionTypeClick: log(&ClickParam{}); return nil;", "expected one parameter literal assigned to Param or returned"},
		{"logging literal alongside real decoder", "case ActionTypeClick: return &ClickParam{};", "case ActionTypeClick: log(&RawActionParam{}); return &ClickParam{};", ""},
		{"reassigned wire variable", "return json.Unmarshal(data, &raw)", "raw = something(); return json.Unmarshal(data, &raw)", "reassigned or shadowed wire variable raw"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := strings.Replace(pipelineFixtureGo, tt.old, tt.replacement, 1)
			dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
			issues, _, err := checkPipelineCoverage(dir, path, nil)
			if tt.want == "" {
				if err != nil || len(issues) != 0 {
					t.Fatalf("got %v / %v", issues, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v; want %q", err, tt.want)
			}
		})
	}
}

func TestPipelineSchemaErrors(t *testing.T) {
	tests := []struct {
		name   string
		change func(schemaObject)
		want   string
	}{
		{"missing defs", func(s schemaObject) { delete(s, "$defs") }, "missing object $defs"},
		{"unrecognized schema", func(s schemaObject) { delete(fixtureDefs(s), "ActionV2") }, "ActionV2"},
		{"unknown parameter shape", func(s schemaObject) {
			fixtureDefs(s)["Click"] = schemaObject{"type": "object", "oneOf": []any{schemaObject{"properties": schemaObject{"a": schemaObject{}}}}}
		}, "unsupported oneOf field alternatives"},
		{"ref cycle", func(s schemaObject) {
			fixtureDefs(s)["Click"] = schemaObject{"$ref": "#/$defs/Loop"}
			fixtureDefs(s)["Loop"] = schemaObject{"$ref": "#/$defs/Click"}
		}, "reference cycle"},
		{"missing reference", func(s schemaObject) { fixtureDefs(s)["Click"] = schemaObject{"$ref": "#/$defs/Missing"} }, "missing object definition Missing"},
		{"external reference", func(s schemaObject) {
			fixtureDefs(s)["Click"] = schemaObject{"$ref": "https://example.com/schema.json"}
		}, "unsupported reference"},
		{"conditional fields", func(s schemaObject) { fixtureDefs(s)["Click"].(schemaObject)["if"] = schemaObject{} }, "unsupported field extraction keyword if"},
		{"invalid enum", func(s schemaObject) { fixtureDefs(s)["ActionEnum"].(schemaObject)["enum"] = []any{123} }, "invalid type"},
		{"unknown branch", func(s schemaObject) { fixtureProps(s, "ClickV2")["type"] = schemaObject{"enum": []any{"Click"}} }, "missing type const"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := pipelineFixtureSchema()
			tt.change(s)
			dir, path := writePipelineFixture(t, s, pipelineFixtureGo)
			_, _, err := checkPipelineCoverage(dir, path, nil)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v; want %q", err, tt.want)
			}
		})
	}
	t.Run("missing file", func(t *testing.T) {
		_, err := readPipelineSchema(filepath.Join(t.TempDir(), "missing.json"))
		if err == nil {
			t.Fatal("expected missing schema error")
		}
	})
	t.Run("broken JSON", func(t *testing.T) {
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), pipelineFixtureGo)
		_ = dir
		if err := os.WriteFile(path, []byte("{"), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := readPipelineSchema(path)
		if err == nil || !strings.Contains(err.Error(), "parse schema") {
			t.Fatalf("got %v", err)
		}
	})
}

func TestPipelineExclusions(t *testing.T) {
	s := pipelineFixtureSchema()
	fixtureProps(s, "Click")["pressure"] = schemaObject{"type": "integer"}
	dir, path := writePipelineFixture(t, s, pipelineFixtureGo)
	issues, excluded, err := checkPipelineCoverage(dir, path, map[string]string{"action.Click.param.pressure": "Intentional fixture omission"})
	if err != nil || len(issues) != 0 || len(excluded) != 1 || !strings.Contains(excluded[0], "Intentional fixture omission") {
		t.Fatalf("got %v / %v / %v", issues, excluded, err)
	}
	_, _, err = checkPipelineCoverage(dir, path, map[string]string{"action.Click.param.pressure": " "})
	if err == nil || !strings.Contains(err.Error(), "nonempty reason") {
		t.Fatalf("got %v", err)
	}
	issues, _, err = checkPipelineCoverage(dir, path, map[string]string{"action.Click.param.target": "Stale reason"})
	if err != nil || !strings.Contains(issueText(issues), "stale pipeline exclusion") {
		t.Fatalf("got %v / %v", issues, err)
	}
}

func TestPipelineUnsupportedCodec(t *testing.T) {
	source := strings.Replace(pipelineFixtureGo, "return json.Marshal(struct", "return anotherEncoder(struct", 1)
	dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
	_, _, err := checkPipelineCoverage(dir, path, nil)
	if err == nil || !strings.Contains(err.Error(), "unsupported WaitFreezesParam.MarshalJSON codec") {
		t.Fatalf("got %v", err)
	}
}

func TestPipelineCustomIntrinsicFields(t *testing.T) {
	s := pipelineFixtureSchema()
	fixtureDefs(s)["CustomActionSchema"] = schemaObject{"$ref": "./custom.action.schema.json"}
	fixtureDefs(s)["Click"].(schemaObject)["allOf"] = []any{schemaObject{"$ref": "#/$defs/CustomActionSchema"}}
	dir, path := writePipelineFixture(t, s, pipelineFixtureGo)
	issues, _, err := checkPipelineCoverage(dir, path, nil)
	if err != nil || len(issues) != 0 {
		t.Fatalf("got %v / %v", issues, err)
	}
}

func replaceFixtureMarshal(source, body string) string {
	start := strings.Index(source, "func (w WaitFreezesParam) MarshalJSON()")
	end := strings.Index(source, "func (w *WaitFreezesParam) UnmarshalJSON")
	return source[:start] + "func (w WaitFreezesParam) MarshalJSON() ([]byte,error) {\n" + body + "\n}\n" + source[end:]
}

const pipelineUnusedClosureMarshal = `
 correct := struct { Time int ` + "`json:\"time\"`" + ` }{}
 unused := func() { _, _ = json.Marshal(correct) }
 _ = unused
 return anotherEncoder(struct { Wrong int ` + "`json:\"wrong\"`" + ` }{})
`

const pipelineIgnoredHelperMarshal = `
 correct := struct { Time int ` + "`json:\"time\"`" + ` }{}
 _, _ = json.Marshal(correct)
 return anotherEncoder(struct { Wrong int ` + "`json:\"wrong\"`" + ` }{})
`

func TestPipelineMarshalReturnTracing(t *testing.T) {
	tests := []struct{ name, body, want string }{
		{"unused closure with correct DTO", pipelineUnusedClosureMarshal, "expected a direct JSON helper call"},
		{"ignored helper with correct DTO", pipelineIgnoredHelperMarshal, "expected a direct JSON helper call"},
		{"returned bytes variable unsupported", `correct := struct { Time int ` + "`json:\"time\"`" + ` }{}; bytes, err := json.Marshal(correct); return bytes, err`, "expected a direct JSON helper call"},
		{"unused closure with wrong DTO ignored", `unused := func() { _, _ = json.Marshal(struct { Wrong int ` + "`json:\"wrong\"`" + ` }{}) }; _ = unused; return json.Marshal(struct { Time int ` + "`json:\"time\"`" + ` }{})`, ""},
		{"ignored helper with wrong DTO ignored", `_, _ = json.Marshal(struct { Wrong int ` + "`json:\"wrong\"`" + ` }{}); return json.Marshal(struct { Time int ` + "`json:\"time\"`" + ` }{})`, ""},
		{"unused closure cannot mask actual returned DTO", `unused := func() { _, _ = json.Marshal(struct { Time int ` + "`json:\"time\"`" + ` }{}) }; _ = unused; return json.Marshal(struct { Wrong int ` + "`json:\"wrong\"`" + ` }{})`, "MarshalJSON field wrong absent from UnmarshalJSON"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, path := writePipelineFixture(t, pipelineFixtureSchema(), replaceFixtureMarshal(pipelineFixtureGo, tt.body))
			issues, _, err := checkPipelineCoverage(dir, path, nil)
			if tt.want == "" {
				if err != nil || len(issues) != 0 {
					t.Fatalf("got %v / %v", issues, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v; want %q", err, tt.want)
			}
		})
	}
}

func TestPipelineInheritedCodecs(t *testing.T) {
	tests := []struct{ name, old, replacement, want string }{
		{"alias to custom codec", "type ClickParam struct { Target string `json:\"target\"` }", `type ClickParam = ClickWire; type ClickWire struct { Target string ` + "`json:\"target\"`" + ` }; func (p ClickWire) MarshalJSON()([]byte,error){return json.Marshal(struct { Wrong int ` + "`json:\"wrong\"`" + ` }{})}`, "unsupported alias inheriting a custom JSON codec"},
		{"anonymous embedded custom codec", "type ClickParam struct { Target string `json:\"target\"` }", `type ClickParam struct { ClickWire; Target string ` + "`json:\"target\"`" + ` }; type ClickWire struct{}; func (p ClickWire) MarshalJSON()([]byte,error){return json.Marshal(struct { Wrong int ` + "`json:\"wrong\"`" + ` }{})}`, "unsupported anonymous embedding promoting a JSON codec"},
		{"local alias preserves methods", "type NoMethod WaitFreezesParam", "type NoMethod = WaitFreezesParam", "unsupported anonymous embedding promoting a JSON codec"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := strings.Replace(pipelineFixtureGo, tt.old, tt.replacement, 1)
			dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
			_, _, err := checkPipelineCoverage(dir, path, nil)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v; want %q", err, tt.want)
			}
		})
	}
}
