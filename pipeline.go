// Package maa provides Go bindings for the MaaFramework.
// For pipeline protocol details, see:
// https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/3.1-PipelineProtocol.md

package maa

import (
	"encoding/json"
	"errors"
	"time"
)

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
		n.Next = next
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
		n.OnError = onError
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
	n.Anchor = anchor
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
	n.Next = next
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
	n.OnError = onError
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

	for i, a := range n.Anchor {
		if a == anchor {
			copy(n.Anchor[i:], n.Anchor[i+1:])
			n.Anchor = n.Anchor[:len(n.Anchor)-1]
			break
		}
	}
	return n
}

// AddNext appends a node to the next list and returns the node for chaining.
func (n *Node) AddNext(name string, opts ...NodeAttributeOption) *Node {
	if name == "" {
		return n
	}

	for _, item := range n.Next {
		if item.Name == name {
			return n
		}
	}

	next := NodeNextItem{
		Name: name,
	}
	for _, opt := range opts {
		opt(&next)
	}
	n.Next = append(n.Next, next)
	return n
}

// RemoveNext removes a node from the next list and returns the node for chaining.
func (n *Node) RemoveNext(name string) *Node {
	if name == "" {
		return n
	}

	for i, item := range n.Next {
		if item.Name == name {
			copy(n.Next[i:], n.Next[i+1:])
			n.Next = n.Next[:len(n.Next)-1]
			break
		}
	}
	return n
}

// AddOnError appends a node to the on_error list and returns the node for chaining.
func (n *Node) AddOnError(name string, opts ...NodeAttributeOption) *Node {
	if name == "" {
		return n
	}

	for _, item := range n.OnError {
		if item.Name == name {
			return n
		}
	}

	onError := NodeNextItem{
		Name: name,
	}

	for _, opt := range opts {
		opt(&onError)
	}

	n.OnError = append(n.OnError, onError)
	return n
}

// RemoveOnError removes a node from the on_error list and returns the node for chaining.
func (n *Node) RemoveOnError(name string) *Node {
	if name == "" {
		return n
	}

	for i, item := range n.OnError {
		if item.Name == name {
			copy(n.OnError[i:], n.OnError[i+1:])
			n.OnError = n.OnError[:len(n.OnError)-1]
			break
		}
	}
	return n
}

// NodeRecognition defines the recognition configuration for a node.
type NodeRecognition struct {
	// Type specifies the recognition algorithm type.
	Type NodeRecognitionType `json:"type,omitempty"`
	// Param specifies the recognition parameters.
	Param NodeRecognitionParam `json:"param,omitempty"`
}

func (nr *NodeRecognition) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type  NodeRecognitionType `json:"type,omitempty"`
		Param json.RawMessage     `json:"param,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	nr.Type = raw.Type

	// If no param provided or null, just return with type set
	if len(raw.Param) == 0 || string(raw.Param) == "null" {
		return nil
	}

	// Unmarshal param based on type
	var param NodeRecognitionParam
	switch nr.Type {
	case NodeRecognitionTypeDirectHit, "":
		param = &NodeDirectHitParam{}
	case NodeRecognitionTypeTemplateMatch:
		param = &NodeTemplateMatchParam{}
	case NodeRecognitionTypeFeatureMatch:
		param = &NodeFeatureMatchParam{}
	case NodeRecognitionTypeColorMatch:
		param = &NodeColorMatchParam{}
	case NodeRecognitionTypeOCR:
		param = &NodeOCRParam{}
	case NodeRecognitionTypeNeuralNetworkClassify:
		param = &NodeNeuralNetworkClassifyParam{}
	case NodeRecognitionTypeNeuralNetworkDetect:
		param = &NodeNeuralNetworkDetectParam{}
	case NodeRecognitionTypeCustom:
		param = &NodeCustomRecognitionParam{}
	default:
		return errors.New("unsupported recognition type: " + string(nr.Type))
	}

	if err := json.Unmarshal(raw.Param, param); err != nil {
		return err
	}
	nr.Param = param
	return nil
}

// NodeRecognitionType defines the available recognition algorithm types.
type NodeRecognitionType string

const (
	NodeRecognitionTypeDirectHit             NodeRecognitionType = "DirectHit"
	NodeRecognitionTypeTemplateMatch         NodeRecognitionType = "TemplateMatch"
	NodeRecognitionTypeFeatureMatch          NodeRecognitionType = "FeatureMatch"
	NodeRecognitionTypeColorMatch            NodeRecognitionType = "ColorMatch"
	NodeRecognitionTypeOCR                   NodeRecognitionType = "OCR"
	NodeRecognitionTypeNeuralNetworkClassify NodeRecognitionType = "NeuralNetworkClassify"
	NodeRecognitionTypeNeuralNetworkDetect   NodeRecognitionType = "NeuralNetworkDetect"
	NodeRecognitionTypeCustom                NodeRecognitionType = "Custom"
)

// NodeRecognitionParam is the interface for recognition parameters.
type NodeRecognitionParam interface {
	isRecognitionParam()
}

// NodeDirectHitParam defines parameters for direct hit recognition.
// DirectHit performs no actual recognition and always succeeds.
type NodeDirectHitParam struct{}

func (n NodeDirectHitParam) isRecognitionParam() {}

// RecDirectHit creates a DirectHit recognition that always succeeds without actual recognition.
func RecDirectHit() *NodeRecognition {
	return &NodeRecognition{
		Type:  NodeRecognitionTypeDirectHit,
		Param: &NodeDirectHitParam{},
	}
}

// NodeTemplateMatchOrderBy defines the ordering options for template match results.
type NodeTemplateMatchOrderBy string

const (
	NodeTemplateMatchOrderByHorizontal NodeTemplateMatchOrderBy = "Horizontal"
	NodeTemplateMatchOrderByVertical   NodeTemplateMatchOrderBy = "Vertical"
	NodeTemplateMatchOrderByScore      NodeTemplateMatchOrderBy = "Score"
	NodeTemplateMatchOrderByRandom     NodeTemplateMatchOrderBy = "Random"
)

// NodeTemplateMatchMethod defines the template matching algorithm (cv::TemplateMatchModes).
type NodeTemplateMatchMethod int

const (
	NodeTemplateMatchMethodSQDIFF_NORMED_Inverted NodeTemplateMatchMethod = 10001 // Normalized squared difference (Inverted)
	NodeTemplateMatchMethodCCORR_NORMED           NodeTemplateMatchMethod = 3     // Normalized cross correlation
	NodeTemplateMatchMethodCCOEFF_NORMED          NodeTemplateMatchMethod = 5     // Normalized correlation coefficient (default, most accurate)
)

// NodeTemplateMatchParam defines parameters for template matching recognition.
type NodeTemplateMatchParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitempty"`
	// Template specifies the template image paths. Required.
	Template []string `json:"template,omitempty"`
	// Threshold specifies the matching threshold [0-1.0]. Default: 0.7.
	Threshold []float64 `json:"threshold,omitempty"`
	// OrderBy specifies the result ordering. Default: Horizontal.
	OrderBy NodeTemplateMatchOrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
	// Method specifies the matching algorithm. 1: SQDIFF_NORMED, 3: CCORR_NORMED, 5: CCOEFF_NORMED. Default: 5.
	Method NodeTemplateMatchMethod `json:"method,omitempty"`
	// GreenMask enables green color masking for transparent areas.
	GreenMask bool `json:"green_mask,omitempty"`
}

func (n NodeTemplateMatchParam) isRecognitionParam() {}

// TemplateMatchOption is a functional option for configuring NodeTemplateMatchParam.
type TemplateMatchOption func(*NodeTemplateMatchParam)

// WithTemplateMatchROI sets the region of interest for template matching.
func WithTemplateMatchROI(roi Target) TemplateMatchOption {
	return func(param *NodeTemplateMatchParam) {
		param.ROI = roi
	}
}

// WithTemplateMatchROIOffset sets the offset applied to ROI.
func WithTemplateMatchROIOffset(offset Rect) TemplateMatchOption {
	return func(param *NodeTemplateMatchParam) {
		param.ROIOffset = offset
	}
}

// WithTemplateMatchThreshold sets the matching threshold.
func WithTemplateMatchThreshold(threshold []float64) TemplateMatchOption {
	return func(param *NodeTemplateMatchParam) {
		param.Threshold = threshold
	}
}

// WithTemplateMatchOrderBy sets the result ordering method.
func WithTemplateMatchOrderBy(orderBy NodeTemplateMatchOrderBy) TemplateMatchOption {
	return func(param *NodeTemplateMatchParam) {
		param.OrderBy = orderBy
	}
}

// WithTemplateMatchIndex sets which match to select from results.
func WithTemplateMatchIndex(index int) TemplateMatchOption {
	return func(param *NodeTemplateMatchParam) {
		param.Index = index
	}
}

// WithTemplateMatchMethod sets the template matching algorithm.
func WithTemplateMatchMethod(method NodeTemplateMatchMethod) TemplateMatchOption {
	return func(param *NodeTemplateMatchParam) {
		param.Method = method
	}
}

// WithTemplateMatchGreenMask enables green color masking for transparent areas.
func WithTemplateMatchGreenMask(greenMask bool) TemplateMatchOption {
	return func(param *NodeTemplateMatchParam) {
		param.GreenMask = greenMask
	}
}

// RecTemplateMatch creates a TemplateMatch recognition with the given template images.
func RecTemplateMatch(template []string, opts ...TemplateMatchOption) *NodeRecognition {
	param := &NodeTemplateMatchParam{
		Template: template,
	}

	for _, opt := range opts {
		opt(param)
	}

	return &NodeRecognition{
		Type:  NodeRecognitionTypeTemplateMatch,
		Param: param,
	}
}

// NodeFeatureMatchOrderBy defines the ordering options for feature match results.
type NodeFeatureMatchOrderBy string

const (
	NodeFeatureMatchOrderByHorizontal NodeFeatureMatchOrderBy = "Horizontal" // Order by x coordinate (default)
	NodeFeatureMatchOrderByVertical   NodeFeatureMatchOrderBy = "Vertical"   // Order by y coordinate
	NodeFeatureMatchOrderByScore      NodeFeatureMatchOrderBy = "Score"      // Order by matching score
	NodeFeatureMatchOrderByArea       NodeFeatureMatchOrderBy = "Area"       // Order by bounding box area
	NodeFeatureMatchOrderByRandom     NodeFeatureMatchOrderBy = "Random"     // Random order
)

// NodeFeatureMatchDetector defines the feature detection algorithms.
type NodeFeatureMatchDetector string

const (
	NodeFeatureMatchMethodSIFT  NodeFeatureMatchDetector = "SIFT"  // Scale-Invariant Feature Transform (default, most accurate)
	NodeFeatureMatchMethodKAZE  NodeFeatureMatchDetector = "KAZE"  // KAZE features for 2D/3D images
	NodeFeatureMatchMethodAKAZE NodeFeatureMatchDetector = "AKAZE" // Accelerated KAZE
	NodeFeatureMatchMethodBRISK NodeFeatureMatchDetector = "BRISK" // Binary Robust Invariant Scalable Keypoints (fast)
	NodeFeatureMatchMethodORB   NodeFeatureMatchDetector = "ORB"   // Oriented FAST and Rotated BRIEF (fast, no scale invariance)
)

// NodeFeatureMatchParam defines parameters for feature matching recognition.
type NodeFeatureMatchParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitempty"`
	// Template specifies the template image paths. Required.
	Template []string `json:"template,omitempty"`
	// Count specifies the minimum number of feature points required (threshold). Default: 4.
	Count int `json:"count,omitempty"`
	// OrderBy specifies the result ordering. Default: Horizontal.
	OrderBy NodeFeatureMatchOrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
	// GreenMask enables green color masking for transparent areas.
	GreenMask bool `json:"green_mask,omitempty"`
	// Detector specifies the feature detector algorithm. Options: SIFT, KAZE, AKAZE, BRISK, ORB. Default: SIFT.
	Detector NodeFeatureMatchDetector `json:"detector,omitempty"`
	// Ratio specifies the matching ratio threshold [0-1.0]. Default: 0.6.
	Ratio float64 `json:"ratio,omitempty"`
}

func (n NodeFeatureMatchParam) isRecognitionParam() {}

// FeatureMatchOption is a functional option for configuring NodeFeatureMatchParam.
type FeatureMatchOption func(*NodeFeatureMatchParam)

// WithFeatureMatchROI sets the region of interest for feature matching.
func WithFeatureMatchROI(roi Target) FeatureMatchOption {
	return func(param *NodeFeatureMatchParam) {
		param.ROI = roi
	}
}

// WithFeatureMatchROIOffset sets the offset applied to ROI.
func WithFeatureMatchROIOffset(offset Rect) FeatureMatchOption {
	return func(param *NodeFeatureMatchParam) {
		param.ROIOffset = offset
	}
}

// WithFeatureMatchCount sets the minimum number of feature points required (threshold).
func WithFeatureMatchCount(count int) FeatureMatchOption {
	return func(param *NodeFeatureMatchParam) {
		param.Count = count
	}
}

// WithFeatureMatchOrderBy sets the result ordering method.
func WithFeatureMatchOrderBy(orderBy NodeFeatureMatchOrderBy) FeatureMatchOption {
	return func(param *NodeFeatureMatchParam) {
		param.OrderBy = orderBy
	}
}

// WithFeatureMatchIndex sets which match to select from results.
func WithFeatureMatchIndex(index int) FeatureMatchOption {
	return func(param *NodeFeatureMatchParam) {
		param.Index = index
	}
}

// WithFeatureMatchGreenMask enables green color masking for transparent areas.
func WithFeatureMatchGreenMask(greenMask bool) FeatureMatchOption {
	return func(param *NodeFeatureMatchParam) {
		param.GreenMask = greenMask
	}
}

// WithFeatureMatchDetector sets the feature detection algorithm.
func WithFeatureMatchDetector(detector NodeFeatureMatchDetector) FeatureMatchOption {
	return func(param *NodeFeatureMatchParam) {
		param.Detector = detector
	}
}

// WithFeatureMatchRatio sets the KNN matching distance ratio threshold.
func WithFeatureMatchRatio(ratio float64) FeatureMatchOption {
	return func(param *NodeFeatureMatchParam) {
		param.Ratio = ratio
	}
}

// RecFeatureMatch creates a FeatureMatch recognition with the given template images.
// Feature matching provides better generalization with perspective and scale invariance.
func RecFeatureMatch(template []string, opts ...FeatureMatchOption) *NodeRecognition {
	param := &NodeFeatureMatchParam{
		Template: template,
	}

	for _, opt := range opts {
		opt(param)
	}

	return &NodeRecognition{
		Type:  NodeRecognitionTypeFeatureMatch,
		Param: param,
	}
}

// NodeColorMatchMethod defines the color space for color matching (cv::ColorConversionCodes).
type NodeColorMatchMethod int

const (
	NodeColorMatchMethodRGB  NodeColorMatchMethod = 4  // RGB color space, 3 channels (default)
	NodeColorMatchMethodHSV  NodeColorMatchMethod = 40 // HSV color space, 3 channels
	NodeColorMatchMethodGRAY NodeColorMatchMethod = 6  // Grayscale, 1 channel
)

// NodeColorMatchOrderBy defines the ordering options for color match results.
type NodeColorMatchOrderBy string

const (
	NodeColorMatchOrderByHorizontal NodeColorMatchOrderBy = "Horizontal" // Order by x coordinate (default)
	NodeColorMatchOrderByVertical   NodeColorMatchOrderBy = "Vertical"   // Order by y coordinate
	NodeColorMatchOrderByScore      NodeColorMatchOrderBy = "Score"      // Order by matching score
	NodeColorMatchOrderByArea       NodeColorMatchOrderBy = "Area"       // Order by region area
	NodeColorMatchOrderByRandom     NodeColorMatchOrderBy = "Random"     // Random order
)

// NodeColorMatchParam defines parameters for color matching recognition.
type NodeColorMatchParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitempty"`
	// Method specifies the color space. 4: RGB (default), 40: HSV, 6: GRAY.
	Method NodeColorMatchMethod `json:"method,omitempty"`
	// Lower specifies the color lower bounds. Required. Inner array length must match method channels.
	Lower [][]int `json:"lower,omitempty"`
	// Upper specifies the color upper bounds. Required. Inner array length must match method channels.
	Upper [][]int `json:"upper,omitempty"`
	// Count specifies the minimum pixel count required (threshold). Default: 1.
	Count int `json:"count,omitempty"`
	// OrderBy specifies the result ordering. Default: Horizontal.
	OrderBy NodeColorMatchOrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
	// Connected enables connected component analysis. Default: false.
	Connected bool `json:"connected,omitempty"`
}

func (n NodeColorMatchParam) isRecognitionParam() {}

// ColorMatchOption is a functional option for configuring NodeColorMatchParam.
type ColorMatchOption func(*NodeColorMatchParam)

// WithColorMatchROI sets the region of interest for color matching.
func WithColorMatchROI(roi Target) ColorMatchOption {
	return func(param *NodeColorMatchParam) {
		param.ROI = roi
	}
}

// WithColorMatchROIOffset sets the offset applied to ROI.
func WithColorMatchROIOffset(offset Rect) ColorMatchOption {
	return func(param *NodeColorMatchParam) {
		param.ROIOffset = offset
	}
}

// WithColorMatchMethod sets the color space for matching.
func WithColorMatchMethod(method NodeColorMatchMethod) ColorMatchOption {
	return func(param *NodeColorMatchParam) {
		param.Method = method
	}
}

// WithColorMatchCount sets the minimum pixel count required (threshold).
func WithColorMatchCount(count int) ColorMatchOption {
	return func(param *NodeColorMatchParam) {
		param.Count = count
	}
}

// WithColorMatchOrderBy sets the result ordering method.
func WithColorMatchOrderBy(orderBy NodeColorMatchOrderBy) ColorMatchOption {
	return func(param *NodeColorMatchParam) {
		param.OrderBy = orderBy
	}
}

// WithColorMatchIndex sets which match to select from results.
func WithColorMatchIndex(index int) ColorMatchOption {
	return func(param *NodeColorMatchParam) {
		param.Index = index
	}
}

// WithColorMatchConnected enables connected component analysis.
func WithColorMatchConnected(connected bool) ColorMatchOption {
	return func(param *NodeColorMatchParam) {
		param.Connected = connected
	}
}

// RecColorMatch creates a ColorMatch recognition with the given color bounds.
func RecColorMatch(lower, upper [][]int, opts ...ColorMatchOption) *NodeRecognition {
	param := &NodeColorMatchParam{
		Lower: lower,
		Upper: upper,
	}

	for _, opt := range opts {
		opt(param)
	}

	return &NodeRecognition{
		Type:  NodeRecognitionTypeColorMatch,
		Param: param,
	}
}

// NodeOCROrderBy defines the ordering options for OCR results.
type NodeOCROrderBy string

const (
	NodeOCROrderByHorizontal NodeOCROrderBy = "Horizontal" // Order by x coordinate (default)
	NodeOCROrderByVertical   NodeOCROrderBy = "Vertical"   // Order by y coordinate
	NodeOCROrderByArea       NodeOCROrderBy = "Area"       // Order by text region area
	NodeOCROrderByLength     NodeOCROrderBy = "Length"     // Order by text length
	NodeOCROrderByRandom     NodeOCROrderBy = "Random"     // Random order
)

// NodeOCRParam defines parameters for OCR text recognition.
type NodeOCRParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitempty"`
	// Expected specifies the expected text results, supports regex. Required.
	Expected []string `json:"expected,omitempty"`
	// Threshold specifies the model confidence threshold [0-1.0]. Default: 0.3.
	Threshold float64 `json:"threshold,omitempty"`
	// Replace specifies text replacement rules for correcting OCR errors.
	Replace [][2]string `json:"replace,omitempty"`
	// OrderBy specifies the result ordering. Default: Horizontal.
	OrderBy NodeOCROrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
	// OnlyRec enables recognition-only mode without detection (requires precise ROI). Default: false.
	OnlyRec bool `json:"only_rec,omitempty"`
	// Model specifies the model folder path relative to model/ocr directory.
	Model string `json:"model,omitempty"`
}

func (n NodeOCRParam) isRecognitionParam() {}

// OCROption is a functional option for configuring NodeOCRParam.
type OCROption func(*NodeOCRParam)

// WithOCRROI sets the region of interest for OCR.
func WithOCRROI(roi Target) OCROption {
	return func(param *NodeOCRParam) {
		param.ROI = roi
	}
}

// WithOCRROIOffset sets the offset applied to ROI.
func WithOCRROIOffset(offset Rect) OCROption {
	return func(param *NodeOCRParam) {
		param.ROIOffset = offset
	}
}

// WithOCRExpected sets the expected text results.
func WithOCRExpected(expected []string) OCROption {
	return func(param *NodeOCRParam) {
		param.Expected = expected
	}
}

// WithOCRThreshold sets the model confidence threshold.
func WithOCRThreshold(th float64) OCROption {
	return func(param *NodeOCRParam) {
		param.Threshold = th
	}
}

// WithOCRReplace sets text replacement rules for correcting OCR errors.
func WithOCRReplace(replace [][2]string) OCROption {
	return func(param *NodeOCRParam) {
		param.Replace = replace
	}
}

// WithOCROrderBy sets the result ordering method.
func WithOCROrderBy(orderBy NodeOCROrderBy) OCROption {
	return func(param *NodeOCRParam) {
		param.OrderBy = orderBy
	}
}

// WithOCRIndex sets which match to select from results.
func WithOCRIndex(index int) OCROption {
	return func(param *NodeOCRParam) {
		param.Index = index
	}
}

// WithOCROnlyRec enables recognition-only mode without text detection.
func WithOCROnlyRec(only bool) OCROption {
	return func(param *NodeOCRParam) {
		param.OnlyRec = only
	}
}

// WithOCRModel sets the model folder path.
func WithOCRModel(model string) OCROption {
	return func(param *NodeOCRParam) {
		param.Model = model
	}
}

// RecOCR creates an OCR recognition with the given expected text patterns.
func RecOCR(opts ...OCROption) *NodeRecognition {
	param := &NodeOCRParam{}

	for _, opt := range opts {
		opt(param)
	}

	return &NodeRecognition{
		Type:  NodeRecognitionTypeOCR,
		Param: param,
	}
}

// NodeNeuralNetworkClassifyOrderBy defines the ordering options for classification results.
type NodeNeuralNetworkClassifyOrderBy string

const (
	NodeNeuralNetworkClassifyOrderByHorizontal NodeNeuralNetworkClassifyOrderBy = "Horizontal" // Order by x coordinate (default)
	NodeNeuralNetworkClassifyOrderByVertical   NodeNeuralNetworkClassifyOrderBy = "Vertical"   // Order by y coordinate
	NodeNeuralNetworkClassifyOrderByScore      NodeNeuralNetworkClassifyOrderBy = "Score"      // Order by confidence score
	NodeNeuralNetworkClassifyOrderByRandom     NodeNeuralNetworkClassifyOrderBy = "Random"     // Random order
)

// NodeNeuralNetworkClassifyParam defines parameters for neural network classification.
type NodeNeuralNetworkClassifyParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitempty"`
	// Labels specifies the class names for debugging and logging. Fills "Unknown" if not provided.
	Labels []string `json:"labels,omitempty"`
	// Model specifies the model folder path relative to model/classify directory. Required. Only ONNX models supported.
	Model string `json:"model,omitempty"`
	// Expected specifies the expected class indices. Required.
	Expected []int `json:"expected,omitempty"`
	// OrderBy specifies the result ordering. Default: Horizontal.
	OrderBy NodeNeuralNetworkClassifyOrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
}

func (n NodeNeuralNetworkClassifyParam) isRecognitionParam() {}

// NeuralClassifyOption is a functional option for configuring NodeNeuralNetworkClassifyParam.
type NeuralClassifyOption func(*NodeNeuralNetworkClassifyParam)

// WithNeuralClassifyROI sets the region of interest for classification.
func WithNeuralClassifyROI(roi Target) NeuralClassifyOption {
	return func(param *NodeNeuralNetworkClassifyParam) {
		param.ROI = roi
	}
}

// WithNeuralClassifyROIOffset sets the offset applied to ROI.
func WithNeuralClassifyROIOffset(offset Rect) NeuralClassifyOption {
	return func(param *NodeNeuralNetworkClassifyParam) {
		param.ROIOffset = offset
	}
}

// WithNeuralClassifyLabels sets the class names for debugging and logging.
func WithNeuralClassifyLabels(labels []string) NeuralClassifyOption {
	return func(param *NodeNeuralNetworkClassifyParam) {
		param.Labels = labels
	}
}

// WithNeuralClassifyExpected sets the expected class indices.
func WithNeuralClassifyExpected(expected []int) NeuralClassifyOption {
	return func(param *NodeNeuralNetworkClassifyParam) {
		param.Expected = expected
	}
}

// WithNeuralClassifyOrderBy sets the result ordering method.
func WithNeuralClassifyOrderBy(orderBy NodeNeuralNetworkClassifyOrderBy) NeuralClassifyOption {
	return func(param *NodeNeuralNetworkClassifyParam) {
		param.OrderBy = orderBy
	}
}

// WithNeuralClassifyIndex sets which match to select from results.
func WithNeuralClassifyIndex(index int) NeuralClassifyOption {
	return func(param *NodeNeuralNetworkClassifyParam) {
		param.Index = index
	}
}

// RecNeuralNetworkClassify creates a NeuralNetworkClassify recognition.
// This classifies images at fixed positions into predefined categories.
func RecNeuralNetworkClassify(model string, opts ...NeuralClassifyOption) *NodeRecognition {
	param := &NodeNeuralNetworkClassifyParam{
		Model: model,
	}

	for _, opt := range opts {
		opt(param)
	}

	return &NodeRecognition{
		Type:  NodeRecognitionTypeNeuralNetworkClassify,
		Param: param,
	}
}

// NodeNeuralNetworkDetectOrderBy defines the ordering options for detection results.
type NodeNeuralNetworkDetectOrderBy string

const (
	NodeNeuralNetworkDetectOrderByHorizontal NodeNeuralNetworkDetectOrderBy = "Horizontal" // Order by x coordinate (default)
	NodeNeuralNetworkDetectOrderByVertical   NodeNeuralNetworkDetectOrderBy = "Vertical"   // Order by y coordinate
	NodeNeuralNetworkDetectOrderByScore      NodeNeuralNetworkDetectOrderBy = "Score"      // Order by confidence score
	NodeNeuralNetworkDetectOrderByArea       NodeNeuralNetworkDetectOrderBy = "Area"       // Order by bounding box area
	NodeNeuralNetworkDetectOrderByRandom     NodeNeuralNetworkDetectOrderBy = "Random"     // Random order
)

// NodeNeuralNetworkDetectParam defines parameters for neural network object detection.
type NodeNeuralNetworkDetectParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitempty"`
	// Labels specifies the class names for debugging and logging. Auto-reads from model metadata if not provided.
	Labels []string `json:"labels,omitempty"`
	// Model specifies the model folder path relative to model/detect directory. Required. Supports YOLOv8/YOLOv11 ONNX models.
	Model string `json:"model,omitempty"`
	// Expected specifies the expected class indices. Required.
	Expected []int `json:"expected,omitempty"`
	// OrderBy specifies the result ordering. Default: Horizontal.
	OrderBy NodeNeuralNetworkDetectOrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
}

func (n NodeNeuralNetworkDetectParam) isRecognitionParam() {}

// NeuralDetectOption is a functional option for configuring NodeNeuralNetworkDetectParam.
type NeuralDetectOption func(*NodeNeuralNetworkDetectParam)

// WithNeuralDetectROI sets the region of interest for detection.
func WithNeuralDetectROI(roi Target) NeuralDetectOption {
	return func(param *NodeNeuralNetworkDetectParam) {
		param.ROI = roi
	}
}

// WithNeuralDetectROIOffset sets the offset applied to ROI.
func WithNeuralDetectROIOffset(offset Rect) NeuralDetectOption {
	return func(param *NodeNeuralNetworkDetectParam) {
		param.ROIOffset = offset
	}
}

// WithNeuralDetectLabels sets the class names for debugging and logging.
func WithNeuralDetectLabels(labels []string) NeuralDetectOption {
	return func(param *NodeNeuralNetworkDetectParam) {
		param.Labels = labels
	}
}

// WithNeuralDetectExpected sets the expected class indices.
func WithNeuralDetectExpected(expected []int) NeuralDetectOption {
	return func(param *NodeNeuralNetworkDetectParam) {
		param.Expected = expected
	}
}

// WithNeuralDetectOrderBy sets the result ordering method.
func WithNeuralDetectOrderBy(orderBy NodeNeuralNetworkDetectOrderBy) NeuralDetectOption {
	return func(param *NodeNeuralNetworkDetectParam) {
		param.OrderBy = orderBy
	}
}

// WithNeuralDetectIndex sets which match to select from results.
func WithNeuralDetectIndex(index int) NeuralDetectOption {
	return func(param *NodeNeuralNetworkDetectParam) {
		param.Index = index
	}
}

// RecNeuralNetworkDetect creates a NeuralNetworkDetect recognition.
// This detects objects at arbitrary positions using deep learning models like YOLO.
func RecNeuralNetworkDetect(model string, opts ...NeuralDetectOption) *NodeRecognition {
	param := &NodeNeuralNetworkDetectParam{
		Model: model,
	}

	for _, opt := range opts {
		opt(param)
	}

	return &NodeRecognition{
		Type:  NodeRecognitionTypeNeuralNetworkDetect,
		Param: param,
	}
}

// NodeCustomRecognitionParam defines parameters for custom recognition handlers.
type NodeCustomRecognitionParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitempty"`
	// CustomRecognition specifies the recognizer name registered via MaaResourceRegisterCustomRecognition. Required.
	CustomRecognition string `json:"custom_recognition,omitempty"`
	// CustomRecognitionParam specifies custom parameters passed to the recognition callback.
	CustomRecognitionParam any `json:"custom_recognition_param,omitempty"`
}

func (n NodeCustomRecognitionParam) isRecognitionParam() {}

// CustomRecognitionOption is a functional option for configuring NodeCustomRecognitionParam.
type CustomRecognitionOption func(*NodeCustomRecognitionParam)

// WithCustomRecognitionROI sets the region of interest for custom recognition.
func WithCustomRecognitionROI(roi Target) CustomRecognitionOption {
	return func(param *NodeCustomRecognitionParam) {
		param.ROI = roi
	}
}

// WithCustomRecognitionROIOffset sets the offset applied to ROI.
func WithCustomRecognitionROIOffset(offset Rect) CustomRecognitionOption {
	return func(param *NodeCustomRecognitionParam) {
		param.ROIOffset = offset
	}
}

// WithCustomRecognitionParam sets custom parameters passed to the recognition callback.
func WithCustomRecognitionParam(customParam any) CustomRecognitionOption {
	return func(param *NodeCustomRecognitionParam) {
		param.CustomRecognitionParam = customParam
	}
}

// RecCustom creates a Custom recognition with the given recognizer name.
func RecCustom(name string, opts ...CustomRecognitionOption) *NodeRecognition {
	param := &NodeCustomRecognitionParam{
		CustomRecognition: name,
	}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeRecognition{
		Type:  NodeRecognitionTypeCustom,
		Param: param,
	}
}

// NodeAction defines the action configuration for a node.
type NodeAction struct {
	// Type specifies the action type.
	Type NodeActionType `json:"type,omitempty"`
	// Param specifies the action parameters.
	Param NodeActionParam `json:"param,omitempty"`
}

func (na *NodeAction) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type  NodeActionType  `json:"type,omitempty"`
		Param json.RawMessage `json:"param,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	na.Type = raw.Type

	// If no param provided or null, just return with type set
	if len(raw.Param) == 0 || string(raw.Param) == "null" {
		return nil
	}

	// Unmarshal param based on type
	var param NodeActionParam
	switch na.Type {
	case NodeActionTypeDoNothing, "":
		param = &NodeDoNothingParam{}
	case NodeActionTypeClick:
		param = &NodeClickParam{}
	case NodeActionTypeLongPress:
		param = &NodeLongPressParam{}
	case NodeActionTypeSwipe:
		param = &NodeSwipeParam{}
	case NodeActionTypeMultiSwipe:
		param = &NodeMultiSwipeParam{}
	case NodeActionTypeTouchDown:
		param = &NodeTouchDownParam{}
	case NodeActionTypeTouchMove:
		param = &NodeTouchMoveParam{}
	case NodeActionTypeTouchUp:
		param = &NodeTouchUpParam{}
	case NodeActionTypeClickKey:
		param = &NodeClickKeyParam{}
	case NodeActionTypeLongPressKey:
		param = &NodeLongPressKeyParam{}
	case NodeActionTypeKeyDown:
		param = &NodeKeyDownParam{}
	case NodeActionTypeKeyUp:
		param = &NodeKeyUpParam{}
	case NodeActionTypeInputText:
		param = &NodeInputTextParam{}
	case NodeActionTypeStartApp:
		param = &NodeStartAppParam{}
	case NodeActionTypeStopApp:
		param = &NodeStopAppParam{}
	case NodeActionTypeStopTask:
		param = &NodeStopTaskParam{}
	case NodeActionTypeScroll:
		param = &NodeScrollParam{}
	case NodeActionTypeCommand:
		param = &NodeCommandParam{}
	case NodeActionTypeCustom:
		param = &NodeCustomActionParam{}
	default:
		return errors.New("unsupported action type: " + string(na.Type))
	}

	if err := json.Unmarshal(raw.Param, param); err != nil {
		return err
	}
	na.Param = param
	return nil
}

// NodeActionType defines the available action types.
type NodeActionType string

const (
	NodeActionTypeDoNothing    NodeActionType = "DoNothing"
	NodeActionTypeClick        NodeActionType = "Click"
	NodeActionTypeLongPress    NodeActionType = "LongPress"
	NodeActionTypeSwipe        NodeActionType = "Swipe"
	NodeActionTypeMultiSwipe   NodeActionType = "MultiSwipe"
	NodeActionTypeTouchDown    NodeActionType = "TouchDown"
	NodeActionTypeTouchMove    NodeActionType = "TouchMove"
	NodeActionTypeTouchUp      NodeActionType = "TouchUp"
	NodeActionTypeClickKey     NodeActionType = "ClickKey"
	NodeActionTypeLongPressKey NodeActionType = "LongPressKey"
	NodeActionTypeKeyDown      NodeActionType = "KeyDown"
	NodeActionTypeKeyUp        NodeActionType = "KeyUp"
	NodeActionTypeInputText    NodeActionType = "InputText"
	NodeActionTypeStartApp     NodeActionType = "StartApp"
	NodeActionTypeStopApp      NodeActionType = "StopApp"
	NodeActionTypeStopTask     NodeActionType = "StopTask"
	NodeActionTypeScroll       NodeActionType = "Scroll"
	NodeActionTypeCommand      NodeActionType = "Command"
	NodeActionTypeShell        NodeActionType = "Shell"
	NodeActionTypeCustom       NodeActionType = "Custom"
)

// NodeActionParam is the interface for action parameters.
type NodeActionParam interface {
	isActionParam()
}

// NodeDoNothingParam defines parameters for do-nothing action.
type NodeDoNothingParam struct{}

func (n NodeDoNothingParam) isActionParam() {}

// ActDoNothing creates a DoNothing action that performs no operation.
func ActDoNothing() *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeDoNothing,
		Param: &NodeDoNothingParam{},
	}
}

// NodeClickParam defines parameters for click action.
type NodeClickParam struct {
	// Target specifies the click target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	TargetOffset Rect `json:"target_offset,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeClickParam) isActionParam() {}

// ClickOption is a functional option for configuring NodeClickParam.
type ClickOption func(*NodeClickParam)

// WithClickTarget sets the click target position.
func WithClickTarget(target Target) ClickOption {
	return func(p *NodeClickParam) {
		p.Target = target
	}
}

// WithClickTargetOffset sets additional offset applied to target.
func WithClickTargetOffset(offset Rect) ClickOption {
	return func(p *NodeClickParam) {
		p.TargetOffset = offset
	}
}

// WithClickContact sets the touch point identifier.
func WithClickContact(contact int) ClickOption {
	return func(p *NodeClickParam) {
		p.Contact = contact
	}
}

// ActClick creates a Click action with the given options.
func ActClick(opts ...ClickOption) *NodeAction {
	param := &NodeClickParam{}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{Type: NodeActionTypeClick, Param: param}
}

// NodeLongPressParam defines parameters for long press action.
type NodeLongPressParam struct {
	// Target specifies the long press target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	TargetOffset Rect `json:"target_offset,omitempty"`
	// Duration specifies the long press duration in milliseconds. Default: 1000.
	Duration int64 `json:"duration,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeLongPressParam) isActionParam() {}

// LongPressOption is a functional option for configuring NodeLongPressParam.
type LongPressOption func(*NodeLongPressParam)

// WithLongPressTarget sets the long press target position.
func WithLongPressTarget(target Target) LongPressOption {
	return func(p *NodeLongPressParam) {
		p.Target = target
	}
}

// WithLongPressTargetOffset sets additional offset applied to target.
func WithLongPressTargetOffset(offset Rect) LongPressOption {
	return func(p *NodeLongPressParam) {
		p.TargetOffset = offset
	}
}

// WithLongPressDuration sets the long press duration.
func WithLongPressDuration(d time.Duration) LongPressOption {
	return func(p *NodeLongPressParam) {
		p.Duration = d.Milliseconds()
	}
}

// WithLongPressContact sets the touch point identifier.
func WithLongPressContact(contact int) LongPressOption {
	return func(p *NodeLongPressParam) { p.Contact = contact }
}

// ActLongPress creates a LongPress action with the given options.
func ActLongPress(opts ...LongPressOption) *NodeAction {
	param := &NodeLongPressParam{}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{Type: NodeActionTypeLongPress, Param: param}
}

// NodeSwipeParam defines parameters for swipe action.
type NodeSwipeParam struct {
	// Begin specifies the swipe start position.
	Begin Target `json:"begin,omitzero"`
	// BeginOffset specifies additional offset applied to begin position.
	BeginOffset Rect `json:"begin_offset,omitempty"`
	// End specifies the swipe end position.
	End []Target `json:"end,omitzero"`
	// EndOffset specifies additional offset applied to end position.
	EndOffset []Rect `json:"end_offset,omitempty"`
	// Duration specifies the swipe duration in milliseconds. Default: 200.
	Duration []int64 `json:"duration,omitempty"`
	// EndHold specifies extra wait time at end position before releasing in milliseconds. Default: 0.
	EndHold []int64 `json:"end_hold,omitempty"`
	// OnlyHover enables hover-only mode without press/release actions. Default: false.
	OnlyHover bool `json:"only_hover,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeSwipeParam) isActionParam() {}

// SwipeOption is a functional option for configuring NodeSwipeParam.
type SwipeOption func(*NodeSwipeParam)

// WithSwipeBegin sets the swipe start position.
func WithSwipeBegin(begin Target) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.Begin = begin
	}
}

// WithSwipeBeginOffset sets additional offset applied to begin position.
func WithSwipeBeginOffset(offset Rect) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.BeginOffset = offset
	}
}

// WithSwipeEnd sets the swipe end position.
func WithSwipeEnd(end []Target) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.End = end
	}
}

// WithSwipeEndOffset sets additional offset applied to end position.
func WithSwipeEndOffset(offset []Rect) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.EndOffset = offset
	}
}

// WithSwipeDuration sets the swipe duration.
func WithSwipeDuration(d []time.Duration) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.Duration = make([]int64, len(d))
		for index, duration := range d {
			p.Duration[index] = duration.Milliseconds()
		}
	}
}

// WithSwipeEndHold sets extra wait time at end position before releasing.
func WithSwipeEndHold(d []time.Duration) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.EndHold = make([]int64, len(d))
		for index, duration := range d {
			p.EndHold[index] = duration.Milliseconds()
		}
	}
}

// WithSwipeOnlyHover enables hover-only mode without press/release actions.
func WithSwipeOnlyHover(only bool) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.OnlyHover = only
	}
}

// WithSwipeContact sets the touch point identifier.
func WithSwipeContact(contact int) SwipeOption {
	return func(p *NodeSwipeParam) {
		p.Contact = contact
	}
}

// ActSwipe creates a Swipe action with the given options.
func ActSwipe(opts ...SwipeOption) *NodeAction {
	param := &NodeSwipeParam{}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{
		Type:  NodeActionTypeSwipe,
		Param: param,
	}
}

// NodeMultiSwipeItem defines a single swipe within a multi-swipe action.
type NodeMultiSwipeItem struct {
	// Starting specifies when this swipe starts within the action in milliseconds. Default: 0.
	Starting int64 `json:"starting,omitempty"`
	// Begin specifies the swipe start position.
	Begin Target `json:"begin,omitzero"`
	// BeginOffset specifies additional offset applied to begin position.
	BeginOffset Rect `json:"begin_offset,omitempty"`
	// End specifies the swipe end position.
	End []Target `json:"end,omitzero"`
	// EndOffset specifies additional offset applied to end position.
	EndOffset []Rect `json:"end_offset,omitempty"`
	// Duration specifies the swipe duration in milliseconds. Default: 200.
	Duration []int64 `json:"duration,omitempty"`
	// EndHold specifies extra wait time at end position before releasing in milliseconds. Default: 0.
	EndHold []int64 `json:"end_hold,omitempty"`
	// OnlyHover enables hover-only mode without press/release actions. Default: false.
	OnlyHover bool `json:"only_hover,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index. Win32: mouse button. Default uses array index if 0.
	Contact int `json:"contact,omitempty"`
}

// NodeMultiSwipeParam defines parameters for multi-finger swipe action.
type NodeMultiSwipeParam struct {
	// Swipes specifies the list of swipe items. Required.
	Swipes []NodeMultiSwipeItem `json:"swipes,omitempty"`
}

func (n NodeMultiSwipeParam) isActionParam() {}

// MultiSwipeItemOption is a functional option for configuring NodeMultiSwipeItem.
type MultiSwipeItemOption func(*NodeMultiSwipeItem)

// WithMultiSwipeItemStarting sets when this swipe starts within the action.
func WithMultiSwipeItemStarting(starting time.Duration) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.Starting = starting.Milliseconds()
	}
}

// WithMultiSwipeItemBegin sets the swipe start position.
func WithMultiSwipeItemBegin(begin Target) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.Begin = begin
	}
}

// WithMultiSwipeItemBeginOffset sets additional offset applied to begin position.
func WithMultiSwipeItemBeginOffset(offset Rect) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.BeginOffset = offset
	}
}

// WithMultiSwipeItemEnd sets the swipe end position.
func WithMultiSwipeItemEnd(end []Target) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.End = end
	}
}

// WithMultiSwipeItemEndOffset sets additional offset applied to end position.
func WithMultiSwipeItemEndOffset(offset []Rect) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.EndOffset = offset
	}
}

// WithMultiSwipeItemDuration sets the swipe duration.
func WithMultiSwipeItemDuration(d []time.Duration) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.Duration = make([]int64, len(d))
		for index, duration := range d {
			i.Duration[index] = duration.Milliseconds()
		}
	}
}

// WithMultiSwipeItemEndHold sets extra wait time at end position before releasing.
func WithMultiSwipeItemEndHold(d []time.Duration) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.EndHold = make([]int64, len(d))
		for index, duration := range d {
			i.EndHold[index] = duration.Milliseconds()
		}
	}
}

// WithMultiSwipeItemOnlyHover enables hover-only mode without press/release actions.
func WithMultiSwipeItemOnlyHover(only bool) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.OnlyHover = only
	}
}

// WithMultiSwipeItemContact sets the touch point identifier.
func WithMultiSwipeItemContact(contact int) MultiSwipeItemOption {
	return func(i *NodeMultiSwipeItem) {
		i.Contact = contact
	}
}

// NewMultiSwipeItem creates a new multi-swipe item with the given options.
func NewMultiSwipeItem(opts ...MultiSwipeItemOption) NodeMultiSwipeItem {
	item := NodeMultiSwipeItem{}
	for _, opt := range opts {
		opt(&item)
	}
	return item
}

// ActMultiSwipe creates a MultiSwipe action for multi-finger swipe gestures.
func ActMultiSwipe(swipes ...NodeMultiSwipeItem) *NodeAction {
	param := &NodeMultiSwipeParam{
		Swipes: swipes,
	}
	return &NodeAction{
		Type:  NodeActionTypeMultiSwipe,
		Param: param,
	}
}

// NodeTouchDownParam defines parameters for touch down action.
type NodeTouchDownParam struct {
	// Target specifies the touch target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	TargetOffset Rect `json:"target_offset,omitempty"`
	// Pressure specifies the touch pressure, range depends on controller implementation. Default: 0.
	Pressure int `json:"pressure,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeTouchDownParam) isActionParam() {}

// TouchDownOption is a functional option for configuring NodeTouchDownParam.
type TouchDownOption func(*NodeTouchDownParam)

// WithTouchDownTarget sets the touch target position.
func WithTouchDownTarget(target Target) TouchDownOption {
	return func(p *NodeTouchDownParam) {
		p.Target = target
	}
}

// WithTouchDownTargetOffset sets additional offset applied to target.
func WithTouchDownTargetOffset(offset Rect) TouchDownOption {
	return func(p *NodeTouchDownParam) {
		p.TargetOffset = offset
	}
}

// WithTouchDownPressure sets the touch pressure.
func WithTouchDownPressure(pressure int) TouchDownOption {
	return func(p *NodeTouchDownParam) {
		p.Pressure = pressure
	}
}

// WithTouchDownContact sets the touch point identifier.
func WithTouchDownContact(contact int) TouchDownOption {
	return func(p *NodeTouchDownParam) {
		p.Contact = contact
	}
}

// ActTouchDown creates a TouchDown action with the given options.
func ActTouchDown(opts ...TouchDownOption) *NodeAction {
	param := &NodeTouchDownParam{}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{Type: NodeActionTypeTouchDown, Param: param}
}

// NodeTouchMoveParam defines parameters for touch move action.
type NodeTouchMoveParam struct {
	// Target specifies the touch target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	TargetOffset Rect `json:"target_offset,omitempty"`
	// Pressure specifies the touch pressure, range depends on controller implementation. Default: 0.
	Pressure int `json:"pressure,omitempty"`
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeTouchMoveParam) isActionParam() {}

// TouchMoveOption is a functional option for configuring NodeTouchMoveParam.
type TouchMoveOption func(*NodeTouchMoveParam)

// WithTouchMoveTarget sets the touch target position.
func WithTouchMoveTarget(target Target) TouchMoveOption {
	return func(p *NodeTouchMoveParam) {
		p.Target = target
	}
}

// WithTouchMoveTargetOffset sets additional offset applied to target.
func WithTouchMoveTargetOffset(offset Rect) TouchMoveOption {
	return func(p *NodeTouchMoveParam) {
		p.TargetOffset = offset
	}
}

// WithTouchMovePressure sets the touch pressure.
func WithTouchMovePressure(pressure int) TouchMoveOption {
	return func(p *NodeTouchMoveParam) {
		p.Pressure = pressure
	}
}

// WithTouchMoveContact sets the touch point identifier.
func WithTouchMoveContact(contact int) TouchMoveOption {
	return func(p *NodeTouchMoveParam) {
		p.Contact = contact
	}
}

// ActTouchMove creates a TouchMove action with the given options.
func ActTouchMove(opts ...TouchMoveOption) *NodeAction {
	param := &NodeTouchMoveParam{}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{Type: NodeActionTypeTouchMove, Param: param}
}

// NodeTouchUpParam defines parameters for touch up action.
type NodeTouchUpParam struct {
	// Contact specifies the touch point identifier. Adb: finger index (0=first finger). Win32: mouse button (0=left, 1=right, 2=middle).
	Contact int `json:"contact,omitempty"`
}

func (n NodeTouchUpParam) isActionParam() {}

// TouchUpOption is a functional option for configuring NodeTouchUpParam.
type TouchUpOption func(*NodeTouchUpParam)

// WithTouchUpContact sets the touch point identifier.
func WithTouchUpContact(contact int) TouchUpOption {
	return func(p *NodeTouchUpParam) {
		p.Contact = contact
	}
}

// ActTouchUp creates a TouchUp action with the given options.
func ActTouchUp(opts ...TouchUpOption) *NodeAction {
	param := &NodeTouchUpParam{}

	for _, opt := range opts {
		opt(param)
	}

	return &NodeAction{Type: NodeActionTypeTouchUp, Param: param}
}

// NodeClickKeyParam defines parameters for key click action.
type NodeClickKeyParam struct {
	// Key specifies the virtual key codes to click. Required.
	Key []int `json:"key,omitempty"`
}

func (n NodeClickKeyParam) isActionParam() {}

// ActClickKey creates a ClickKey action with the given virtual key codes.
func ActClickKey(keys []int) *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeClickKey,
		Param: &NodeClickKeyParam{Key: keys},
	}
}

// NodeLongPressKeyParam defines parameters for long press key action.
type NodeLongPressKeyParam struct {
	// Key specifies the virtual key code to press. Required.
	Key []int `json:"key,omitempty"`
	// Duration specifies the long press duration in milliseconds. Default: 1000.
	Duration int64 `json:"duration,omitempty"`
}

func (n NodeLongPressKeyParam) isActionParam() {}

// LongPressKeyOption is a functional option for configuring NodeLongPressKeyParam.
type LongPressKeyOption func(*NodeLongPressKeyParam)

// WithLongPressKeyDuration sets the long press duration.
func WithLongPressKeyDuration(d time.Duration) LongPressKeyOption {
	return func(p *NodeLongPressKeyParam) { p.Duration = d.Milliseconds() }
}

// ActLongPressKey creates a LongPressKey action with the given virtual key code.
func ActLongPressKey(key []int, opts ...LongPressKeyOption) *NodeAction {
	param := &NodeLongPressKeyParam{Key: key}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{Type: NodeActionTypeLongPressKey, Param: param}
}

// NodeKeyDownParam defines parameters for key down action.
type NodeKeyDownParam struct {
	// Key specifies the virtual key code to press down. Required.
	Key int `json:"key,omitempty"`
}

func (n NodeKeyDownParam) isActionParam() {}

// ActKeyDown creates a KeyDown action that presses the key without releasing.
func ActKeyDown(key int) *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeKeyDown,
		Param: &NodeKeyDownParam{Key: key},
	}
}

// NodeKeyUpParam defines parameters for key up action.
type NodeKeyUpParam struct {
	// Key specifies the virtual key code to release. Required.
	Key int `json:"key,omitempty"`
}

func (n NodeKeyUpParam) isActionParam() {}

// ActKeyUp creates a KeyUp action that releases a previously pressed key.
func ActKeyUp(key int) *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeKeyUp,
		Param: &NodeKeyUpParam{Key: key},
	}
}

// NodeInputTextParam defines parameters for text input action.
type NodeInputTextParam struct {
	// InputText specifies the text to input. Some controllers only support ASCII. Required.
	InputText string `json:"input_text,omitempty"`
}

func (n NodeInputTextParam) isActionParam() {}

// ActInputText creates an InputText action with the given text.
func ActInputText(input string) *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeInputText,
		Param: &NodeInputTextParam{InputText: input},
	}
}

// NodeStartAppParam defines parameters for start app action.
type NodeStartAppParam struct {
	// Package specifies the package name or activity to start. Required.
	Package string `json:"package,omitempty"`
}

func (n NodeStartAppParam) isActionParam() {}

// ActStartApp creates a StartApp action with the given package name or activity.
func ActStartApp(pkg string) *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeStartApp,
		Param: &NodeStartAppParam{Package: pkg},
	}
}

// NodeStopAppParam defines parameters for stop app action.
type NodeStopAppParam struct {
	// Package specifies the package name to stop. Required.
	Package string `json:"package,omitempty"`
}

func (n NodeStopAppParam) isActionParam() {}

// ActStopApp creates a StopApp action with the given package name.
func ActStopApp(pkg string) *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeStopApp,
		Param: &NodeStopAppParam{Package: pkg},
	}
}

// NodeStopTaskParam defines parameters for stop task action.
// This action stops the current task chain.
type NodeStopTaskParam struct{}

func (n NodeStopTaskParam) isActionParam() {}

// ActStopTask creates a StopTask action that stops the current task chain.
func ActStopTask() *NodeAction {
	return &NodeAction{
		Type:  NodeActionTypeStopTask,
		Param: &NodeStopTaskParam{},
	}
}

type NodeScrollParam struct {
	Dx int `json:"dx,omitempty"`
	Dy int `json:"dy,omitempty"`
}

func (n NodeScrollParam) isActionParam() {}

type ScrollOption func(*NodeScrollParam)

func WithScrollDx(dx int) ScrollOption {
	return func(p *NodeScrollParam) {
		p.Dx = dx
	}
}

func WithScrollDy(dy int) ScrollOption {
	return func(p *NodeScrollParam) {
		p.Dy = dy
	}
}

func ActScroll(opts ...ScrollOption) *NodeAction {
	param := &NodeScrollParam{}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{
		Type:  NodeActionTypeScroll,
		Param: param,
	}
}

// NodeCommandParam defines parameters for command execution action.
type NodeCommandParam struct {
	// Exec specifies the program path to execute. Required.
	Exec string `json:"exec,omitempty"`
	// Args specifies the command arguments. Supports runtime placeholders:
	// {ENTRY}: task entry name, {NODE}: current node name,
	// {IMAGE}: screenshot file path, {BOX}: recognition target [x,y,w,h],
	// {RESOURCE_DIR}: last loaded resource directory, {LIBRARY_DIR}: MaaFW library directory.
	Args []string `json:"args,omitempty"`
	// Detach enables detached mode to run without waiting for completion. Default: false.
	Detach bool `json:"detach,omitempty"`
}

func (n NodeCommandParam) isActionParam() {}

// CommandOption is a functional option for configuring NodeCommandParam.
type CommandOption func(*NodeCommandParam)

// WithCommandArgs sets the command arguments.
func WithCommandArgs(args []string) CommandOption {
	return func(p *NodeCommandParam) { p.Args = args }
}

// WithCommandDetach enables detached mode to run without waiting for completion.
func WithCommandDetach(detach bool) CommandOption {
	return func(p *NodeCommandParam) { p.Detach = detach }
}

// ActCommand creates a Command action with the given executable path.
func ActCommand(exec string, opts ...CommandOption) *NodeAction {
	param := &NodeCommandParam{Exec: exec}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{Type: NodeActionTypeCommand, Param: param}
}

// NodeShellParam defines parameters for shell command execution action.
type NodeShellParam struct {
	Cmd string `json:"cmd,omitempty"`
}

func (n NodeShellParam) isActionParam() {}

// ActShell creates a Shell action with the given command.
// This is only valid for ADB controllers. If the controller is not an ADB controller, the action will fail.
// The output of the command can be obtained in the action detail by MaaTaskerGetActionDetail.
func ActShell(cmd string) *NodeAction {
	return &NodeAction{Type: NodeActionTypeShell, Param: &NodeShellParam{Cmd: cmd}}
}

// NodeCustomActionParam defines parameters for custom action handlers.
type NodeCustomActionParam struct {
	// Target specifies the action target position.
	Target Target `json:"target,omitzero"`
	// TargetOffset specifies additional offset applied to target.
	TargetOffset Rect `json:"target_offset,omitempty"`
	// CustomAction specifies the action name registered via MaaResourceRegisterCustomAction. Required.
	CustomAction string `json:"custom_action,omitempty"`
	// CustomActionParam specifies custom parameters passed to the action callback.
	CustomActionParam any `json:"custom_action_param,omitempty"`
}

func (n NodeCustomActionParam) isActionParam() {}

// CustomActionOption is a functional option for configuring NodeCustomActionParam.
type CustomActionOption func(*NodeCustomActionParam)

// WithCustomActionTarget sets the action target position.
func WithCustomActionTarget(target Target) CustomActionOption {
	return func(param *NodeCustomActionParam) {
		param.Target = target
	}
}

// WithCustomActionTargetOffset sets additional offset applied to target.
func WithCustomActionTargetOffset(offset Rect) CustomActionOption {
	return func(param *NodeCustomActionParam) {
		param.TargetOffset = offset
	}
}

// WithCustomActionParam sets custom parameters passed to the action callback.
func WithCustomActionParam(customParam any) CustomActionOption {
	return func(param *NodeCustomActionParam) {
		param.CustomActionParam = customParam
	}
}

// ActCustom creates a Custom action with the given action name.
func ActCustom(name string, opts ...CustomActionOption) *NodeAction {
	param := &NodeCustomActionParam{
		CustomAction: name,
	}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeAction{
		Type:  NodeActionTypeCustom,
		Param: param,
	}
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
