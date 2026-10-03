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
