package checker

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Pipeline coverage intentionally extracts field names, not JSON validation rules.
// Only these upstream metadata and extension hooks are outside the typed v2 model.
var pipelineIgnoredDefs = map[string]bool{
	"jsonComments": true, "jsonCode": true, "jsonDocument": true, "jsonKeywords": true,
	"CustomActionSchema": true, "CustomRecognitionSchema": true,
}

type schemaObject map[string]any

type pipelineSchema struct {
	defs      schemaObject
	node      map[string]bool
	envelopes map[string]map[string]bool
	types     map[string]map[string]map[string]bool
	nested    map[string]map[string]bool
}

func readPipelineSchema(path string) (*pipelineSchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read schema %s: %w", path, err)
	}
	var doc schemaObject
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse schema %s: %w", path, err)
	}
	defs, ok := object(doc["$defs"])
	if !ok {
		return nil, fmt.Errorf("schema %s: missing object $defs", path)
	}
	s := &pipelineSchema{defs: defs, envelopes: map[string]map[string]bool{}, types: map[string]map[string]map[string]bool{}, nested: map[string]map[string]bool{}}
	node, err := s.definition("Node")
	if err != nil {
		return nil, err
	}
	if err := s.checkNodeShape(node); err != nil {
		return nil, err
	}
	// Node's allOf includes both v1 and v2 formats. Its intrinsic properties are
	// shared; never merge the v1 flat parameter fields into this inventory.
	props, ok := object(node["properties"])
	if !ok {
		return nil, fmt.Errorf("schema Node: missing object properties")
	}
	s.node = map[string]bool{}
	for name, value := range props {
		p, ok := object(value)
		if !ok {
			return nil, fmt.Errorf("schema Node.%s: property is not an object", name)
		}
		if p["deprecated"] != true {
			s.node[name] = true
		}
	}
	for _, kind := range []string{"action", "recognition"} {
		if err := s.extractTypes(kind); err != nil {
			return nil, err
		}
	}
	for _, name := range []string{"SwipeListItem", "WaitFreezes", "NodeAttr"} {
		def, err := s.definition(name)
		if err != nil {
			return nil, err
		}
		fields, err := s.fields(def, name, map[string]bool{name: true})
		if err != nil {
			return nil, err
		}
		s.nested[name] = fields
	}
	// The inline sub-recognition inventory is optional: older schemas predate it
	// and only its absence is tolerated.
	if _, exists := s.defs["SubRecognitionInline"]; exists {
		def, err := s.definition("SubRecognitionInline")
		if err != nil {
			return nil, err
		}
		fields, err := s.fields(def, "SubRecognitionInline", map[string]bool{"SubRecognitionInline": true})
		if err != nil {
			return nil, err
		}
		s.nested["SubRecognitionInline"] = fields
	}
	return s, nil
}

var pipelineNodeAllOfRefs = map[string]bool{
	"#/$defs/jsonComments":      true,
	"#/$defs/RecognitionFormat": true,
	"#/$defs/ActionFormat":      true,
}

func (s *pipelineSchema) checkNodeShape(node schemaObject) error {
	if value, ok := node["type"]; ok && value != "object" {
		return fmt.Errorf("schema Node: expected object type")
	}
	if value, exists := node["allOf"]; exists {
		parts, ok := value.([]any)
		if !ok || len(parts) == 0 {
			return fmt.Errorf("schema Node: allOf is not a nonempty array")
		}
		seen := map[string]bool{}
		for _, part := range parts {
			p, ok := object(part)
			if !ok || len(p) != 1 {
				return fmt.Errorf("schema Node: unsupported allOf branch")
			}
			ref, ok := p["$ref"].(string)
			if !ok || !pipelineNodeAllOfRefs[ref] {
				return fmt.Errorf("schema Node: unsupported allOf reference %v", p["$ref"])
			}
			if seen[ref] {
				return fmt.Errorf("schema Node: duplicate allOf reference %s", ref)
			}
			seen[ref] = true
			if _, err := s.definition(strings.TrimPrefix(ref, "#/$defs/")); err != nil {
				return fmt.Errorf("schema Node: unresolvable allOf reference %s: %w", ref, err)
			}
		}
	}
	for _, key := range []string{"$ref", "$dynamicRef", "$recursiveRef", "anyOf", "oneOf", "if", "then", "else", "not", "dependentSchemas", "patternProperties"} {
		if _, exists := node[key]; exists {
			return fmt.Errorf("schema Node: unsupported field extraction keyword %s", key)
		}
	}
	return nil
}

func object(value any) (schemaObject, bool) {
	m, ok := value.(map[string]any)
	return schemaObject(m), ok
}

func (s *pipelineSchema) definition(name string) (schemaObject, error) {
	m, ok := object(s.defs[name])
	if !ok {
		return nil, fmt.Errorf("schema: missing object definition %s", name)
	}
	return m, nil
}

func (s *pipelineSchema) ref(value any, path string) (string, schemaObject, error) {
	ref, ok := value.(string)
	if !ok || !strings.HasPrefix(ref, "#/$defs/") || strings.Contains(strings.TrimPrefix(ref, "#/$defs/"), "/") {
		return "", nil, fmt.Errorf("schema %s: unsupported reference %v (expected local definition)", path, value)
	}
	name := strings.TrimPrefix(ref, "#/$defs/")
	m, err := s.definition(name)
	return name, m, err
}

func (s *pipelineSchema) fields(m schemaObject, path string, stack map[string]bool) (map[string]bool, error) {
	out := map[string]bool{}
	shape := false
	if value, ok := m["type"]; ok && value != "object" {
		return nil, fmt.Errorf("schema %s: expected object type", path)
	}
	if ref, ok := m["$ref"]; ok {
		name, target, err := s.ref(ref, path)
		if err != nil {
			return nil, err
		}
		// Ignore only the referenced definition; sibling keywords still apply.
		if !pipelineIgnoredDefs[name] {
			if stack[name] {
				return nil, fmt.Errorf("schema %s: reference cycle at %s", path, name)
			}
			stack[name] = true
			fields, err := s.fields(target, path+" -> "+name, stack)
			delete(stack, name)
			if err != nil {
				return nil, err
			}
			for f := range fields {
				out[f] = true
			}
		}
		shape = true
	}
	if value, exists := m["properties"]; exists {
		props, ok := object(value)
		if !ok {
			return nil, fmt.Errorf("schema %s: properties is not an object", path)
		}
		for name, value := range props {
			if _, ok := object(value); !ok {
				return nil, fmt.Errorf("schema %s.%s: property is not an object", path, name)
			}
			out[name] = true
		}
		shape = true
	}
	if value, exists := m["allOf"]; exists {
		parts, ok := value.([]any)
		if !ok || len(parts) == 0 {
			return nil, fmt.Errorf("schema %s: allOf is not a nonempty array", path)
		}
		for i, part := range parts {
			p, ok := object(part)
			if !ok {
				return nil, fmt.Errorf("schema %s.allOf[%d]: not an object", path, i)
			}
			fields, err := s.fields(p, fmt.Sprintf("%s.allOf[%d]", path, i), stack)
			if err != nil {
				return nil, err
			}
			for f := range fields {
				out[f] = true
			}
		}
		shape = true
	}
	// required-only alternatives choose whether a Custom callback name or its
	// metadata code override is supplied. They add no intrinsic field names.
	for _, key := range []string{"anyOf", "oneOf"} {
		if value, exists := m[key]; exists {
			parts, ok := value.([]any)
			if !ok || len(parts) == 0 {
				return nil, fmt.Errorf("schema %s: unsupported %s", path, key)
			}
			for _, part := range parts {
				p, ok := object(part)
				if !ok || len(p) != 1 || p["required"] == nil {
					return nil, fmt.Errorf("schema %s: unsupported %s field alternatives", path, key)
				}
				if _, ok := p["required"].([]any); !ok {
					return nil, fmt.Errorf("schema %s: invalid required alternative", path)
				}
			}
		}
	}
	for _, key := range []string{"$dynamicRef", "$recursiveRef", "if", "then", "else", "not", "dependentSchemas", "patternProperties"} {
		if _, exists := m[key]; exists {
			return nil, fmt.Errorf("schema %s: unsupported field extraction keyword %s", path, key)
		}
	}
	if !shape {
		return nil, fmt.Errorf("schema %s: unrecognized object field shape", path)
	}
	return out, nil
}

var pipelineEnvelopeKeywords = map[string]bool{
	"type": true, "properties": true, "allOf": true, "anyOf": true,
	"unevaluatedProperties": true, "title": true, "description": true,
	"markdownDescription": true, "$comment": true,
}

var pipelineBranchKeywords = map[string]bool{
	"properties": true, "dependentSchemas": true, "title": true,
	"description": true, "markdownDescription": true, "$comment": true,
}

// checkEnvelopeShape rejects envelope alternatives that the field extraction
// cannot follow. The only accepted allOf entries are the ignored metadata
// extensions, which add no typed fields.
func (s *pipelineSchema) checkEnvelopeShape(prefix, kind string, wrapper schemaObject) error {
	for _, key := range sortedPipelineKeys(wrapper) {
		if !pipelineEnvelopeKeywords[key] {
			return fmt.Errorf("schema %sV2.%s: unsupported envelope keyword %s", prefix, kind, key)
		}
	}
	value, exists := wrapper["allOf"]
	if !exists {
		return nil
	}
	parts, ok := value.([]any)
	if !ok || len(parts) == 0 {
		return fmt.Errorf("schema %sV2.%s: allOf is not a nonempty array", prefix, kind)
	}
	for _, part := range parts {
		p, ok := object(part)
		if !ok || len(p) != 1 {
			return fmt.Errorf("schema %sV2.%s: unsupported envelope allOf branch", prefix, kind)
		}
		ref, ok := p["$ref"].(string)
		name := strings.TrimPrefix(ref, "#/$defs/")
		if !ok || !strings.HasPrefix(ref, "#/$defs/") || strings.Contains(name, "/") || !pipelineIgnoredDefs[name] {
			return fmt.Errorf("schema %sV2.%s: unsupported envelope allOf reference %v", prefix, kind, p["$ref"])
		}
		if _, err := s.definition(name); err != nil {
			return fmt.Errorf("schema %sV2.%s: unresolvable envelope allOf reference %s: %w", prefix, kind, ref, err)
		}
	}
	return nil
}

// checkEnvelopeBranch keeps a v2 type branch to the type/param envelope so an
// extra branch property cannot hide an unmapped field.
func (s *pipelineSchema) checkEnvelopeBranch(name string, def, props schemaObject) error {
	for _, key := range sortedPipelineKeys(def) {
		if !pipelineBranchKeywords[key] {
			return fmt.Errorf("schema %s: unsupported branch keyword %s", name, key)
		}
	}
	for _, key := range sortedPipelineKeys(props) {
		if key != "type" && key != "param" {
			return fmt.Errorf("schema %s: unsupported branch property %s", name, key)
		}
	}
	if value, exists := def["dependentSchemas"]; exists {
		if err := s.checkBranchDependents(name, value); err != nil {
			return err
		}
	}
	return nil
}

// checkBranchDependents accepts only required-only dependent schemas. They
// select existing envelope fields and cannot add new ones.
func (s *pipelineSchema) checkBranchDependents(name string, value any) error {
	obj, ok := object(value)
	if !ok || len(obj) == 0 {
		return fmt.Errorf("schema %s: dependentSchemas is not a nonempty object", name)
	}
	for _, key := range sortedPipelineKeys(obj) {
		sub, ok := object(obj[key])
		if !ok || len(sub) != 1 {
			return fmt.Errorf("schema %s.dependentSchemas.%s: unsupported shape", name, key)
		}
		ref, ok := sub["$ref"].(string)
		if !ok {
			return fmt.Errorf("schema %s.dependentSchemas.%s: unsupported reference %v", name, key, sub["$ref"])
		}
		target, targetDef, err := s.ref(ref, name+".dependentSchemas."+key)
		if err != nil {
			return err
		}
		if err := checkRequiredOnlyDefinition(target, targetDef); err != nil {
			return err
		}
	}
	return nil
}

func checkRequiredOnlyDefinition(name string, def schemaObject) error {
	for _, key := range sortedPipelineKeys(def) {
		switch key {
		case "anyOf", "oneOf", "title", "description", "markdownDescription", "$comment":
		default:
			return fmt.Errorf("schema %s: dependent schema must only require existing fields (keyword %s)", name, key)
		}
	}
	for _, key := range []string{"anyOf", "oneOf"} {
		value, exists := def[key]
		if !exists {
			continue
		}
		parts, ok := value.([]any)
		if !ok || len(parts) == 0 {
			return fmt.Errorf("schema %s: invalid %s", name, key)
		}
		for _, part := range parts {
			p, ok := object(part)
			if !ok || len(p) != 1 || p["required"] == nil {
				return fmt.Errorf("schema %s: unsupported %s alternative", name, key)
			}
			if _, ok := p["required"].([]any); !ok {
				return fmt.Errorf("schema %s: invalid required alternative", name)
			}
		}
	}
	return nil
}

func (s *pipelineSchema) extractTypes(kind string) error {
	prefix := "Action"
	if kind == "recognition" {
		prefix = "Recognition"
	}
	enum, err := s.definition(prefix + "Enum")
	if err != nil {
		return err
	}
	values, ok := enum["enum"].([]any)
	if !ok || len(values) == 0 {
		return fmt.Errorf("schema %sEnum: missing nonempty enum", prefix)
	}
	types := map[string]map[string]bool{}
	for _, value := range values {
		name, ok := value.(string)
		if !ok || name == "" {
			return fmt.Errorf("schema %sEnum: invalid type %v", prefix, value)
		}
		if _, duplicate := types[name]; duplicate {
			return fmt.Errorf("schema %sEnum: duplicate type %s", prefix, name)
		}
		types[name] = nil
	}
	v2, err := s.definition(prefix + "V2")
	if err != nil {
		return err
	}
	props, ok := object(v2["properties"])
	if !ok {
		return fmt.Errorf("schema %sV2: missing properties", prefix)
	}
	wrapper, ok := object(props[kind])
	if !ok {
		return fmt.Errorf("schema %sV2: missing %s object", prefix, kind)
	}
	if err := s.checkEnvelopeShape(prefix, kind, wrapper); err != nil {
		return err
	}
	envelopeProps, ok := object(wrapper["properties"])
	if !ok {
		return fmt.Errorf("schema %sV2.%s: missing envelope properties", prefix, kind)
	}
	envelope := map[string]bool{}
	for name, value := range envelopeProps {
		if _, ok := object(value); !ok {
			return fmt.Errorf("schema %sV2.%s.%s: envelope property is not an object", prefix, kind, name)
		}
		if name != "type" && name != "param" {
			return fmt.Errorf("schema %sV2.%s: unsupported envelope property %s", prefix, kind, name)
		}
		envelope[name] = true
	}
	for _, name := range []string{"type", "param"} {
		if !envelope[name] {
			return fmt.Errorf("schema %sV2.%s: missing %s envelope property", prefix, kind, name)
		}
	}
	s.envelopes[kind] = envelope
	branches, ok := wrapper["anyOf"].([]any)
	if !ok || len(branches) == 0 {
		return fmt.Errorf("schema %sV2.%s: missing anyOf branches", prefix, kind)
	}
	for _, branch := range branches {
		b, ok := object(branch)
		if !ok || len(b) != 1 {
			return fmt.Errorf("schema %sV2.%s: unsupported branch", prefix, kind)
		}
		name, def, err := s.ref(b["$ref"], prefix+"V2."+kind)
		if err != nil {
			return err
		}
		if name == prefix+"PropertyFieldsV2" {
			continue
		}
		p, ok := object(def["properties"])
		if !ok {
			return fmt.Errorf("schema %s: missing branch properties", name)
		}
		t, ok := object(p["type"])
		if !ok {
			return fmt.Errorf("schema %s: missing type const", name)
		}
		typeName, ok := t["const"].(string)
		if !ok || typeName == "" {
			return fmt.Errorf("schema %s: missing type const", name)
		}
		if _, exists := types[typeName]; !exists {
			return fmt.Errorf("schema %s: type %s absent from %sEnum", name, typeName, prefix)
		}
		if types[typeName] != nil {
			return fmt.Errorf("schema %s: duplicate branch for %s", name, typeName)
		}
		if err := s.checkEnvelopeBranch(name, def, p); err != nil {
			return err
		}
		param, ok := object(p["param"])
		if !ok {
			return fmt.Errorf("schema %s: missing param object", name)
		}
		fields, err := s.fields(param, kind+"."+typeName+".param", map[string]bool{name: true})
		if err != nil {
			return err
		}
		types[typeName] = fields
	}
	// Keep enum-only additions in the inventory. They still require a Go enum
	// and typed decoder even when the schema omitted a branch accidentally.
	s.types[kind] = types
	return nil
}

func sortedPipelineKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
