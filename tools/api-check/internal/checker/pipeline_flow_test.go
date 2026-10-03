package checker

import (
	"strings"
	"testing"
)

func TestPipelineDiscardedCaseCompositeDoesNotCoverParam(t *testing.T) {
	source := replaceFixtureFunc(pipelineFixtureGo, "func (r *Recognition) UnmarshalJSON(data []byte) error {", `func (r *Recognition) UnmarshalJSON(data []byte) error {
 var raw struct { Type RecognitionType `+"`json:\"type\"`"+`; Param json.RawMessage `+"`json:\"param\"`"+` }
 if err := unmarshalJSON(data, &raw); err != nil { return err }
 switch raw.Type {
 case RecognitionTypeDirectHit:
   _ = Recognition{Type: raw.Type, Param: &DirectHitParam{}}
   r.Param = nil
 }
 return nil
}`)
	dir, schema := writePipelineFixture(t, pipelineFixtureSchema(), source)
	issues, _, err := checkPipelineCoverage(dir, schema, nil)
	if err == nil && !strings.Contains(issueText(issues), "missing typed Go decoder case") {
		t.Fatalf("discarded composite established coverage: issues=%v err=%v", issues, err)
	}
}
