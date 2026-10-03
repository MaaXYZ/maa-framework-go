package checker

import (
	"encoding/json"
	"strings"
	"testing"
)

func replaceFixtureUnmarshal(source, body string) string {
	start := strings.Index(source, "func (w *WaitFreezesParam) UnmarshalJSON")
	if start < 0 {
		return source
	}
	return source[:start] + "func (w *WaitFreezesParam) UnmarshalJSON(data []byte) error {\n" + body + "\n}\n"
}

// TestPipelineCodecLexicalShadowing rejects same-name local types and variables
// instead of resolving them through a flat name map. Both directions must fail
// extraction rather than pass with the wrong DTO.
func TestPipelineCodecLexicalShadowing(t *testing.T) {
	const wrongDTO = "type DTO struct { Wrong int `json:\"wrong\"` }"
	const timeDTO = "type DTO struct { Time int `json:\"time\"` }"
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "marshal shadowed local type",
			source: replaceFixtureMarshal(pipelineFixtureGo, wrongDTO+"; { "+timeDTO+"; _ = DTO{} }; return json.Marshal(DTO{})"),
			want:   "unsupported shadowed local type DTO",
		},
		{
			name:   "unmarshal shadowed local type",
			source: replaceFixtureUnmarshal(pipelineFixtureGo, wrongDTO+"; { "+timeDTO+"; _ = DTO{} }; var raw DTO; return json.Unmarshal(data, &raw)"),
			want:   "unsupported shadowed local type DTO",
		},
		{
			name:   "marshal block shadowed wire variable",
			source: replaceFixtureMarshal(pipelineFixtureGo, "raw := struct { Time int `json:\"time\"` }{}; { raw := struct { Wrong int `json:\"wrong\"` }{}; _ = raw }; return json.Marshal(raw)"),
			want:   "reassigned or shadowed wire variable raw",
		},
		{
			name:   "unmarshal block shadowed wire variable",
			source: replaceFixtureUnmarshal(pipelineFixtureGo, "raw := struct { Time int `json:\"time\"` }{}; { raw := struct { Wrong int `json:\"wrong\"` }{}; _ = raw }; return json.Unmarshal(data, &raw)"),
			want:   "reassigned or shadowed wire variable raw",
		},
		{
			name:   "variable reuses a local type name",
			source: replaceFixtureUnmarshal(pipelineFixtureGo, wrongDTO+"\n var raw DTO\n { DTO := struct { Time int `json:\"time\"` }{}; _ = DTO }\n if err := json.Unmarshal(data, &raw); err != nil { return err }\n *w = WaitFreezesParam{}\n return nil"),
			want:   "unsupported shadowed local type DTO (reused as a variable)",
		},
		{
			name: "inner type shadows a package type",
			source: replaceFixtureUnmarshal(
				strings.Replace(pipelineFixtureGo, "type WaitFreezesParam struct", "type WaitFreezesWire struct { Wrong int `json:\"wrong\"` }\ntype WaitFreezesParam struct", 1),
				"{ type WaitFreezesWire struct { Time int `json:\"time\"` }; _ = WaitFreezesWire{} }\n var raw WaitFreezesWire\n if err := json.Unmarshal(data, &raw); err != nil { return err }\n *w = WaitFreezesParam{}\n return nil"),
			want: "unsupported shadowed local type WaitFreezesWire (shadows a package type)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.source == pipelineFixtureGo {
				t.Fatalf("fixture mutation did not apply")
			}
			dir, path := writePipelineFixture(t, pipelineFixtureSchema(), tt.source)
			_, _, err := checkPipelineCoverage(dir, path, nil)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v; want %q", err, tt.want)
			}
		})
	}
}

func TestPipelineCodecMultipleDTOs(t *testing.T) {
	t.Run("consistent unmarshal DTOs", func(t *testing.T) {
		source := replaceFixtureUnmarshal(pipelineFixtureGo, "type NoMethod WaitFreezesParam\n var raw struct { NoMethod; Time int `json:\"time\"` }\n _ = json.Unmarshal(data, &raw)\n if err := json.Unmarshal(data, &raw); err != nil { return err }\n *w = WaitFreezesParam(raw.NoMethod)\n return nil")
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
	t.Run("conflicting unmarshal DTOs", func(t *testing.T) {
		source := replaceFixtureUnmarshal(pipelineFixtureGo, "type NoMethod WaitFreezesParam\n _ = json.Unmarshal(data, &struct { NoMethod; Time int `json:\"time\"` }{})\n if err := json.Unmarshal(data, &struct { NoMethod; Wrong int `json:\"wrong\"` }{}); err != nil { return err }\n *w = WaitFreezesParam{}\n return nil")
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		_, _, err := checkPipelineCoverage(dir, path, nil)
		if err == nil || !strings.Contains(err.Error(), "multiple JSON wire DTOs with different fields") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("consistent marshal returns", func(t *testing.T) {
		source := replaceFixtureMarshal(pipelineFixtureGo, "if true { return json.Marshal(struct { Time int `json:\"time\"` }{}) }\n return json.Marshal(struct { Time int `json:\"time\"` }{})")
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
	t.Run("conflicting marshal returns", func(t *testing.T) {
		source := replaceFixtureMarshal(pipelineFixtureGo, "if true { return json.Marshal(struct { Time int `json:\"time\"` }{}) }\n return json.Marshal(struct { Wrong int `json:\"wrong\"` }{})")
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		_, _, err := checkPipelineCoverage(dir, path, nil)
		if err == nil || !strings.Contains(err.Error(), "multiple JSON wire DTOs with different fields") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("unmarshal never assigns the receiver", func(t *testing.T) {
		source := replaceFixtureUnmarshal(pipelineFixtureGo, "var probe struct { Time int `json:\"time\"` }\n _ = json.Unmarshal(data, &probe)\n return nil")
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		_, _, err := checkPipelineCoverage(dir, path, nil)
		if err == nil || !strings.Contains(err.Error(), "UnmarshalJSON never assigns its receiver") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("value receiver cannot decode", func(t *testing.T) {
		source := strings.Replace(pipelineFixtureGo, "func (w *WaitFreezesParam) UnmarshalJSON(data []byte) error {", "func (w WaitFreezesParam) UnmarshalJSON(data []byte) error {", 1)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		_, _, err := checkPipelineCoverage(dir, path, nil)
		if err == nil || !strings.Contains(err.Error(), "UnmarshalJSON must use a pointer receiver") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("new wire value and pointer receiver field", func(t *testing.T) {
		source := replaceFixtureUnmarshal(pipelineFixtureGo, "type NoMethod WaitFreezesParam\n raw := new(struct { NoMethod; Time int `json:\"time\"` })\n if err := json.Unmarshal(data, raw); err != nil { return err }\n (*w).Time = raw.Time\n return nil")
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
}

// TestPipelineJSONTagSemantics pins the encoding/json behavior the checker
// models: only an exact "-" ignores a field, while "-," names it "-" and
// ",option" keeps the Go field name.
func TestPipelineJSONTagSemantics(t *testing.T) {
	type tags struct {
		Ignored int `json:"-"`
		Dash    int `json:"-,"`
		Named   int `json:"named,omitempty"`
		Default int `json:",omitempty"`
	}
	data, err := json.Marshal(tags{Ignored: 1, Dash: 2, Named: 3, Default: 4})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `{"-":2,"named":3,"Default":4}`; got != want {
		t.Fatalf("encoding/json roundtrip mismatch: got %s; want %s", got, want)
	}
	var decoded tags
	if err := json.Unmarshal([]byte(`{"-":7,"named":8,"Default":9}`), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Ignored != 0 || decoded.Dash != 7 || decoded.Named != 8 || decoded.Default != 9 {
		t.Fatalf("unexpected decoding: %+v", decoded)
	}
}

func TestPipelineTagCoverage(t *testing.T) {
	t.Run("dash-comma names the field", func(t *testing.T) {
		source := strings.Replace(pipelineFixtureGo, "type ClickParam struct {", "type ClickParam struct { Extra int `json:\"-,\"`;", 1)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if text := issueText(issues); !strings.Contains(text, "action.Click.param.-: Go field absent from schema (ClickParam,") {
			t.Fatalf("got %s", text)
		}
	})
	t.Run("exact dash still ignores the field", func(t *testing.T) {
		source := strings.Replace(pipelineFixtureGo, "type ClickParam struct {", "type ClickParam struct { Ignored int `json:\"-\"`;", 1)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
	t.Run("empty name option uses the Go field name", func(t *testing.T) {
		source := strings.Replace(pipelineFixtureGo, "Target string `json:\"target\"`", "Target string `json:\",omitempty\"`", 1)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if text := issueText(issues); !strings.Contains(text, "action.Click.param.Target: Go field absent from schema (ClickParam,") {
			t.Fatalf("got %s", text)
		}
	})
}
