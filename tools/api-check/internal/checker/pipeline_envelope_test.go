package checker

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestPipelineEnvelopeRoundtrip mirrors the fixture DTOs through encoding/json
// so the checker's static envelope assumptions (type/param and the inline
// recognition wrapper) are validated against real wire behavior.
func TestPipelineEnvelopeRoundtrip(t *testing.T) {
	type actionEnvelope struct {
		Type  string          `json:"type,omitempty"`
		Param json.RawMessage `json:"param,omitempty"`
	}
	encoded, err := json.Marshal(actionEnvelope{Type: "Click", Param: json.RawMessage(`{"target":[1,2,3,4]}`)})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(encoded), `{"type":"Click","param":{"target":[1,2,3,4]}}`; got != want {
		t.Fatalf("envelope encoding mismatch: got %s; want %s", got, want)
	}
	var decoded actionEnvelope
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Type != "Click" || string(decoded.Param) != `{"target":[1,2,3,4]}` {
		t.Fatalf("unexpected envelope decoding: %+v", decoded)
	}

	type recognition struct {
		Type  string          `json:"type,omitempty"`
		Param json.RawMessage `json:"param,omitempty"`
	}
	type inlineSub struct {
		SubName     string      `json:"sub_name,omitempty"`
		Recognition recognition `json:"recognition"`
	}
	encoded, err = json.Marshal(inlineSub{SubName: "sub1", Recognition: recognition{Type: "DirectHit"}})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(encoded), `{"sub_name":"sub1","recognition":{"type":"DirectHit"}}`; got != want {
		t.Fatalf("inline encoding mismatch: got %s; want %s", got, want)
	}
	var inline inlineSub
	if err := json.Unmarshal(encoded, &inline); err != nil {
		t.Fatal(err)
	}
	if inline.SubName != "sub1" || inline.Recognition.Type != "DirectHit" {
		t.Fatalf("unexpected inline decoding: %+v", inline)
	}
}

// TestPipelineEnvelopeFields requires the Action and Recognition envelope to
// expose type/param consistently in both encoding and decoding.
func TestPipelineEnvelopeFields(t *testing.T) {
	t.Run("encoding only rename is a direction mismatch", func(t *testing.T) {
		source := strings.Replace(pipelineFixtureGo, "type Action struct { Type ActionType `json:\"type\"`", "type Action struct { Type ActionType `json:\"wrong\"`", 1)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		_, _, err := checkPipelineCoverage(dir, path, nil)
		if err == nil || !strings.Contains(err.Error(), "JSON field wrong absent from UnmarshalJSON wire DTO") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("decoding only rename is a direction mismatch", func(t *testing.T) {
		source := strings.Replace(pipelineFixtureGo, "var raw struct { Type ActionType `json:\"type\"`", "var raw struct { Type ActionType `json:\"wrong\"`", 1)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		_, _, err := checkPipelineCoverage(dir, path, nil)
		if err == nil || !strings.Contains(err.Error(), "JSON field type absent from UnmarshalJSON wire DTO") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("recognition envelope rename in both directions", func(t *testing.T) {
		source := strings.Replace(pipelineFixtureGo, "Type RecognitionType `json:\"type\"`", "Type RecognitionType `json:\"wrong\"`", 2)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil {
			t.Fatal(err)
		}
		text := issueText(issues)
		if !strings.Contains(text, "recognition.envelope.type: missing Go field (Recognition)") || !strings.Contains(text, "recognition.envelope.wrong: Go field absent from schema (Recognition,") {
			t.Fatalf("got %s", text)
		}
	})
	t.Run("decoy envelope DTO is rejected", func(t *testing.T) {
		source := replaceFixtureFunc(pipelineFixtureGo, "func (a *Action) UnmarshalJSON(data []byte) error {", `func (a *Action) UnmarshalJSON(data []byte) error {
 var decoy struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 unmarshalJSON(data, &decoy)
 var raw struct { Type ActionType `+"`json:\"wrong\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 unmarshalJSON(data, &raw)
 param, _ := decodeActionParam(raw.Type, raw.Param)
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		_, _, err := checkPipelineCoverage(dir, path, nil)
		if err == nil || !strings.Contains(err.Error(), "multiple JSON wire DTOs with different fields") {
			t.Fatalf("got %v", err)
		}
	})
}

// TestPipelineInlineSubRecognitionCoverage validates the optional
// SubRecognitionInline inventory: sub_name plus the wrapped recognition
// envelope, while extension and metadata exclusions keep working.
func TestPipelineInlineSubRecognitionCoverage(t *testing.T) {
	t.Run("absent schema definition is tolerated", func(t *testing.T) {
		s := pipelineFixtureSchema()
		delete(fixtureDefs(s), "SubRecognitionInline")
		dir, path := writePipelineFixture(t, s, pipelineFixtureGo)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
	t.Run("renamed sub_name fails both directions", func(t *testing.T) {
		source := strings.ReplaceAll(pipelineFixtureGo, "`json:\"sub_name,omitempty\"`", "`json:\"sub\"`")
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil {
			t.Fatal(err)
		}
		text := issueText(issues)
		if !strings.Contains(text, "SubRecognitionInline.sub_name: missing Go field (InlineSubRecognition)") || !strings.Contains(text, "SubRecognitionInline.sub: Go field absent from schema (InlineSubRecognition,") {
			t.Fatalf("got %s", text)
		}
	})
	t.Run("missing recognition envelope", func(t *testing.T) {
		source := strings.Replace(pipelineFixtureGo, "; Recognition Recognition `json:\"recognition\"`", "", 1)
		source = strings.Replace(source, "; Recognition json.RawMessage `json:\"recognition,omitempty\"`", "", 1)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if text := issueText(issues); !strings.Contains(text, "SubRecognitionInline.recognition: missing inline recognition envelope") {
			t.Fatalf("got %s", text)
		}
	})
	t.Run("renamed recognition envelope", func(t *testing.T) {
		source := strings.Replace(pipelineFixtureGo, "`json:\"recognition\"`", "`json:\"other\"`", 1)
		source = strings.Replace(source, "`json:\"recognition,omitempty\"`", "`json:\"other,omitempty\"`", 1)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if text := issueText(issues); !strings.Contains(text, "SubRecognitionInline.other: Go field absent from schema (InlineSubRecognition,") {
			t.Fatalf("got %s", text)
		}
	})
	t.Run("schema metadata and custom hook refs stay ignored", func(t *testing.T) {
		s := pipelineFixtureSchema()
		defs := fixtureDefs(s)
		defs["CustomRecognitionSchema"] = schemaObject{"$ref": "./custom.recognition.schema.json"}
		defs["SubRecognitionInline"] = schemaObject{
			"allOf":      []any{schemaObject{"$ref": "#/$defs/jsonComments"}, schemaObject{"$ref": "#/$defs/CustomRecognitionSchema"}},
			"properties": schemaObject{"sub_name": schemaObject{"type": "string"}},
		}
		dir, path := writePipelineFixture(t, s, pipelineFixtureGo)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
}
