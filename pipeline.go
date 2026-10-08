// Package maa provides Go bindings for the MaaFramework.
// Typed pipeline builders use pipeline v2 JSON with nested action and recognition
// objects. Decoding also accepts the legacy flat pipeline format and normalizes
// it into the v2 model; encoding always emits v2.
// For pipeline protocol details, see:
// https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/3.1-PipelineProtocol.md
package maa

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Pipeline represents a collection of nodes that define a task flow. A task
// is a logical sequential structure of nodes connected in a specific order,
// representing the entire process from start to finish; its first node is
// the entry.
type Pipeline struct {
	nodes map[string]*Node
}

// NewPipeline creates a new empty Pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{
		nodes: make(map[string]*Node),
	}
}

// MarshalJSON implements the json.Marshaler interface.
func (p *Pipeline) MarshalJSON() ([]byte, error) {
	return marshalJSON(p.nodes)
}

// UnmarshalJSON replaces the pipeline with a JSON object keyed by node name.
// Each node decodes via [Node.UnmarshalJSON] and takes its Name from the map key.
// An anchor string or list of strings sets each named anchor to the containing
// node, matching the native parser. Object-form targets pass through without
// checking whether they exist. Decoding a Node directly accepts only the object
// form. An omitted anchor remains nil; an empty list or object becomes a non-nil
// empty map and re-encodes as an empty object, clearing the node's anchor
// configuration when applied to MaaFramework, as documented by [Node].
// JSON null decodes to an empty pipeline. On error the pipeline is unchanged.
func (p *Pipeline) UnmarshalJSON(data []byte) error {
	var nodes map[string]json.RawMessage
	if err := unmarshalJSON(data, &nodes); err != nil {
		return err
	}
	decoded := make(map[string]*Node, len(nodes))
	for name, raw := range nodes {
		normalized, err := normalizeNodeAnchor(raw, name)
		if err != nil {
			return err
		}
		var node Node
		if err := unmarshalJSON(normalized, &node); err != nil {
			return err
		}
		node.Name = name
		decoded[name] = &node
	}
	p.nodes = decoded
	return nil
}

// normalizeNodeAnchor rewrites a node's anchor shorthand — a single anchor
// name or a list of anchor names — into the object form resolved to the
// node's own name, matching the native parse_anchor. The object form and
// every other field pass through untouched; null and non-string array
// entries are errors, matching the native parser.
func normalizeNodeAnchor(data []byte, nodeName string) ([]byte, error) {
	var fields map[string]json.RawMessage
	if err := unmarshalJSON(data, &fields); err != nil {
		return nil, err
	}
	value, ok := fields["anchor"]
	if !ok {
		return data, nil
	}
	trimmed := bytes.TrimSpace(value)
	resolved := map[string]string{}
	switch {
	case trimmed[0] == '"':
		var anchor string
		if err := unmarshalJSON(trimmed, &anchor); err != nil {
			return nil, err
		}
		resolved[anchor] = nodeName
	case trimmed[0] == '[':
		var anchors StringList
		if err := unmarshalJSON(trimmed, &anchors); err != nil {
			return nil, err
		}
		for _, anchor := range anchors {
			resolved[anchor] = nodeName
		}
	case trimmed[0] == '{':
		return data, nil
	default:
		return nil, fmt.Errorf("anchor must be a string, array, or object")
	}
	encoded, err := marshalJSON(resolved)
	if err != nil {
		return nil, err
	}
	fields["anchor"] = encoded
	return marshalJSON(fields)
}

// AddNode adds a node to the pipeline and returns the pipeline for chaining.
func (p *Pipeline) AddNode(node *Node) *Pipeline {
	p.nodes[node.Name] = node
	return p
}

// RemoveNode removes a node by name and returns the pipeline for chaining.
func (p *Pipeline) RemoveNode(name string) *Pipeline {
	delete(p.nodes, name)
	return p
}

// Clear resets the pipeline nodes; preserves chaining behavior.
func (p *Pipeline) Clear() *Pipeline {
	p.nodes = make(map[string]*Node)
	return p
}

// GetNode returns a node by name with an existence flag.
func (p *Pipeline) GetNode(name string) (*Node, bool) {
	node, ok := p.nodes[name]
	return node, ok
}

// HasNode reports whether a node with the given name exists.
func (p *Pipeline) HasNode(name string) bool {
	_, ok := p.nodes[name]
	return ok
}

// Len returns the number of nodes in the pipeline.
func (p *Pipeline) Len() int {
	return len(p.nodes)
}
