package maa

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
)

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

	if len(raw.Param) == 0 || string(raw.Param) == "null" {
		return nil
	}

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
	case NodeRecognitionTypeAnd:
		param = &NodeAndRecognitionParam{}
	case NodeRecognitionTypeOr:
		param = &NodeOrRecognitionParam{}
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

// WithBoxIndex sets which sub-recognition result's box to use as the final box.
// Only effective when the recognition type is And.
func (nr *NodeRecognition) WithBoxIndex(idx int) *NodeRecognition {
	if p, ok := nr.Param.(*NodeAndRecognitionParam); ok {
		p.BoxIndex = idx
	}
	return nr
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
	NodeRecognitionTypeAnd                   NodeRecognitionType = "And"
	NodeRecognitionTypeOr                    NodeRecognitionType = "Or"
	NodeRecognitionTypeCustom                NodeRecognitionType = "Custom"
)

// NodeRecognitionParam is the interface for recognition parameters.
type NodeRecognitionParam interface {
	isRecognitionParam()
}

// OrderBy defines the ordering options for recognition results.
// Different recognition types support different subsets of these values.
type OrderBy string

const (
	OrderByHorizontal OrderBy = "Horizontal"
	OrderByVertical   OrderBy = "Vertical"
	OrderByScore      OrderBy = "Score"
	OrderByArea       OrderBy = "Area"
	OrderByLength     OrderBy = "Length"
	OrderByRandom     OrderBy = "Random"
	OrderByExpected   OrderBy = "Expected"
)

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
	OrderBy OrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
	// Method specifies the matching algorithm. 1: SQDIFF_NORMED, 3: CCORR_NORMED, 5: CCOEFF_NORMED. Default: 5.
	Method NodeTemplateMatchMethod `json:"method,omitempty"`
	// GreenMask enables green color masking for transparent areas.
	GreenMask bool `json:"green_mask,omitempty"`
}

func (n NodeTemplateMatchParam) isRecognitionParam() {}

// RecTemplateMatch creates a TemplateMatch recognition with the given parameters.
func RecTemplateMatch(p NodeTemplateMatchParam) *NodeRecognition {
	return &NodeRecognition{
		Type:  NodeRecognitionTypeTemplateMatch,
		Param: &p,
	}
}

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
	OrderBy OrderBy `json:"order_by,omitempty"`
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

// RecFeatureMatch creates a FeatureMatch recognition with the given parameters.
// Feature matching provides better generalization with perspective and scale invariance.
func RecFeatureMatch(p NodeFeatureMatchParam) *NodeRecognition {
	return &NodeRecognition{
		Type:  NodeRecognitionTypeFeatureMatch,
		Param: &p,
	}
}

// NodeColorMatchMethod defines the color space for color matching (cv::ColorConversionCodes).
type NodeColorMatchMethod int

const (
	NodeColorMatchMethodRGB  NodeColorMatchMethod = 4  // RGB color space, 3 channels (default)
	NodeColorMatchMethodHSV  NodeColorMatchMethod = 40 // HSV color space, 3 channels
	NodeColorMatchMethodGRAY NodeColorMatchMethod = 6  // Grayscale, 1 channel
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
	OrderBy OrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
	// Connected enables connected component analysis. Default: false.
	Connected bool `json:"connected,omitempty"`
}

func (n NodeColorMatchParam) isRecognitionParam() {}

// RecColorMatch creates a ColorMatch recognition with the given parameters.
func RecColorMatch(p NodeColorMatchParam) *NodeRecognition {
	return &NodeRecognition{
		Type:  NodeRecognitionTypeColorMatch,
		Param: &p,
	}
}

// NodeOCRParam defines parameters for OCR text recognition.
type NodeOCRParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitempty"`
	// Expected specifies the expected text results, supports regex.
	Expected []string `json:"expected,omitempty"`
	// Threshold specifies the model confidence threshold [0-1.0]. Default: 0.3.
	Threshold float64 `json:"threshold,omitempty"`
	// Replace specifies text replacement rules for correcting OCR errors.
	Replace [][2]string `json:"replace,omitempty"`
	// OrderBy specifies the result ordering. Default: Horizontal.
	OrderBy OrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
	// OnlyRec enables recognition-only mode without detection (requires precise ROI). Default: false.
	OnlyRec bool `json:"only_rec,omitempty"`
	// Model specifies the model folder path relative to model/ocr directory.
	Model string `json:"model,omitempty"`
	// ColorFilter specifies a ColorMatch node name whose color parameters (method, lower, upper)
	// are used to binarize the image before OCR. Nodes with this field set will not participate in batch optimization.
	ColorFilter string `json:"color_filter,omitempty"`
}

func (n NodeOCRParam) isRecognitionParam() {}

// RecOCR creates an OCR recognition with the given parameters.
// All fields are optional; pass no argument for defaults.
func RecOCR(p ...NodeOCRParam) *NodeRecognition {
	var param NodeOCRParam
	if len(p) > 0 {
		param = p[0]
	}
	return &NodeRecognition{
		Type:  NodeRecognitionTypeOCR,
		Param: &param,
	}
}

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
	OrderBy OrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
}

func (n NodeNeuralNetworkClassifyParam) isRecognitionParam() {}

// RecNeuralNetworkClassify creates a NeuralNetworkClassify recognition with the given parameters.
// This classifies images at fixed positions into predefined categories.
func RecNeuralNetworkClassify(p NodeNeuralNetworkClassifyParam) *NodeRecognition {
	return &NodeRecognition{
		Type:  NodeRecognitionTypeNeuralNetworkClassify,
		Param: &p,
	}
}

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
	OrderBy OrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
}

func (n NodeNeuralNetworkDetectParam) isRecognitionParam() {}

// RecNeuralNetworkDetect creates a NeuralNetworkDetect recognition with the given parameters.
// This detects objects at arbitrary positions using deep learning models like YOLO.
func RecNeuralNetworkDetect(p NodeNeuralNetworkDetectParam) *NodeRecognition {
	return &NodeRecognition{
		Type:  NodeRecognitionTypeNeuralNetworkDetect,
		Param: &p,
	}
}

// SubRecognitionItem is one element of And all_of / Or any_of.
// It is either a node name (string reference) or an inline recognition (object with type, param, sub_name).
// GetNodeData from C++ outputs: all_of/any_of as array of string | object; this type supports both.
type SubRecognitionItem struct {
	// NodeName is set when the JSON value is a string (reference to another node by name).
	NodeName string
	// Inline is set when the JSON value is an object (inline recognition with type, param, sub_name).
	Inline *InlineSubRecognition
}

// UnmarshalJSON supports both string (node name) and object (inline recognition).
func (s *SubRecognitionItem) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil
	}
	if trimmed[0] == '"' {
		var nodeName string
		if err := json.Unmarshal(data, &nodeName); err != nil {
			return err
		}
		s.NodeName = nodeName
		s.Inline = nil
		return nil
	}
	if trimmed[0] == '{' {
		inline := &InlineSubRecognition{}
		if err := json.Unmarshal(data, inline); err != nil {
			return err
		}
		s.NodeName = ""
		s.Inline = inline
		return nil
	}
	return errors.New("SubRecognitionItem: expected string or object")
}

// MarshalJSON outputs a string when NodeName is set, otherwise the inline object.
func (s SubRecognitionItem) MarshalJSON() ([]byte, error) {
	if s.NodeName != "" {
		return json.Marshal(s.NodeName)
	}
	if s.Inline != nil {
		return json.Marshal(s.Inline)
	}
	return []byte("null"), nil
}

// Ref returns a SubRecognitionItem that references another node by name.
func Ref(nodeName string) SubRecognitionItem {
	return SubRecognitionItem{NodeName: nodeName}
}

// Inline builds a SubRecognitionItem from a recognition; optional name is the sub_name.
// Example: RecOr(Inline(RecTemplateMatch(...)), Inline(RecColorMatch(...)))
// Example: RecAnd(Ref("A"), Inline(RecDirectHit(), "sub1")).WithBoxIndex(2)
func Inline(rec *NodeRecognition, name ...string) SubRecognitionItem {
	subName := ""
	if len(name) > 0 {
		subName = name[0]
	}
	return SubRecognitionItem{Inline: newInlineSub(subName, rec)}
}

// InlineSubRecognition is an inline sub-recognition element (object form in all_of/any_of).
// It has sub_name plus type and param; used for both And and Or.
type InlineSubRecognition struct {
	SubName string `json:"sub_name,omitempty"`
	NodeRecognition
}

func (n *InlineSubRecognition) UnmarshalJSON(data []byte) error {
	type Alias struct {
		SubName string `json:"sub_name,omitempty"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	n.SubName = alias.SubName

	if err := json.Unmarshal(data, &n.NodeRecognition); err != nil {
		return err
	}
	return nil
}

func newInlineSub(subName string, recognition *NodeRecognition) *InlineSubRecognition {
	return &InlineSubRecognition{
		SubName:         subName,
		NodeRecognition: *recognition,
	}
}

// NodeAndRecognitionParam defines parameters for AND recognition.
// AllOf elements are either node name strings or inline recognitions.
type NodeAndRecognitionParam struct {
	AllOf    []SubRecognitionItem `json:"all_of,omitempty"`
	BoxIndex int                  `json:"box_index,omitempty"`
}

func (n NodeAndRecognitionParam) isRecognitionParam() {}

// RecAnd creates an AND recognition that requires all sub-recognitions to succeed.
// Use WithBoxIndex to set which result's box to use.
// Example: RecAnd(Ref("NodeA"), Inline(RecDirectHit(), "sub1")).WithBoxIndex(2)
func RecAnd(items ...SubRecognitionItem) *NodeRecognition {
	param := &NodeAndRecognitionParam{AllOf: slices.Clone(items)}
	return &NodeRecognition{Type: NodeRecognitionTypeAnd, Param: param}
}

// NodeOrRecognitionParam defines parameters for OR recognition.
// AnyOf elements are either node name strings or inline recognitions.
type NodeOrRecognitionParam struct {
	AnyOf []SubRecognitionItem `json:"any_of,omitempty"`
}

func (n NodeOrRecognitionParam) isRecognitionParam() {}

// RecOr creates an OR recognition that succeeds if any sub-recognition succeeds.
func RecOr(anyOf ...SubRecognitionItem) *NodeRecognition {
	param := &NodeOrRecognitionParam{
		AnyOf: slices.Clone(anyOf),
	}
	return &NodeRecognition{
		Type:  NodeRecognitionTypeOr,
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

// RecCustom creates a Custom recognition with the given parameters.
func RecCustom(p NodeCustomRecognitionParam) *NodeRecognition {
	return &NodeRecognition{
		Type:  NodeRecognitionTypeCustom,
		Param: &p,
	}
}
