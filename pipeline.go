// Package maa provides Go bindings for the MaaFramework.
// For pipeline protocol details, see:
// https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/3.1-PipelineProtocol.md

package maa

import (
	"encoding/json"
)

// Node is a single unit of work in a pipeline.
//
// Task is a logical sequential structure consisting of several Nodes connected in a specific order,
// representing the entire process from start to finish.
//
// Entry is the first node in a task.
//
// Pipeline is a collection of all nodes.

// Pipeline represents a collection of nodes that define a task flow.
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
	return json.Marshal(p.nodes)
}

// AddNode adds a node to the pipeline and returns the pipeline for chaining.
func (p *Pipeline) AddNode(node *Node) *Pipeline {
	p.nodes[node.Name] = node
	return p
}
