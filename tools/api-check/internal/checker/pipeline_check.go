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
	compare := func(path, typeName string, expected map[string]bool) error {
		actual, err := goModel.fields(typeName)
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
	if err := compare("node", "Node", schema.node); err != nil {
		return nil, nil, err
	}
	for _, kind := range []string{"action", "recognition"} {
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
				if err := compare(path+".param", param, fields); err != nil {
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
		if err := compare(pair[0], pair[1], schema.nested[pair[0]]); err != nil {
			return nil, nil, err
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
