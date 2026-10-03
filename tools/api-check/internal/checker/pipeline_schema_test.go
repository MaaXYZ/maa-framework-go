package checker

import (
	"strings"
	"testing"
)

func TestPipelineIgnoredRefSiblings(t *testing.T) {
	for _, name := range sortedPipelineKeys(pipelineIgnoredDefs) {
		for _, shape := range []string{"reference only", "properties", "allOf"} {
			t.Run(name+"/"+shape, func(t *testing.T) {
				s := pipelineFixtureSchema()
				defs := fixtureDefs(s)
				// The referenced extension stays opaque, including its external ref.
				defs[name] = schemaObject{"$ref": "./extension.schema.json"}
				param := schemaObject{"$ref": "#/$defs/" + name}
				props := schemaObject{"pressure": schemaObject{"type": "integer"}}
				switch shape {
				case "properties":
					param["properties"] = props
				case "allOf":
					param["allOf"] = []any{schemaObject{"properties": props}}
				}
				defs["DirectHit"] = param
				dir, path := writePipelineFixture(t, s, pipelineFixtureGo)
				issues, excluded, err := checkPipelineCoverage(dir, path, nil)
				if err != nil || len(excluded) != 0 {
					t.Fatalf("got %v / %v / %v", issues, excluded, err)
				}
				if shape == "reference only" {
					if len(issues) != 0 {
						t.Fatalf("unexpected issues: %s", issueText(issues))
					}
				} else if len(issues) != 1 || issues[0].message != "recognition.DirectHit.param.pressure: missing Go field (DirectHitParam)" {
					t.Fatalf("expected missing sibling pressure field; got %s", issueText(issues))
				}
			})
		}
	}
}

func TestPipelineIgnoredRefSiblingErrors(t *testing.T) {
	type testCase struct {
		name, key string
		value     any
		want      string
	}
	tests := []testCase{
		{"invalid properties", "properties", []any{}, "properties is not an object"},
		{"invalid property", "properties", schemaObject{"pressure": false}, "property is not an object"},
		{"empty allOf", "allOf", []any{}, "allOf is not a nonempty array"},
		{"invalid allOf branch", "allOf", []any{false}, "allOf[0]: not an object"},
		{"cyclic allOf sibling", "allOf", []any{schemaObject{"$ref": "#/$defs/DirectHit"}}, "reference cycle"},
		{"anyOf fields", "anyOf", []any{schemaObject{"properties": schemaObject{}}}, "unsupported anyOf field alternatives"},
		{"oneOf fields", "oneOf", []any{schemaObject{"properties": schemaObject{}}}, "unsupported oneOf field alternatives"},
	}
	for _, key := range []string{"$dynamicRef", "$recursiveRef", "if", "then", "else", "not", "dependentSchemas", "patternProperties"} {
		tests = append(tests, testCase{key, key, schemaObject{}, "unsupported field extraction keyword " + key})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := pipelineFixtureSchema()
			fixtureDefs(s)["DirectHit"] = schemaObject{"$ref": "#/$defs/jsonComments", tt.key: tt.value}
			dir, path := writePipelineFixture(t, s, pipelineFixtureGo)
			_, _, err := checkPipelineCoverage(dir, path, nil)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v; want %q", err, tt.want)
			}
		})
	}
}

func TestPipelineNodeShape(t *testing.T) {
	tests := []struct {
		name  string
		apply func(schemaObject)
		want  string
	}{
		{"missing allowed reference", func(s schemaObject) {
			fixtureDefs(s)["Node"].(schemaObject)["allOf"] = []any{schemaObject{"$ref": "#/$defs/RecognitionFormat"}}
		}, "unresolvable allOf reference #/$defs/RecognitionFormat"},
		{"duplicate allowed reference", func(s schemaObject) {
			fixtureDefs(s)["Node"].(schemaObject)["allOf"] = []any{schemaObject{"$ref": "#/$defs/jsonComments"}, schemaObject{"$ref": "#/$defs/jsonComments"}}
		}, "duplicate allOf reference #/$defs/jsonComments"},
		{"unsupported reference", func(s schemaObject) {
			fixtureDefs(s)["Node"].(schemaObject)["allOf"] = []any{schemaObject{"$ref": "#/$defs/Unknown"}}
		}, "unsupported allOf reference #/$defs/Unknown"},
		{"malformed branch", func(s schemaObject) {
			fixtureDefs(s)["Node"].(schemaObject)["allOf"] = []any{"not-an-object"}
		}, "unsupported allOf branch"},
		{"branch with siblings", func(s schemaObject) {
			fixtureDefs(s)["Node"].(schemaObject)["allOf"] = []any{schemaObject{"$ref": "#/$defs/jsonComments", "properties": schemaObject{}}}
		}, "unsupported allOf branch"},
		{"non-string reference", func(s schemaObject) {
			fixtureDefs(s)["Node"].(schemaObject)["allOf"] = []any{schemaObject{"$ref": 7}}
		}, "unsupported allOf reference 7"},
		{"empty allOf", func(s schemaObject) {
			fixtureDefs(s)["Node"].(schemaObject)["allOf"] = []any{}
		}, "allOf is not a nonempty array"},
		{"pattern properties", func(s schemaObject) {
			fixtureDefs(s)["Node"].(schemaObject)["patternProperties"] = schemaObject{"^x": schemaObject{"type": "string"}}
		}, "unsupported field extraction keyword patternProperties"},
		{"recursive reference", func(s schemaObject) {
			fixtureDefs(s)["Node"].(schemaObject)["$recursiveRef"] = "#"
		}, "unsupported field extraction keyword $recursiveRef"},
		{"direct reference", func(s schemaObject) {
			fixtureDefs(s)["Node"].(schemaObject)["$ref"] = "#/$defs/Node"
		}, "unsupported field extraction keyword $ref"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := pipelineFixtureSchema()
			tt.apply(s)
			dir, path := writePipelineFixture(t, s, pipelineFixtureGo)
			_, _, err := checkPipelineCoverage(dir, path, nil)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v; want %q", err, tt.want)
			}
		})
	}
	t.Run("valid allOf references", func(t *testing.T) {
		s := pipelineFixtureSchema()
		defs := fixtureDefs(s)
		defs["RecognitionFormat"] = schemaObject{"type": "object", "properties": schemaObject{}}
		defs["ActionFormat"] = schemaObject{"type": "object", "properties": schemaObject{}}
		defs["Node"].(schemaObject)["allOf"] = []any{
			schemaObject{"$ref": "#/$defs/jsonComments"},
			schemaObject{"$ref": "#/$defs/RecognitionFormat"},
			schemaObject{"$ref": "#/$defs/ActionFormat"},
		}
		dir, path := writePipelineFixture(t, s, pipelineFixtureGo)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
}

func TestPipelineEnumBranchErrors(t *testing.T) {
	tests := []struct {
		name  string
		apply func(schemaObject)
		want  string
	}{
		{"duplicate enum value", func(s schemaObject) {
			fixtureDefs(s)["ActionEnum"].(schemaObject)["enum"] = []any{"Click", "Click"}
		}, "duplicate type Click"},
		{"empty enum", func(s schemaObject) {
			fixtureDefs(s)["ActionEnum"].(schemaObject)["enum"] = []any{}
		}, "missing nonempty enum"},
		{"branch absent from enum", func(s schemaObject) {
			defs := fixtureDefs(s)
			defs["FutureV2"] = schemaObject{"properties": schemaObject{"type": schemaObject{"const": "Future"}, "param": schemaObject{"properties": schemaObject{}}}}
			wrapper := fixtureProps(s, "ActionV2")["action"].(schemaObject)
			wrapper["anyOf"] = append(wrapper["anyOf"].([]any), schemaObject{"$ref": "#/$defs/FutureV2"})
		}, "type Future absent from ActionEnum"},
		{"duplicate branch", func(s schemaObject) {
			wrapper := fixtureProps(s, "ActionV2")["action"].(schemaObject)
			wrapper["anyOf"] = append(wrapper["anyOf"].([]any), schemaObject{"$ref": "#/$defs/ClickV2"})
		}, "duplicate branch for Click"},
		{"branch without properties", func(s schemaObject) {
			fixtureDefs(s)["ClickV2"] = schemaObject{"properties": false}
		}, "missing branch properties"},
		{"branch without param", func(s schemaObject) {
			fixtureDefs(s)["ClickV2"] = schemaObject{"properties": schemaObject{"type": schemaObject{"const": "Click"}}}
		}, "missing param object"},
		{"missing envelope properties", func(s schemaObject) {
			fixtureProps(s, "ActionV2")["action"] = schemaObject{"anyOf": []any{schemaObject{"$ref": "#/$defs/ClickV2"}}}
		}, "missing envelope properties"},
		{"invalid envelope property", func(s schemaObject) {
			envelope := fixtureProps(s, "ActionV2")["action"].(schemaObject)["properties"].(schemaObject)
			envelope["type"] = false
		}, "envelope property is not an object"},
		{"missing envelope param", func(s schemaObject) {
			envelope := fixtureProps(s, "ActionV2")["action"].(schemaObject)["properties"].(schemaObject)
			delete(envelope, "param")
		}, "missing param envelope property"},
		{"envelope oneOf alternative", func(s schemaObject) {
			fixtureProps(s, "ActionV2")["action"].(schemaObject)["oneOf"] = []any{schemaObject{"$ref": "#/$defs/ClickV2"}}
		}, "unsupported envelope keyword oneOf"},
		{"envelope allOf adds fields", func(s schemaObject) {
			fixtureDefs(s)["ExtraEnv"] = schemaObject{"properties": schemaObject{"ghost": schemaObject{"type": "string"}}}
			fixtureProps(s, "ActionV2")["action"].(schemaObject)["allOf"] = []any{schemaObject{"$ref": "#/$defs/ExtraEnv"}}
		}, "unsupported envelope allOf reference #/$defs/ExtraEnv"},
		{"branch extra property", func(s schemaObject) {
			fixtureDefs(s)["ClickV2"].(schemaObject)["properties"].(schemaObject)["ghost"] = schemaObject{"type": "string"}
		}, "unsupported branch property ghost"},
		{"branch with field alternative", func(s schemaObject) {
			fixtureDefs(s)["ClickV2"].(schemaObject)["allOf"] = []any{schemaObject{"properties": schemaObject{"ghost": schemaObject{"type": "string"}}}}
		}, "unsupported branch keyword allOf"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := pipelineFixtureSchema()
			tt.apply(s)
			dir, path := writePipelineFixture(t, s, pipelineFixtureGo)
			_, _, err := checkPipelineCoverage(dir, path, nil)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v; want %q", err, tt.want)
			}
		})
	}
	t.Run("duplicate branch is rejected before coverage", func(t *testing.T) {
		s := pipelineFixtureSchema()
		wrapper := fixtureProps(s, "ActionV2")["action"].(schemaObject)
		wrapper["anyOf"] = []any{schemaObject{"$ref": "#/$defs/ClickV2"}, schemaObject{"$ref": "#/$defs/ClickV2"}}
		dir, path := writePipelineFixture(t, s, pipelineFixtureGo)
		if _, _, err := checkPipelineCoverage(dir, path, nil); err == nil || !strings.Contains(err.Error(), "duplicate branch for Click") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("metadata allOf and dependentSchemas are accepted", func(t *testing.T) {
		s := pipelineFixtureSchema()
		defs := fixtureDefs(s)
		defs["TypeWithDependent"] = schemaObject{"anyOf": []any{schemaObject{"required": []any{"param"}}}}
		fixtureProps(s, "ActionV2")["action"].(schemaObject)["allOf"] = []any{schemaObject{"$ref": "#/$defs/jsonComments"}}
		defs["ClickV2"].(schemaObject)["dependentSchemas"] = schemaObject{"type": schemaObject{"$ref": "#/$defs/TypeWithDependent"}}
		dir, path := writePipelineFixture(t, s, pipelineFixtureGo)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
	t.Run("dependent schema cannot add fields", func(t *testing.T) {
		s := pipelineFixtureSchema()
		defs := fixtureDefs(s)
		defs["ExtraDependent"] = schemaObject{"properties": schemaObject{"ghost": schemaObject{"type": "string"}}}
		defs["ClickV2"].(schemaObject)["dependentSchemas"] = schemaObject{"type": schemaObject{"$ref": "#/$defs/ExtraDependent"}}
		dir, path := writePipelineFixture(t, s, pipelineFixtureGo)
		_, _, err := checkPipelineCoverage(dir, path, nil)
		if err == nil || !strings.Contains(err.Error(), "dependent schema must only require existing fields") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("dependent schema must reference a definition", func(t *testing.T) {
		s := pipelineFixtureSchema()
		fixtureDefs(s)["ClickV2"].(schemaObject)["dependentSchemas"] = schemaObject{"type": schemaObject{"required": []any{"param"}}}
		dir, path := writePipelineFixture(t, s, pipelineFixtureGo)
		_, _, err := checkPipelineCoverage(dir, path, nil)
		if err == nil || !strings.Contains(err.Error(), "unsupported reference") {
			t.Fatalf("got %v", err)
		}
	})
}
