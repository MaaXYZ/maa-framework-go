package checker

import (
	"fmt"
	"strings"
)

func checkPipelineCoverage(repoRoot, schemaPath string, exclusions map[string]string) ([]issue, []string, error) {
	for path, reason := range exclusions {
		if strings.TrimSpace(path) == "" || strings.TrimSpace(reason) == "" {
			return nil, nil, fmt.Errorf("pipeline exclusion %q requires an exact issue path and a nonempty reason", path)
		}
	}
	schema, err := readPipelineSchema(schemaPath)
	if err != nil {
		return nil, nil, err
	}
	goModel, err := readPipelineGo(repoRoot)
	if err != nil {
		return nil, nil, err
	}
	problems := map[string][]string{}
	add := func(path, message string) { problems[path] = append(problems[path], path+": "+message) }
	compare := func(path, typeName string, expected map[string]bool, extract func(string) (pipelineFieldSet, error)) error {
		actual, err := extract(typeName)
		if err != nil {
			return err
		}
		for _, field := range sortedPipelineKeys(expected) {
			if _, ok := actual[field]; !ok {
				add(path+"."+field, "missing Go field ("+typeName+")")
			}
		}
		for _, field := range sortedPipelineKeys(actual) {
			if !expected[field] {
				add(path+"."+field, "Go field absent from schema ("+typeName+", "+actual[field]+")")
			}
		}
		return nil
	}
	compareFields := func(path, typeName string, expected map[string]bool) error {
		return compare(path, typeName, expected, goModel.fields)
	}
	if err := compare("node", "Node", schema.node, goModel.fields); err != nil {
		return nil, nil, err
	}
	for _, kind := range []string{"action", "recognition"} {
		typeName := "Action"
		if kind == "recognition" {
			typeName = "Recognition"
		}
		// The envelope carries type/param in both directions. A mutated struct
		// tag or a diverging decode DTO is a coverage error, not silent support.
		if err := compare(kind+".envelope", typeName, schema.envelopes[kind], goModel.envelopeFields); err != nil {
			return nil, nil, err
		}
		decoder, err := goModel.decoderParams(kind)
		if err != nil {
			return nil, nil, err
		}
		enumNames := map[string]bool{}
		for _, value := range goModel.enums[kind] {
			enumNames[value] = true
		}
		for _, name := range sortedPipelineKeys(schema.types[kind]) {
			path := kind + "." + name
			if !enumNames[name] {
				add(path+".type", "missing Go type constant")
			}
			param := decoder[name]
			if param == "" {
				add(path+".decoder", "missing typed Go decoder case")
			}
			fields := schema.types[kind][name]
			if fields == nil {
				add(path+".schema", "enum type has no v2 schema branch")
				continue
			}
			if param != "" {
				if err := compareFields(path+".param", param, fields); err != nil {
					return nil, nil, err
				}
			}
		}
		for _, name := range sortedPipelineKeys(enumNames) {
			if _, ok := schema.types[kind][name]; !ok {
				add(kind+"."+name+".type", "Go type constant absent from schema")
			}
		}
		for _, name := range sortedPipelineKeys(decoder) {
			if _, ok := schema.types[kind][name]; !ok {
				add(kind+"."+name+".decoder", "typed Go decoder case absent from schema")
			}
		}
	}
	for _, pair := range [][2]string{{"SwipeListItem", "MultiSwipeItem"}, {"WaitFreezes", "WaitFreezesParam"}, {"NodeAttr", "NextItem"}} {
		if err := compareFields(pair[0], pair[1], schema.nested[pair[0]]); err != nil {
			return nil, nil, err
		}
	}
	// The inline sub-recognition wraps the standard recognition envelope under
	// "recognition"; the schema def only declares sub_name, so that envelope key
	// is validated through the Recognition envelope check instead.
	if expected, ok := schema.nested["SubRecognitionInline"]; ok {
		actual, err := goModel.envelopeFields("InlineSubRecognition")
		if err != nil {
			return nil, nil, err
		}
		for _, field := range sortedPipelineKeys(expected) {
			if _, ok := actual[field]; !ok {
				add("SubRecognitionInline."+field, "missing Go field (InlineSubRecognition)")
			}
		}
		if _, ok := actual["recognition"]; !ok {
			add("SubRecognitionInline.recognition", "missing inline recognition envelope (InlineSubRecognition)")
		}
		for _, field := range sortedPipelineKeys(actual) {
			if expected[field] || field == "recognition" {
				continue
			}
			add("SubRecognitionInline."+field, "Go field absent from schema (InlineSubRecognition, "+actual[field]+")")
		}
	}
	var issues []issue
	var excluded []string
	for _, path := range sortedPipelineKeys(problems) {
		if reason, ok := exclusions[path]; ok {
			excluded = append(excluded, fmt.Sprintf("pipeline_exclusion: %s (%s)", path, strings.TrimSpace(reason)))
			continue
		}
		for _, message := range problems[path] {
			issues = append(issues, issue{section: sectionPipeline, message: message})
		}
	}
	for _, path := range sortedPipelineKeys(exclusions) {
		if _, exists := problems[path]; !exists {
			issues = append(issues, issue{section: sectionPipeline, message: path + ": stale pipeline exclusion (no current difference)"})
		}
	}
	return issues, excluded, nil
}
