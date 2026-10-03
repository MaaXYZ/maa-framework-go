package checker

import (
	"strings"
	"testing"
)

// replaceFixtureFunc replaces one top-level fixture function. The fixture keeps
// its closing braces at column zero so the first "\n}\n" ends the function.
func replaceFixtureFunc(source, signature, replacement string) string {
	start := strings.Index(source, signature)
	if start < 0 {
		return source
	}
	end := strings.Index(source[start:], "\n}\n")
	if end < 0 {
		return source
	}
	return source[:start] + replacement + "\n" + source[start+end+3:]
}

// TestPipelineDecoderFlowRegressions locks in that typed decoder coverage comes
// from the real UnmarshalJSON data flow. Every previously accepted false pass
// (discarded helper results, unrelated helpers, fixed discriminants, missing or
// cyclic helpers, conflicting mappings) must now fail extraction or report the
// missing typed case.
func TestPipelineDecoderFlowRegressions(t *testing.T) {
	const actionUnmarshal = "func (a *Action) UnmarshalJSON(data []byte) error {"
	const recognitionUnmarshal = "func (r *Recognition) UnmarshalJSON(data []byte) error {"
	const helper = "func decodeActionParam(t ActionType, data json.RawMessage) (ActionParam, error) {"
	tests := []struct {
		name   string
		source string
		want   string
		issues []string
	}{
		{
			name: "unused helper result",
			source: replaceFixtureFunc(pipelineFixtureGo, actionUnmarshal, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 _ = decodeActionParam(raw.Type, raw.Param)
 a.Param = &RawActionParam{}
 return nil
}`),
			issues: []string{"action.Click.decoder: missing typed Go decoder case"},
		},
		{
			name: "unrelated helper result",
			source: replaceFixtureFunc(pipelineFixtureGo, actionUnmarshal, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 _ = actionSideEffect(raw.Type)
 a.Param = &RawActionParam{}
 return nil
}
func actionSideEffect(t ActionType) ActionParam {
 switch t { case ActionTypeClick: return &ClickParam{} }
 return nil
}`),
			issues: []string{"action.Click.decoder: missing typed Go decoder case"},
		},
		{
			name: "helper switch on fixed constant",
			source: replaceFixtureFunc(pipelineFixtureGo, helper, `func decodeActionParam(t ActionType, data json.RawMessage) (ActionParam, error) {
 var param ActionParam
 switch ActionTypeClick { case ActionTypeClick: param = &ClickParam{} }
 return param, nil
}`),
			want: "does not switch on its ActionType parameter",
		},
		{
			name: "helper switch on unrelated variable",
			source: replaceFixtureFunc(pipelineFixtureGo, helper, `func decodeActionParam(t ActionType, data json.RawMessage) (ActionParam, error) {
 var param ActionParam
 switch data { case ActionTypeClick: param = &ClickParam{} }
 return param, nil
}`),
			want: "does not switch on its ActionType parameter",
		},
		{
			name: "direct switch on fixed constant",
			source: replaceFixtureFunc(pipelineFixtureGo, recognitionUnmarshal, `func (r *Recognition) UnmarshalJSON(data []byte) error {
 var raw struct { Type RecognitionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 switch RecognitionTypeDirectHit { case RecognitionTypeDirectHit: r.Param = &DirectHitParam{} }
 return nil
}`),
			issues: []string{"recognition.DirectHit.decoder: missing typed Go decoder case"},
		},
		{
			name:   "missing helper declaration",
			source: strings.Replace(pipelineFixtureGo, "func decodeActionParam(t ActionType", "func decodeActionParamRenamed(t ActionType", 1),
			want:   "unsupported action decoder helper decodeActionParam (missing declaration)",
		},
		{
			name: "cyclic helper",
			source: replaceFixtureFunc(pipelineFixtureGo, helper, `func decodeActionParam(t ActionType, data json.RawMessage) (ActionParam, error) {
 var param ActionParam
 switch t {
 case ActionTypeClick: param, _ = decodeActionParam(t, data)
 }
 return param, nil
}`),
			want: "cyclic action decoder helper decodeActionParam",
		},
		{
			name: "switch on undecoded receiver type",
			source: replaceFixtureFunc(pipelineFixtureGo, actionUnmarshal, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 switch a.Type {
 case ActionTypeClick: a.Param = &ClickParam{}
 }
 return nil
}`),
			issues: []string{"action.Click.decoder: missing typed Go decoder case"},
		},
		{
			name: "helper not called with decoded type",
			source: replaceFixtureFunc(pipelineFixtureGo, actionUnmarshal, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 param, err := decodeActionParam(ActionTypeClick, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`),
			want: "is not called with the decoded ActionType",
		},
		{
			name: "conflicting decoder parameters",
			source: replaceFixtureFunc(pipelineFixtureGo, recognitionUnmarshal, `func (r *Recognition) UnmarshalJSON(data []byte) error {
 var raw struct { Type RecognitionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 switch raw.Type {
 case RecognitionTypeDirectHit: r.Param = &DirectHitParam{}
 case RecognitionTypeDirectHit: r.Param = &ClickParam{}
 }
 return nil
}`),
			want: "conflicting decoder parameters",
		},
		{
			name: "case shadows returned helper variable",
			source: replaceFixtureFunc(pipelineFixtureGo, helper, `func decodeActionParam(t ActionType, data json.RawMessage) (ActionParam, error) {
 var param ActionParam
 switch t {
 case ActionTypeClick: param := &ClickParam{}; _ = param
 }
 return param, nil
}`),
			want: "ambiguous shadowed action decoder variable param",
		},
		{
			name: "case shadows direct decoder variable",
			source: replaceFixtureFunc(pipelineFixtureGo, recognitionUnmarshal, `func (r *Recognition) UnmarshalJSON(data []byte) error {
 var raw struct { Type RecognitionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 var param RecognitionParam
 switch raw.Type {
 case RecognitionTypeDirectHit: param := &DirectHitParam{}; _ = param
 }
 r.Param = param
 return nil
}`),
			want: "ambiguous shadowed recognition decoder variable param",
		},
		{
			name:   "ambiguous shadowed decoder variable",
			source: strings.Replace(pipelineFixtureGo, "*a = Action{Type: raw.Type, Param: param}", "param := 1\n _ = param\n *a = Action{Type: raw.Type, Param: param}", 1),
			want:   "ambiguous shadowed action decoder variable param",
		},
		{
			name: "case coverage overwritten by later assignment",
			source: replaceFixtureFunc(pipelineFixtureGo, recognitionUnmarshal, `func (r *Recognition) UnmarshalJSON(data []byte) error {
 var raw struct { Type RecognitionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 switch raw.Type { case RecognitionTypeDirectHit: r.Param = &DirectHitParam{} }
 r.Type = raw.Type
 r.Param = &RawActionParam{}
 return nil
}`),
			want: "ambiguous recognition decoder cases overwritten by the receiver Param assignment",
		},
		{
			name: "case assigns an unconsumed variable",
			source: replaceFixtureFunc(pipelineFixtureGo, actionUnmarshal, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 var unused ActionParam
 switch raw.Type { case ActionTypeClick: unused = &ClickParam{} }
 _ = unused
 a.Type = raw.Type
 a.Param = &RawActionParam{}
 return nil
}`),
			issues: []string{"action.Click.decoder: missing typed Go decoder case"},
		},
		{
			name: "discarded helper result composite",
			source: replaceFixtureFunc(pipelineFixtureGo, actionUnmarshal, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 _ = Action{Type: raw.Type, Param: param}
 a.Type = raw.Type
 a.Param = &RawActionParam{}
 return nil
}`),
			issues: []string{"action.Click.decoder: missing typed Go decoder case"},
		},
		{
			name: "helper result overwritten after declaration",
			source: replaceFixtureFunc(pipelineFixtureGo, actionUnmarshal, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 param = &RawActionParam{}
 a.Type = raw.Type
 a.Param = param
 return nil
}`),
			want: "assigned more than once",
		},
		{
			name: "envelope discriminant decoded from param",
			source: replaceFixtureFunc(pipelineFixtureGo, actionUnmarshal, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 var inner struct { Type ActionType `+"`json:\"type\"`"+` }
 unmarshalJSON(raw.Param, &inner)
 switch inner.Type { case ActionTypeClick: a.Param = &ClickParam{} }
 return nil
}`),
			want: "multiple JSON wire DTOs with different fields",
		},
		{
			name: "receiver type pinned to a constant",
			source: replaceFixtureFunc(pipelineFixtureGo, recognitionUnmarshal, `func (r *Recognition) UnmarshalJSON(data []byte) error {
 var raw struct { Type RecognitionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 r.Type = raw.Type
 r.Type = RecognitionTypeDirectHit
 switch r.Type { case RecognitionTypeDirectHit: r.Param = &DirectHitParam{} }
 return nil
}`),
			issues: []string{"recognition.DirectHit.decoder: missing typed Go decoder case"},
		},
		{
			name: "helper parameter overwritten outside switch",
			source: replaceFixtureFunc(pipelineFixtureGo, helper, `func decodeActionParam(t ActionType, data json.RawMessage) (ActionParam, error) {
 var param ActionParam
 switch t { case ActionTypeClick: param = &ClickParam{} }
 param = &RawActionParam{}
 return param, nil
}`),
			want: "assigns case parameter param outside its switch",
		},
		{
			name: "unreachable helper switch",
			source: replaceFixtureFunc(pipelineFixtureGo, helper, `func decodeActionParam(t ActionType, data json.RawMessage) (ActionParam, error) {
 var param ActionParam
 return param, nil
 switch t { case ActionTypeClick: param = &ClickParam{} }
}`),
			want: "unreachable action decoder helper decodeActionParam switch",
		},
		{
			name: "helper shadowed by a local closure",
			source: replaceFixtureFunc(pipelineFixtureGo, actionUnmarshal, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 decodeActionParam := func(t ActionType, data json.RawMessage) (ActionParam, error) { return &RawActionParam{}, nil }
 param, _ := decodeActionParam(raw.Type, raw.Param)
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`),
			want: "unsupported action decoder helper decodeActionParam (shadowed by a local variable)",
		},
		{
			name: "receiver alias hides a later write",
			source: replaceFixtureFunc(pipelineFixtureGo, actionUnmarshal, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 param, _ := decodeActionParam(raw.Type, raw.Param)
 *a = Action{Type: raw.Type, Param: param}
 p := a
 p.Param = &RawActionParam{}
 return nil
}`),
			want: "unsupported action decoder receiver alias p",
		},
		{
			name: "receiver rebound before the switch",
			source: replaceFixtureFunc(pipelineFixtureGo, actionUnmarshal, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 a = &Action{Type: raw.Type}
 switch raw.Type { case ActionTypeClick: a.Param = &ClickParam{} }
 return nil
}`),
			want: "unsupported action decoder receiver reassignment",
		},
		{
			name: "receiver method call may mutate Param",
			source: replaceFixtureFunc(pipelineFixtureGo, actionUnmarshal, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 param, _ := decodeActionParam(raw.Type, raw.Param)
 *a = Action{Type: raw.Type, Param: param}
 a.reset()
 return nil
}
func (a *Action) reset() { a.Param = &RawActionParam{} }
`),
			want: "unsupported action decoder receiver method call reset",
		},
		{
			name: "receiver written from a closure",
			source: replaceFixtureFunc(pipelineFixtureGo, actionUnmarshal, `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 param, _ := decodeActionParam(raw.Type, raw.Param)
 *a = Action{Type: raw.Type, Param: param}
 func() { a.Param = &RawActionParam{} }()
 return nil
}`),
			want: "unsupported action decoder receiver write inside a closure",
		},
		{
			name: "case variable consumed before it is assigned",
			source: replaceFixtureFunc(pipelineFixtureGo, recognitionUnmarshal, `func (r *Recognition) UnmarshalJSON(data []byte) error {
 var raw struct { Type RecognitionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 var param RecognitionParam
 r.Param = param
 switch raw.Type { case RecognitionTypeDirectHit: param = &DirectHitParam{} }
 r.Type = raw.Type
 return nil
}`),
			want: "consumed before it is assigned",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.source == pipelineFixtureGo {
				t.Fatalf("fixture mutation did not apply")
			}
			dir, path := writePipelineFixture(t, pipelineFixtureSchema(), tt.source)
			issues, _, err := checkPipelineCoverage(dir, path, nil)
			if tt.want == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				text := issueText(issues)
				for _, want := range tt.issues {
					if !strings.Contains(text, want) {
						t.Fatalf("missing %q in %q", want, text)
					}
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v; want %q", err, tt.want)
			}
		})
	}
}

// TestPipelineDecoderTrueCoverage verifies that functional helper chains and
// direct case assignments still establish coverage.
func TestPipelineDecoderTrueCoverage(t *testing.T) {
	t.Run("direct assignment for every action case", func(t *testing.T) {
		source := replaceFixtureFunc(pipelineFixtureGo, "func (a *Action) UnmarshalJSON(data []byte) error {", `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 switch raw.Type {
 case ActionTypeClick: a.Param = &ClickParam{}
 }
 return nil
}`)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
	t.Run("helper result via Param field assignment", func(t *testing.T) {
		source := replaceFixtureFunc(pipelineFixtureGo, "func (a *Action) UnmarshalJSON(data []byte) error {", `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 a.Type = raw.Type
 a.Param = param
 return nil
}`)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
	t.Run("case assigns interface variable then Param", func(t *testing.T) {
		source := replaceFixtureFunc(pipelineFixtureGo, "func (r *Recognition) UnmarshalJSON(data []byte) error {", `func (r *Recognition) UnmarshalJSON(data []byte) error {
 var raw struct { Type RecognitionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 var param RecognitionParam
 switch raw.Type {
 case RecognitionTypeDirectHit: param = &DirectHitParam{}
 }
 r.Type = raw.Type
 r.Param = param
 return nil
}`)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
	t.Run("declared interface variable receives helper result", func(t *testing.T) {
		source := replaceFixtureFunc(pipelineFixtureGo, "func (a *Action) UnmarshalJSON(data []byte) error {", `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 var param ActionParam
 param = decodeActionParam(raw.Type, raw.Param)
 *a = Action{Type: raw.Type, Param: param}
 return nil
}`)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
	t.Run("switch on receiver type assigned from decoded envelope", func(t *testing.T) {
		source := replaceFixtureFunc(pipelineFixtureGo, "func (r *Recognition) UnmarshalJSON(data []byte) error {", `func (r *Recognition) UnmarshalJSON(data []byte) error {
 var raw struct { Type RecognitionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 r.Type = raw.Type
 switch r.Type {
 case RecognitionTypeDirectHit: r.Param = &DirectHitParam{}
 }
 return nil
}`)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
	t.Run("unrelated switch alongside the helper", func(t *testing.T) {
		source := replaceFixtureFunc(pipelineFixtureGo, "func (a *Action) UnmarshalJSON(data []byte) error {", `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 switch raw.Type { case ActionTypeClick: validateClick() }
 param, err := decodeActionParam(raw.Type, raw.Param)
 if err != nil { return err }
 *a = Action{Type: raw.Type, Param: param}
 return nil
}
func validateClick() {}
`)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
	t.Run("case writes receiver composite", func(t *testing.T) {
		source := replaceFixtureFunc(pipelineFixtureGo, "func (a *Action) UnmarshalJSON(data []byte) error {", `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 switch raw.Type { case ActionTypeClick: *a = Action{Type: raw.Type, Param: &ClickParam{}} }
 return nil
}`)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
	t.Run("case writes through a parenthesized pointer receiver", func(t *testing.T) {
		source := replaceFixtureFunc(pipelineFixtureGo, "func (a *Action) UnmarshalJSON(data []byte) error {", `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 switch raw.Type { case ActionTypeClick: (*a).Param = &ClickParam{} }
 return nil
}`)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
	t.Run("receiver composite through an alias", func(t *testing.T) {
		source := strings.Replace(pipelineFixtureGo, "type RawActionParam struct", "type ActionAlias = Action\ntype RawActionParam struct", 1)
		source = replaceFixtureFunc(source, "func (a *Action) UnmarshalJSON(data []byte) error {", `func (a *Action) UnmarshalJSON(data []byte) error {
 var raw struct { Type ActionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 param, _ := decodeActionParam(raw.Type, raw.Param)
 *a = ActionAlias{Type: raw.Type, Param: param}
 return nil
}`)
		dir, path := writePipelineFixture(t, pipelineFixtureSchema(), source)
		issues, _, err := checkPipelineCoverage(dir, path, nil)
		if err != nil || len(issues) != 0 {
			t.Fatalf("got %v / %v", issues, err)
		}
	})
}
