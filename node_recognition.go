package maa

import (
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
		Template: slices.Clone(template),
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
		Template: slices.Clone(template),
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
		Lower: slices.Clone(lower),
		Upper: slices.Clone(upper),
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
		param.Expected = slices.Clone(expected)
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
		param.Replace = slices.Clone(replace)
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
		param.Labels = slices.Clone(labels)
	}
}

// WithNeuralClassifyExpected sets the expected class indices.
func WithNeuralClassifyExpected(expected []int) NeuralClassifyOption {
	return func(param *NodeNeuralNetworkClassifyParam) {
		param.Expected = slices.Clone(expected)
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
		param.Labels = slices.Clone(labels)
	}
}

// WithNeuralDetectExpected sets the expected class indices.
func WithNeuralDetectExpected(expected []int) NeuralDetectOption {
	return func(param *NodeNeuralNetworkDetectParam) {
		param.Expected = slices.Clone(expected)
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

type NodeAndRecognitionItem struct {
	SubName          string `json:"sub_name,omitempty"`
	*NodeRecognition `json:"recognition,omitempty"`
}

// AndItem creates a NodeAndRecognitionItem with the given sub-name and recognition.
// If subName is empty, only the recognition will be used.
func AndItem(subName string, recognition *NodeRecognition) NodeAndRecognitionItem {
	return NodeAndRecognitionItem{
		SubName:         subName,
		NodeRecognition: recognition,
	}
}

// NodeAndRecognitionParam defines parameters for AND recognition.
type NodeAndRecognitionParam struct {
	AllOf    []NodeAndRecognitionItem `json:"all_of,omitempty"`
	BoxIndex int                      `json:"box_index,omitempty"`
}

func (n NodeAndRecognitionParam) isRecognitionParam() {}

// AndRecognitionOption is a functional option for configuring NodeAndRecognitionParam.
type AndRecognitionOption func(*NodeAndRecognitionParam)

// WithAndRecognitionBoxIndex sets which recognition result's box to use as the final box.
func WithAndRecognitionBoxIndex(boxIndex int) AndRecognitionOption {
	return func(param *NodeAndRecognitionParam) {
		param.BoxIndex = boxIndex
	}
}

// RecAnd creates an AND recognition that requires all sub-recognitions to succeed.
func RecAnd(allOf []NodeAndRecognitionItem, opts ...AndRecognitionOption) *NodeRecognition {
	param := &NodeAndRecognitionParam{
		AllOf: slices.Clone(allOf),
	}
	for _, opt := range opts {
		opt(param)
	}
	return &NodeRecognition{
		Type:  NodeRecognitionTypeAnd,
		Param: param,
	}
}

// NodeOrRecognitionParam defines parameters for OR recognition.
type NodeOrRecognitionParam struct {
	AnyOf []*NodeRecognition `json:"any_of,omitempty"`
}

func (n NodeOrRecognitionParam) isRecognitionParam() {}

// RecOr creates an OR recognition that succeeds if any sub-recognition succeeds.
func RecOr(anyOf []*NodeRecognition) *NodeRecognition {
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
