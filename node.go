package maa

import (
	"slices"
	"time"
)

// Node represents a single task node in the pipeline.
type Node struct {
	Name string `json:"-"`

	// Anchor specifies the anchor name that can be referenced in next or on_error lists via [Anchor] attribute.
	Anchor []string `json:"anchor,omitempty"`

	// Recognition defines how this node recognizes targets on screen.
	Recognition *NodeRecognition `json:"recognition,omitempty"`
	// Action defines what action to perform when recognition succeeds.
	Action *NodeAction `json:"action,omitempty"`
	// Next specifies the list of possible next nodes to execute.
	Next []NodeNextItem `json:"next,omitempty"`
	// RateLimit sets the minimum interval between recognition attempts in milliseconds. Default: 1000.
	RateLimit *int64 `json:"rate_limit,omitempty"`
	// Timeout sets the maximum time to wait for recognition in milliseconds. Default: 20000.
	Timeout *int64 `json:"timeout,omitempty"`
	// OnError specifies nodes to execute when recognition times out or action execution fails.
	OnError []NodeNextItem `json:"on_error,omitempty"`
	// Inverse inverts the recognition result. Default: false.
	Inverse bool `json:"inverse,omitempty"`
	// Enabled determines whether this node is active. Default: true.
	Enabled *bool `json:"enabled,omitempty"`
	// MaxHit sets the maximum hit count of the node. Default: unlimited.
	MaxHit *uint64 `json:"max_hit,omitempty"`
	// PreDelay sets the delay before action execution in milliseconds. Default: 200.
	PreDelay *int64 `json:"pre_delay,omitempty"`
	// PostDelay sets the delay after action execution in milliseconds. Default: 200.
	PostDelay *int64 `json:"post_delay,omitempty"`
	// PreWaitFreezes waits for screen to stabilize before action execution.
	PreWaitFreezes *NodeWaitFreezes `json:"pre_wait_freezes,omitempty"`
	// PostWaitFreezes waits for screen to stabilize after action.
	PostWaitFreezes *NodeWaitFreezes `json:"post_wait_freezes,omitempty"`
	// Repeat specifies the number of times to repeat the node. Default: 1.
	Repeat *uint64 `json:"repeat,omitempty"`
	// RepeatDelay sets the delay between repetitions in milliseconds. Default: 0.
	RepeatDelay *int64 `json:"repeat_delay,omitempty"`
	// RepeatWaitFreezes waits for screen to stabilize between repetitions.
	RepeatWaitFreezes *NodeWaitFreezes `json:"repeat_wait_freezes,omitempty"`
	// Focus specifies custom focus data.
	Focus any `json:"focus,omitempty"`
	// Attach provides additional custom data for the node.
	Attach map[string]any `json:"attach,omitempty"`
}

// NodeOption is a functional option for configuring a Node.
type NodeOption func(*Node)

// WithRecognition sets the recognition for the node.
func WithRecognition(rec *NodeRecognition) NodeOption {
	return func(n *Node) {
		n.Recognition = rec
	}
}

// WithAction sets the action for the node.
func WithAction(act *NodeAction) NodeOption {
	return func(n *Node) {
		n.Action = act
	}
}

// WithNext sets the next nodes list for the node.
func WithNext(next []NodeNextItem) NodeOption {
	return func(n *Node) {
		n.Next = slices.Clone(next)
	}
}

// WithRateLimit sets the rate limit for the node.
func WithRateLimit(rateLimit time.Duration) NodeOption {
	return func(n *Node) {
		d := rateLimit.Milliseconds()
		n.RateLimit = &d
	}
}

// WithTimeout sets the timeout for the node.
func WithTimeout(timeout time.Duration) NodeOption {
	return func(n *Node) {
		d := timeout.Milliseconds()
		n.Timeout = &d
	}
}

// WithOnError sets the error handling nodes for the node.
func WithOnError(onError []NodeNextItem) NodeOption {
	return func(n *Node) {
		n.OnError = slices.Clone(onError)
	}
}

// WithInverse sets whether to invert the recognition result.
func WithInverse(inverse bool) NodeOption {
	return func(n *Node) {
		n.Inverse = inverse
	}
}

// WithEnabled sets whether the node is enabled.
func WithEnabled(enabled bool) NodeOption {
	return func(n *Node) {
		n.Enabled = &enabled
	}
}

// WithMaxHit sets the maximum hit count of the node.
func WithMaxHit(maxHit uint64) NodeOption {
	return func(n *Node) {
		n.MaxHit = &maxHit
	}
}

// WithPreDelay sets the delay before action execution.
func WithPreDelay(preDelay time.Duration) NodeOption {
	return func(n *Node) {
		d := preDelay.Milliseconds()
		n.PreDelay = &d
	}
}

// WithPostDelay sets the delay after action execution.
func WithPostDelay(postDelay time.Duration) NodeOption {
	return func(n *Node) {
		d := postDelay.Milliseconds()
		n.PostDelay = &d
	}
}

// WithPreWaitFreezes sets the pre-action wait freezes configuration.
func WithPreWaitFreezes(waitFreezes *NodeWaitFreezes) NodeOption {
	return func(n *Node) {
		n.PreWaitFreezes = waitFreezes
	}
}

// WithPostWaitFreezes sets the post-action wait freezes configuration.
func WithPostWaitFreezes(waitFreezes *NodeWaitFreezes) NodeOption {
	return func(n *Node) {
		n.PostWaitFreezes = waitFreezes
	}
}

// WithRepeat sets the number of times to repeat the node.
func WithRepeat(repeat uint64) NodeOption {
	return func(n *Node) {
		n.Repeat = &repeat
	}
}

// WithRepeatDelay sets the delay between repetitions.
func WithRepeatDelay(repeatDelay time.Duration) NodeOption {
	return func(n *Node) {
		d := repeatDelay.Milliseconds()
		n.RepeatDelay = &d
	}
}

// WithRepeatWaitFreezes sets the wait freezes configuration between repetitions.
func WithRepeatWaitFreezes(waitFreezes *NodeWaitFreezes) NodeOption {
	return func(n *Node) {
		n.RepeatWaitFreezes = waitFreezes
	}
}

// WithFocus sets the focus data for the node.
func WithFocus(focus any) NodeOption {
	return func(n *Node) {
		n.Focus = focus
	}
}

// WithAttach sets the attached custom data for the node.
func WithAttach(attach map[string]any) NodeOption {
	return func(n *Node) {
		n.Attach = attach
	}
}

// NewNode creates a new Node with the given name and options.
func NewNode(name string, opts ...NodeOption) *Node {
	n := &Node{
		Name: name,
	}

	for _, opt := range opts {
		opt(n)
	}

	return n
}

// SetAnchor sets the anchor for the node and returns the node for chaining.
func (n *Node) SetAnchor(anchor []string) *Node {
	n.Anchor = slices.Clone(anchor)
	return n
}

// SetRecognition sets the recognition for the node and returns the node for chaining.
func (n *Node) SetRecognition(rec *NodeRecognition) *Node {
	n.Recognition = rec
	return n
}

// SetAction sets the action for the node and returns the node for chaining.
func (n *Node) SetAction(act *NodeAction) *Node {
	n.Action = act
	return n
}

// SetNext sets the next nodes list for the node and returns the node for chaining.
func (n *Node) SetNext(next []NodeNextItem) *Node {
	n.Next = slices.Clone(next)
	return n
}

// SetRateLimit sets the rate limit for the node and returns the node for chaining.
func (n *Node) SetRateLimit(rateLimit time.Duration) *Node {
	d := rateLimit.Milliseconds()
	n.RateLimit = &d
	return n
}

// SetTimeout sets the timeout for the node and returns the node for chaining.
func (n *Node) SetTimeout(timeout time.Duration) *Node {
	d := timeout.Milliseconds()
	n.Timeout = &d
	return n
}

// SetOnError sets the error handling nodes for the node and returns the node for chaining.
func (n *Node) SetOnError(onError []NodeNextItem) *Node {
	n.OnError = slices.Clone(onError)
	return n
}

// SetInverse sets whether to invert the recognition result and returns the node for chaining.
func (n *Node) SetInverse(inverse bool) *Node {
	n.Inverse = inverse
	return n
}

// SetEnabled sets whether the node is enabled and returns the node for chaining.
func (n *Node) SetEnabled(enabled bool) *Node {
	n.Enabled = &enabled
	return n
}

// SetMaxHit sets the maximum hit count of the node and returns the node for chaining.
func (n *Node) SetMaxHit(maxHit uint64) *Node {
	n.MaxHit = &maxHit
	return n
}

// SetPreDelay sets the delay before action execution and returns the node for chaining.
func (n *Node) SetPreDelay(preDelay time.Duration) *Node {
	d := preDelay.Milliseconds()
	n.PreDelay = &d
	return n
}

// SetPostDelay sets the delay after action execution and returns the node for chaining.
func (n *Node) SetPostDelay(postDelay time.Duration) *Node {
	d := postDelay.Milliseconds()
	n.PostDelay = &d
	return n
}

// SetPreWaitFreezes sets the pre-action wait freezes configuration and returns the node for chaining.
func (n *Node) SetPreWaitFreezes(preWaitFreezes *NodeWaitFreezes) *Node {
	n.PreWaitFreezes = preWaitFreezes
	return n
}

// SetPostWaitFreezes sets the post-action wait freezes configuration and returns the node for chaining.
func (n *Node) SetPostWaitFreezes(postWaitFreezes *NodeWaitFreezes) *Node {
	n.PostWaitFreezes = postWaitFreezes
	return n
}

// SetRepeat sets the number of times to repeat the node and returns the node for chaining.
func (n *Node) SetRepeat(repeat uint64) *Node {
	n.Repeat = &repeat
	return n
}

// SetRepeatDelay sets the delay between repetitions and returns the node for chaining.
func (n *Node) SetRepeatDelay(repeatDelay time.Duration) *Node {
	d := repeatDelay.Milliseconds()
	n.RepeatDelay = &d
	return n
}

// SetRepeatWaitFreezes sets the wait freezes configuration between repetitions and returns the node for chaining.
func (n *Node) SetRepeatWaitFreezes(repeatWaitFreezes *NodeWaitFreezes) *Node {
	n.RepeatWaitFreezes = repeatWaitFreezes
	return n
}

// SetFocus sets the focus data for the node and returns the node for chaining.
func (n *Node) SetFocus(focus any) *Node {
	n.Focus = focus
	return n
}

// SetAttach sets the attached custom data for the node and returns the node for chaining.
func (n *Node) SetAttach(attach map[string]any) *Node {
	n.Attach = attach
	return n
}

// NodeNextItem represents an item in the next or on_error list.
type NodeNextItem struct {
	// Name is the name of the target node.
	Name string `json:"name"`
	// JumpBack indicates whether to jump back to the parent node after this node's chain completes.
	JumpBack bool `json:"jump_back"`
	// Anchor indicates whether this node should be set as the anchor.
	Anchor bool `json:"anchor"`
}

// FormatName returns the name with attribute prefixes, e.g. [JumpBack]NodeA.
func (i NodeNextItem) FormatName() string {
	name := i.Name
	if i.JumpBack {
		name = "[JumpBack]" + name
	}
	if i.Anchor {
		name = "[Anchor]" + name
	}
	return name
}

// NodeAttributeOption is a functional option for configuring NodeNextItem attributes.
type NodeAttributeOption func(*NodeNextItem)

// WithJumpBack enables the jump-back mechanism. When this node matches, the system returns
// to the parent node after completing this node's chain, and continues recognizing from the start of next list.
func WithJumpBack() NodeAttributeOption {
	return func(i *NodeNextItem) {
		i.JumpBack = true
	}
}

// WithAnchor enables anchor reference. The name field will be treated as an anchor name
// and resolved to the last node that set this anchor at runtime.
func WithAnchor() NodeAttributeOption {
	return func(i *NodeNextItem) {
		i.Anchor = true
	}
}

// AddAnchor appends an anchor to the node and returns the node for chaining.
func (n *Node) AddAnchor(anchor string) *Node {
	if anchor == "" {
		return n
	}

	for _, a := range n.Anchor {
		if a == anchor {
			return n
		}
	}
	n.Anchor = append(n.Anchor, anchor)
	return n
}

// RemoveAnchor removes an anchor from the node and returns the node for chaining.
func (n *Node) RemoveAnchor(anchor string) *Node {
	if anchor == "" {
		return n
	}

	n.Anchor = slices.DeleteFunc(n.Anchor, func(a string) bool {
		return a == anchor
	})

	return n
}

// AddNext appends a node to the next list and returns the node for chaining.
func (n *Node) AddNext(name string, opts ...NodeAttributeOption) *Node {
	if name == "" {
		return n
	}

	newItem := NodeNextItem{
		Name: name,
	}
	for _, opt := range opts {
		opt(&newItem)
	}

	found := false
	for i, item := range n.Next {
		if item.Name == name {
			n.Next[i] = newItem
			found = true
			break
		}
	}

	if !found {
		n.Next = append(n.Next, newItem)
	}
	return n
}

// RemoveNext removes a node from the next list and returns the node for chaining.
func (n *Node) RemoveNext(name string) *Node {
	if name == "" {
		return n
	}

	n.Next = slices.DeleteFunc(n.Next, func(item NodeNextItem) bool {
		return item.Name == name
	})

	return n
}

// AddOnError appends a node to the on_error list and returns the node for chaining.
func (n *Node) AddOnError(name string, opts ...NodeAttributeOption) *Node {
	if name == "" {
		return n
	}

	newItem := NodeNextItem{
		Name: name,
	}
	for _, opt := range opts {
		opt(&newItem)
	}

	found := false
	for i, item := range n.OnError {
		if item.Name == name {
			n.OnError[i] = newItem
			found = true
			break
		}
	}

	if !found {
		n.OnError = append(n.OnError, newItem)
	}
	return n
}

// RemoveOnError removes a node from the on_error list and returns the node for chaining.
func (n *Node) RemoveOnError(name string) *Node {
	if name == "" {
		return n
	}

	n.OnError = slices.DeleteFunc(n.OnError, func(item NodeNextItem) bool {
		return item.Name == name
	})

	return n
}

// NodeWaitFreezes defines parameters for waiting until screen stabilizes.
// The screen is considered stable when there are no significant changes for a continuous period.
type NodeWaitFreezes struct {
	// Time specifies the duration in milliseconds that the screen must remain stable. Default: 1.
	Time int64 `json:"time,omitempty"`
	// Target specifies the region to monitor for changes.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	TargetOffset Rect `json:"target_offset,omitempty"`
	// Threshold specifies the template matching threshold for detecting changes. Default: 0.95.
	Threshold float64 `json:"threshold,omitempty"`
	// Method specifies the template matching algorithm (cv::TemplateMatchModes). Default: 5.
	Method int `json:"method,omitempty"`
	// RateLimit specifies the minimum interval between checks in milliseconds. Default: 1000.
	RateLimit int64 `json:"rate_limit,omitempty"`
	// Timeout specifies the maximum wait time in milliseconds. Default: 20000.
	Timeout int64 `json:"timeout,omitempty"`
}

// WaitFreezesOption is a functional option for configuring NodeWaitFreezes.
type WaitFreezesOption func(*NodeWaitFreezes)

// WithWaitFreezesTime sets the duration that the screen must remain stable.
func WithWaitFreezesTime(d time.Duration) WaitFreezesOption {
	return func(w *NodeWaitFreezes) {
		w.Time = d.Milliseconds()
	}
}

// WithWaitFreezesTarget sets the region to monitor for changes.
func WithWaitFreezesTarget(target Target) WaitFreezesOption {
	return func(w *NodeWaitFreezes) {
		w.Target = target
	}
}

// WithWaitFreezesTargetOffset sets additional offset applied to target.
func WithWaitFreezesTargetOffset(offset Rect) WaitFreezesOption {
	return func(w *NodeWaitFreezes) {
		w.TargetOffset = offset
	}
}

// WithWaitFreezesThreshold sets the template matching threshold for detecting changes.
func WithWaitFreezesThreshold(th float64) WaitFreezesOption {
	return func(w *NodeWaitFreezes) {
		w.Threshold = th
	}
}

// WithWaitFreezesMethod sets the template matching algorithm.
func WithWaitFreezesMethod(m int) WaitFreezesOption {
	return func(w *NodeWaitFreezes) {
		w.Method = m
	}
}

// WithWaitFreezesRateLimit sets the minimum interval between checks.
func WithWaitFreezesRateLimit(d time.Duration) WaitFreezesOption {
	return func(w *NodeWaitFreezes) {
		w.RateLimit = d.Milliseconds()
	}
}

// WithWaitFreezesTimeout sets the maximum wait time.
func WithWaitFreezesTimeout(d time.Duration) WaitFreezesOption {
	return func(w *NodeWaitFreezes) {
		w.Timeout = d.Milliseconds()
	}
}

// WaitFreezes creates a NodeWaitFreezes configuration with the given options.
func WaitFreezes(opts ...WaitFreezesOption) *NodeWaitFreezes {
	w := &NodeWaitFreezes{}
	for _, opt := range opts {
		opt(w)
	}
	return w
}
