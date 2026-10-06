package maa

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
)

// Recognition defines a node's recognition using the pipeline v2 type/param object format.
// Known types decode to their typed parameters. Unrecognized type names retain
// their parameter JSON as *RawRecognitionParam; this does not establish native support.
// Unknown fields outside param, and unmodeled fields of known parameters, are not retained.
type Recognition struct {
	// Type specifies the recognition algorithm type.
	Type RecognitionType `json:"type,omitempty"`
	// Param specifies the recognition parameters.
	// A nil Param omits param when encoding. For unknown types, an explicit JSON null is retained.
	Param RecognitionParam `json:"param,omitempty"`
}

// UnmarshalJSON decodes a pipeline v2 recognition. Errors in known parameter types
// are returned without falling back to raw JSON. On error the recognition is unchanged.
func (nr *Recognition) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type  RecognitionType `json:"type,omitempty"`
		Param json.RawMessage `json:"param,omitempty"`
	}
	if err := unmarshalJSON(data, &raw); err != nil {
		return err
	}

	param, err := decodeRecognitionParam(raw.Type, raw.Param)
	if err != nil {
		return err
	}
	*nr = Recognition{Type: raw.Type, Param: param}
	return nil
}

func decodeRecognitionParam(recognitionType RecognitionType, data []byte) (RecognitionParam, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var param RecognitionParam
	switch recognitionType {
	case RecognitionTypeDirectHit, "":
		param = &DirectHitParam{}
	case RecognitionTypeTemplateMatch:
		param = &TemplateMatchParam{}
	case RecognitionTypeFeatureMatch:
		param = &FeatureMatchParam{}
	case RecognitionTypeColorMatch:
		param = &ColorMatchParam{}
	case RecognitionTypeOCR:
		param = &OCRParam{}
	case RecognitionTypeNeuralNetworkClassify:
		param = &NeuralNetworkClassifyParam{}
	case RecognitionTypeNeuralNetworkDetect:
		param = &NeuralNetworkDetectParam{}
	case RecognitionTypeAnd:
		param = &AndRecognitionParam{}
	case RecognitionTypeOr:
		param = &OrRecognitionParam{}
	case RecognitionTypeCustom:
		param = &CustomRecognitionParam{}
	default:
		param = new(RawRecognitionParam)
	}

	if _, raw := param.(*RawRecognitionParam); !raw && bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil, nil
	}
	if err := unmarshalJSON(data, param); err != nil {
		return nil, err
	}
	return param, nil
}

// SetBoxIndex sets which sub-recognition result's box to use as the final box.
// Only effective when the recognition type is And.
func (nr *Recognition) SetBoxIndex(idx int) *Recognition {
	if p, ok := nr.Param.(*AndRecognitionParam); ok {
		p.BoxIndex = idx
	}
	return nr
}

// RecognitionType names a pipeline v2 recognition. Constants identify the types
// modeled by this package; other names may be used with RawRecognitionParam if the native library supports them.
type RecognitionType string

const (
	RecognitionTypeDirectHit             RecognitionType = "DirectHit"
	RecognitionTypeTemplateMatch         RecognitionType = "TemplateMatch"
	RecognitionTypeFeatureMatch          RecognitionType = "FeatureMatch"
	RecognitionTypeColorMatch            RecognitionType = "ColorMatch"
	RecognitionTypeOCR                   RecognitionType = "OCR"
	RecognitionTypeNeuralNetworkClassify RecognitionType = "NeuralNetworkClassify"
	RecognitionTypeNeuralNetworkDetect   RecognitionType = "NeuralNetworkDetect"
	RecognitionTypeAnd                   RecognitionType = "And"
	RecognitionTypeOr                    RecognitionType = "Or"
	RecognitionTypeCustom                RecognitionType = "Custom"
)

// RecognitionParam is the interface for typed recognition parameters and RawRecognitionParam.
type RecognitionParam interface {
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

// DirectHitParam defines parameters for direct hit recognition.
// It performs no image matching and uses the first resolved ROI as its result box.
// Recognition fails if the ROI cannot be resolved, for example an unavailable node reference.
type DirectHitParam struct {
	// ROI specifies the region to return. The zero value inherits the existing ROI or defaults to the whole image.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies an offset applied to the ROI.
	// Nil inherits the existing/default offset; a pointer to a zero Rect clears it explicitly.
	ROIOffset *Rect `json:"roi_offset,omitempty"`
}

func (n DirectHitParam) isRecognitionParam() {}

// RecDirectHit creates a DirectHit recognition with the default ROI, without image matching.
// To choose an ROI, set the returned recognition's Param to a DirectHitParam.
func RecDirectHit() *Recognition {
	return &Recognition{
		Type:  RecognitionTypeDirectHit,
		Param: &DirectHitParam{},
	}
}

// TemplateMatchOrderBy defines the ordering options for template matching results.
type TemplateMatchOrderBy OrderBy

const (
	TemplateMatchOrderByHorizontal = TemplateMatchOrderBy(OrderByHorizontal)
	TemplateMatchOrderByVertical   = TemplateMatchOrderBy(OrderByVertical)
	TemplateMatchOrderByScore      = TemplateMatchOrderBy(OrderByScore)
	TemplateMatchOrderByRandom     = TemplateMatchOrderBy(OrderByRandom)
)

// TemplateMatchMethod defines the template matching algorithm (cv::TemplateMatchModes).
type TemplateMatchMethod int

const (
	TemplateMatchMethodSQDIFF_NORMED          TemplateMatchMethod = 1     // Normalized squared difference
	TemplateMatchMethodSQDIFF_NORMED_Inverted TemplateMatchMethod = 10001 // Normalized squared difference (Inverted)
	TemplateMatchMethodCCORR_NORMED           TemplateMatchMethod = 3     // Normalized cross correlation
	TemplateMatchMethodCCOEFF_NORMED          TemplateMatchMethod = 5     // Normalized correlation coefficient (default, most accurate)
)

// TemplateMatchParam defines parameters for template matching recognition.
type TemplateMatchParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitzero"`
	// Template specifies the template image paths. Required.
	Template StringList `json:"template,omitzero"`
	// Threshold specifies the matching threshold [0-1.0]. Default: 0.7.
	Threshold []float64 `json:"threshold,omitzero"`
	// OrderBy specifies how results are sorted. Default: Horizontal. Options: Horizontal | Vertical | Score | Random.
	OrderBy TemplateMatchOrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
	// Method specifies the matching algorithm: 1 (SQDIFF_NORMED), 3 (CCORR_NORMED),
	// 5 (CCOEFF_NORMED, default), or 10001 (inverted SQDIFF_NORMED).
	Method TemplateMatchMethod `json:"method,omitempty"`
	// GreenMask enables green color masking for transparent areas.
	GreenMask bool `json:"green_mask,omitempty"`
}

func (n TemplateMatchParam) isRecognitionParam() {}

// RecTemplateMatch creates a TemplateMatch recognition with the given parameters.
func RecTemplateMatch(p TemplateMatchParam) *Recognition {
	param := p
	param.Template = slices.Clone(p.Template)
	param.Threshold = slices.Clone(p.Threshold)
	return &Recognition{
		Type:  RecognitionTypeTemplateMatch,
		Param: &param,
	}
}

// UnmarshalJSON normalizes a scalar template and a scalar threshold to
// one-element lists. Invalid parameter values leave the receiver unchanged.
func (p *TemplateMatchParam) UnmarshalJSON(data []byte) error {
	raw := struct {
		ROI       Target                  `json:"roi,omitzero"`
		ROIOffset Rect                    `json:"roi_offset,omitzero"`
		Template  StringList              `json:"template,omitzero"`
		Threshold templateMatchThresholds `json:"threshold,omitzero"`
		OrderBy   TemplateMatchOrderBy    `json:"order_by,omitempty"`
		Index     int                     `json:"index,omitempty"`
		Method    TemplateMatchMethod     `json:"method,omitempty"`
		GreenMask bool                    `json:"green_mask,omitempty"`
	}{
		ROI:       p.ROI,
		ROIOffset: p.ROIOffset,
		Template:  slices.Clone(p.Template),
		Threshold: templateMatchThresholds(slices.Clone(p.Threshold)),
		OrderBy:   p.OrderBy,
		Index:     p.Index,
		Method:    p.Method,
		GreenMask: p.GreenMask,
	}
	if err := unmarshalJSON(data, &raw); err != nil {
		return err
	}
	*p = TemplateMatchParam{
		ROI:       raw.ROI,
		ROIOffset: raw.ROIOffset,
		Template:  raw.Template,
		Threshold: []float64(raw.Threshold),
		OrderBy:   raw.OrderBy,
		Index:     raw.Index,
		Method:    raw.Method,
		GreenMask: raw.GreenMask,
	}
	return nil
}

type templateMatchThresholds []float64

func (t *templateMatchThresholds) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	var values []*float64
	if len(data) > 0 && data[0] == '[' {
		if err := unmarshalJSON(data, &values); err != nil {
			return err
		}
	} else {
		var value *float64
		if err := unmarshalJSON(data, &value); err != nil {
			return err
		}
		values = []*float64{value}
	}
	thresholds := make(templateMatchThresholds, len(values))
	for i, value := range values {
		if value == nil {
			return errors.New("template match threshold must contain only numbers")
		}
		thresholds[i] = *value
	}
	*t = thresholds
	return nil
}

// FeatureMatchOrderBy defines the ordering options for feature matching results.
type FeatureMatchOrderBy OrderBy

const (
	FeatureMatchOrderByHorizontal = FeatureMatchOrderBy(OrderByHorizontal)
	FeatureMatchOrderByVertical   = FeatureMatchOrderBy(OrderByVertical)
	FeatureMatchOrderByScore      = FeatureMatchOrderBy(OrderByScore)
	FeatureMatchOrderByArea       = FeatureMatchOrderBy(OrderByArea)
	FeatureMatchOrderByRandom     = FeatureMatchOrderBy(OrderByRandom)
)

// FeatureMatchDetector defines the feature detection algorithms.
type FeatureMatchDetector string

const (
	FeatureMatchMethodSIFT  FeatureMatchDetector = "SIFT"  // Scale-Invariant Feature Transform (default, most accurate)
	FeatureMatchMethodKAZE  FeatureMatchDetector = "KAZE"  // KAZE features for 2D/3D images
	FeatureMatchMethodAKAZE FeatureMatchDetector = "AKAZE" // Accelerated KAZE
	FeatureMatchMethodBRISK FeatureMatchDetector = "BRISK" // Binary Robust Invariant Scalable Keypoints (fast)
	FeatureMatchMethodORB   FeatureMatchDetector = "ORB"   // Oriented FAST and Rotated BRIEF (fast, no scale invariance)
)

// FeatureMatchParam defines parameters for feature matching recognition.
type FeatureMatchParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitzero"`
	// Template specifies the template image paths. Required.
	Template StringList `json:"template,omitzero"`
	// Count specifies the minimum number of feature points required (threshold). Default: 4.
	Count int `json:"count,omitempty"`
	// OrderBy specifies how results are sorted. Default: Horizontal. Options: Horizontal | Vertical | Score | Area | Random.
	OrderBy FeatureMatchOrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
	// GreenMask enables green color masking for transparent areas.
	GreenMask bool `json:"green_mask,omitempty"`
	// Detector specifies the feature detector algorithm. Options: SIFT, KAZE, AKAZE, BRISK, ORB. Default: SIFT.
	Detector FeatureMatchDetector `json:"detector,omitempty"`
	// Ratio specifies the matching ratio threshold [0-1.0]. Default: 0.6.
	Ratio float64 `json:"ratio,omitempty"`
}

func (n FeatureMatchParam) isRecognitionParam() {}

// RecFeatureMatch creates a FeatureMatch recognition with the given parameters.
// Feature matching provides better generalization with perspective and scale invariance.
func RecFeatureMatch(p FeatureMatchParam) *Recognition {
	param := p
	param.Template = slices.Clone(p.Template)
	return &Recognition{
		Type:  RecognitionTypeFeatureMatch,
		Param: &param,
	}
}

// ColorMatchMethod defines the color space for color matching (cv::ColorConversionCodes).
type ColorMatchMethod int

const (
	ColorMatchMethodRGB  ColorMatchMethod = 4  // RGB color space, 3 channels (default)
	ColorMatchMethodHSV  ColorMatchMethod = 40 // HSV color space, 3 channels
	ColorMatchMethodGRAY ColorMatchMethod = 6  // Grayscale, 1 channel
)

// ColorMatchOrderBy defines the ordering options for color matching results.
type ColorMatchOrderBy OrderBy

const (
	ColorMatchOrderByHorizontal = ColorMatchOrderBy(OrderByHorizontal)
	ColorMatchOrderByVertical   = ColorMatchOrderBy(OrderByVertical)
	ColorMatchOrderByScore      = ColorMatchOrderBy(OrderByScore)
	ColorMatchOrderByArea       = ColorMatchOrderBy(OrderByArea)
	ColorMatchOrderByRandom     = ColorMatchOrderBy(OrderByRandom)
)

// ColorMatchParam defines parameters for color matching recognition.
type ColorMatchParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitzero"`
	// Method specifies the color space. 4: RGB (default), 40: HSV, 6: GRAY.
	Method ColorMatchMethod `json:"method,omitempty"`
	// Lower specifies the color lower bounds. Required. Inner array length must match method channels.
	Lower [][]int `json:"lower,omitempty"`
	// Upper specifies the color upper bounds. Required. Inner array length must match method channels.
	Upper [][]int `json:"upper,omitempty"`
	// Count specifies the minimum pixel count required (threshold). Default: 1.
	Count int `json:"count,omitempty"`
	// OrderBy specifies how results are sorted. Default: Horizontal. Options: Horizontal | Vertical | Score | Area | Random.
	OrderBy ColorMatchOrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
	// Connected enables connected component analysis. Default: false.
	Connected bool `json:"connected,omitempty"`
}

func (n ColorMatchParam) isRecognitionParam() {}

// clone2DInt returns a deep copy of s so the caller does not share backing arrays.
func clone2DInt(s [][]int) [][]int {
	if s == nil {
		return nil
	}
	out := make([][]int, len(s))
	for i := range s {
		out[i] = slices.Clone(s[i])
	}
	return out
}

// RecColorMatch creates a ColorMatch recognition with the given parameters.
func RecColorMatch(p ColorMatchParam) *Recognition {
	param := p
	param.Lower = clone2DInt(p.Lower)
	param.Upper = clone2DInt(p.Upper)
	return &Recognition{
		Type:  RecognitionTypeColorMatch,
		Param: &param,
	}
}

// OCROrderBy defines the ordering options for OCR results.
type OCROrderBy OrderBy

const (
	OCROrderByHorizontal = OCROrderBy(OrderByHorizontal)
	OCROrderByVertical   = OCROrderBy(OrderByVertical)
	OCROrderByArea       = OCROrderBy(OrderByArea)
	OCROrderByLength     = OCROrderBy(OrderByLength)
	OCROrderByRandom     = OCROrderBy(OrderByRandom)
	OCROrderByExpected   = OCROrderBy(OrderByExpected)
)

// OCRParam defines parameters for OCR text recognition.
type OCRParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitzero"`
	// Expected specifies the expected text results, supports regex.
	// JSON input may be a single string or an array of strings.
	// Nil is omitted to inherit the existing/default value; an empty
	// non-nil list clears it.
	Expected StringList `json:"expected,omitzero"`
	// Threshold specifies the model confidence threshold [0-1.0]. Default: 0.3.
	Threshold float64 `json:"threshold,omitempty"`
	// Replace specifies text replacement rules for correcting OCR errors.
	Replace [][2]string `json:"replace,omitempty"`
	// OrderBy specifies how results are sorted. Default: Horizontal. Options: Horizontal | Vertical | Area | Length | Random | Expected.
	OrderBy OCROrderBy `json:"order_by,omitempty"`
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

func (n OCRParam) isRecognitionParam() {}

// RecOCR creates an OCR recognition with the given parameters.
// All fields are optional; pass OCRParam{} for defaults.
func RecOCR(p OCRParam) *Recognition {
	param := p
	param.Expected = slices.Clone(p.Expected)
	param.Replace = slices.Clone(p.Replace)
	return &Recognition{
		Type:  RecognitionTypeOCR,
		Param: &param,
	}
}

// NeuralNetworkClassifyOrderBy defines the ordering options for neural network classification results.
type NeuralNetworkClassifyOrderBy OrderBy

const (
	NeuralNetworkClassifyOrderByHorizontal = NeuralNetworkClassifyOrderBy(OrderByHorizontal)
	NeuralNetworkClassifyOrderByVertical   = NeuralNetworkClassifyOrderBy(OrderByVertical)
	NeuralNetworkClassifyOrderByScore      = NeuralNetworkClassifyOrderBy(OrderByScore)
	NeuralNetworkClassifyOrderByRandom     = NeuralNetworkClassifyOrderBy(OrderByRandom)
	NeuralNetworkClassifyOrderByExpected   = NeuralNetworkClassifyOrderBy(OrderByExpected)
)

// NeuralNetworkClassifyParam defines parameters for neural network classification.
type NeuralNetworkClassifyParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitzero"`
	// Labels specifies class names for label selection, debugging, and logging. Fills "Unknown" if not provided.
	// JSON input may be a single string or an array of strings.
	// Nil is omitted to inherit the existing/default labels; an empty
	// non-nil list clears them.
	Labels StringList `json:"labels,omitzero"`
	// Model specifies the model folder path relative to model/classify directory. Required. Only ONNX models supported.
	Model string `json:"model,omitempty"`
	// Expected selects class indices or labels, preserving their order.
	// Nil inherits the existing/default selection; an empty list matches all classes.
	Expected ClassSelectors `json:"expected,omitzero"`
	// OrderBy specifies how results are sorted. Default: Horizontal. Options: Horizontal | Vertical | Score | Random | Expected.
	OrderBy NeuralNetworkClassifyOrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
}

func (n NeuralNetworkClassifyParam) isRecognitionParam() {}

// RecNeuralNetworkClassify creates a NeuralNetworkClassify recognition with the given parameters.
// This classifies images at fixed positions into predefined categories.
func RecNeuralNetworkClassify(p NeuralNetworkClassifyParam) *Recognition {
	param := p
	param.Labels = slices.Clone(p.Labels)
	param.Expected = slices.Clone(p.Expected)
	return &Recognition{
		Type:  RecognitionTypeNeuralNetworkClassify,
		Param: &param,
	}
}

// NeuralNetworkDetectOrderBy defines the ordering options for neural network detection results.
type NeuralNetworkDetectOrderBy OrderBy

const (
	NeuralNetworkDetectOrderByHorizontal = NeuralNetworkDetectOrderBy(OrderByHorizontal)
	NeuralNetworkDetectOrderByVertical   = NeuralNetworkDetectOrderBy(OrderByVertical)
	NeuralNetworkDetectOrderByScore      = NeuralNetworkDetectOrderBy(OrderByScore)
	NeuralNetworkDetectOrderByArea       = NeuralNetworkDetectOrderBy(OrderByArea)
	NeuralNetworkDetectOrderByRandom     = NeuralNetworkDetectOrderBy(OrderByRandom)
	NeuralNetworkDetectOrderByExpected   = NeuralNetworkDetectOrderBy(OrderByExpected)
)

// NeuralNetworkDetectParam defines parameters for neural network object detection.
type NeuralNetworkDetectParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitzero"`
	// Labels specifies class names for label selection, debugging, and logging. Auto-reads from model metadata if not provided.
	// JSON input may be a single string or an array of strings.
	// Nil is omitted to inherit the existing/default labels; an empty
	// non-nil list clears them.
	Labels StringList `json:"labels,omitzero"`
	// Model specifies the model folder path relative to model/detect directory. Required. Supports YOLOv8/YOLOv11 ONNX models.
	Model string `json:"model,omitempty"`
	// Expected selects class indices or labels, preserving their order.
	// Nil inherits the existing/default selection; an empty list matches all
	// classes, subject to Threshold.
	Expected ClassSelectors `json:"expected,omitzero"`
	// Threshold specifies confidence thresholds in [0, 1], in Expected order.
	// Nil or an empty list inherits the existing thresholds or defaults to 0.3.
	// A single threshold applies to all expected classes; otherwise lengths must match.
	// A nonempty list containing 0 sends zero explicitly.
	// JSON input may be a number or an array; encoding always uses an array.
	Threshold []float64 `json:"threshold,omitempty"`
	// OrderBy specifies how results are sorted. Default: Horizontal. Options: Horizontal | Vertical | Score | Area | Random | Expected
	OrderBy NeuralNetworkDetectOrderBy `json:"order_by,omitempty"`
	// Index specifies which match to select from results.
	Index int `json:"index,omitempty"`
}

func (n NeuralNetworkDetectParam) isRecognitionParam() {}

// UnmarshalJSON normalizes a scalar threshold to a one-element list.
// Invalid parameter values leave the receiver unchanged.
func (p *NeuralNetworkDetectParam) UnmarshalJSON(data []byte) error {
	raw := struct {
		ROI       Target                        `json:"roi,omitzero"`
		ROIOffset Rect                          `json:"roi_offset,omitzero"`
		Labels    StringList                    `json:"labels,omitzero"`
		Model     string                        `json:"model,omitempty"`
		Expected  ClassSelectors                `json:"expected,omitzero"`
		Threshold neuralNetworkDetectThresholds `json:"threshold,omitempty"`
		OrderBy   NeuralNetworkDetectOrderBy    `json:"order_by,omitempty"`
		Index     int                           `json:"index,omitempty"`
	}{
		ROI:       p.ROI,
		ROIOffset: p.ROIOffset,
		Labels:    slices.Clone(p.Labels),
		Model:     p.Model,
		Expected:  slices.Clone(p.Expected),
		Threshold: neuralNetworkDetectThresholds(slices.Clone(p.Threshold)),
		OrderBy:   p.OrderBy,
		Index:     p.Index,
	}
	if err := unmarshalJSON(data, &raw); err != nil {
		return err
	}
	*p = NeuralNetworkDetectParam{
		ROI:       raw.ROI,
		ROIOffset: raw.ROIOffset,
		Labels:    raw.Labels,
		Model:     raw.Model,
		Expected:  raw.Expected,
		Threshold: []float64(raw.Threshold),
		OrderBy:   raw.OrderBy,
		Index:     raw.Index,
	}
	return nil
}

type neuralNetworkDetectThresholds []float64

func (t *neuralNetworkDetectThresholds) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	var values []*float64
	if len(data) > 0 && data[0] == '[' {
		if err := unmarshalJSON(data, &values); err != nil {
			return err
		}
	} else {
		var value *float64
		if err := unmarshalJSON(data, &value); err != nil {
			return err
		}
		values = []*float64{value}
	}
	thresholds := make(neuralNetworkDetectThresholds, len(values))
	for i, value := range values {
		if value == nil {
			return errors.New("neural network detection threshold must contain only numbers")
		}
		thresholds[i] = *value
	}
	*t = thresholds
	return nil
}

// RecNeuralNetworkDetect creates a NeuralNetworkDetect recognition with the given parameters.
// This detects objects at arbitrary positions using deep learning models like YOLO.
func RecNeuralNetworkDetect(p NeuralNetworkDetectParam) *Recognition {
	param := p
	param.Labels = slices.Clone(p.Labels)
	param.Expected = slices.Clone(p.Expected)
	param.Threshold = slices.Clone(p.Threshold)
	return &Recognition{
		Type:  RecognitionTypeNeuralNetworkDetect,
		Param: &param,
	}
}

// SubRecognitionItem is one element of And all_of / Or any_of.
// It is either a node name (string reference) or a v2 inline recognition with a recognition object and optional sub_name.
// GetNodeData from C++ outputs: all_of/any_of as array of string | object; this type supports both.
type SubRecognitionItem struct {
	// NodeName is set when the JSON value is a string (reference to another node by name).
	NodeName string
	// Inline is set when the JSON value is an object (inline recognition with type, param, sub_name).
	// The v2 JSON envelope nests type and param under recognition.
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
		if err := unmarshalJSON(data, &nodeName); err != nil {
			return err
		}
		s.NodeName = nodeName
		s.Inline = nil
		return nil
	}
	if trimmed[0] == '{' {
		inline := &InlineSubRecognition{}
		if err := unmarshalJSON(data, inline); err != nil {
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
		return marshalJSON(s.NodeName)
	}
	if s.Inline != nil {
		return marshalJSON(s.Inline)
	}
	return []byte("null"), nil
}

// Ref returns a SubRecognitionItem that references another node by name.
func Ref(nodeName string) SubRecognitionItem {
	return SubRecognitionItem{NodeName: nodeName}
}

// Inline builds a SubRecognitionItem from a recognition; optional name is the sub_name.
// Example: RecOr(Inline(RecTemplateMatch(...)), Inline(RecColorMatch(...)))
// Example: RecAnd(Ref("A"), Inline(RecDirectHit(), "sub1")).SetBoxIndex(2)
func Inline(rec *Recognition, name ...string) SubRecognitionItem {
	subName := ""
	if len(name) > 0 {
		subName = name[0]
	}
	return SubRecognitionItem{Inline: newInlineSub(subName, rec)}
}

// InlineSubRecognition is a v2 inline sub-recognition in all_of/any_of.
// JSON uses {"sub_name": "...", "recognition": {"type": "...", "param": {...}}}.
// SubName is retained by both And and Or; only And uses it to resolve later sub-recognition ROIs.
type InlineSubRecognition struct {
	SubName string `json:"sub_name,omitempty"`
	Recognition
}

// MarshalJSON encodes the v2 recognition envelope used by MaaFramework.
func (n InlineSubRecognition) MarshalJSON() ([]byte, error) {
	return marshalJSON(struct {
		SubName     string      `json:"sub_name,omitempty"`
		Recognition Recognition `json:"recognition"`
	}{SubName: n.SubName, Recognition: n.Recognition})
}

func (n *InlineSubRecognition) UnmarshalJSON(data []byte) error {
	type Alias struct {
		SubName     string          `json:"sub_name,omitempty"`
		Recognition json.RawMessage `json:"recognition,omitempty"`
	}
	var alias Alias
	if err := unmarshalJSON(data, &alias); err != nil {
		return err
	}

	decoded := *n
	decoded.SubName = alias.SubName
	if len(alias.Recognition) > 0 {
		if err := unmarshalJSON(alias.Recognition, &decoded.Recognition); err != nil {
			return err
		}
	} else if err := unmarshalJSON(data, &decoded.Recognition); err != nil {
		return err
	}
	*n = decoded
	return nil
}

func newInlineSub(subName string, recognition *Recognition) *InlineSubRecognition {
	return &InlineSubRecognition{
		SubName:     subName,
		Recognition: *recognition,
	}
}

// AndRecognitionParam defines parameters for AND recognition.
// AllOf elements are either node name strings or inline recognitions.
type AndRecognitionParam struct {
	AllOf    []SubRecognitionItem `json:"all_of,omitempty"`
	BoxIndex int                  `json:"box_index,omitempty"`
}

func (n AndRecognitionParam) isRecognitionParam() {}

// RecAnd creates an AND recognition that requires all sub-recognitions to succeed.
// Use SetBoxIndex to set which result's box to use.
// Example: RecAnd(Ref("NodeA"), Inline(RecDirectHit(), "sub1")).SetBoxIndex(2)
func RecAnd(items ...SubRecognitionItem) *Recognition {
	param := &AndRecognitionParam{AllOf: slices.Clone(items)}
	return &Recognition{Type: RecognitionTypeAnd, Param: param}
}

// OrRecognitionParam defines parameters for OR recognition.
// AnyOf elements are either node name strings or inline recognitions.
type OrRecognitionParam struct {
	AnyOf []SubRecognitionItem `json:"any_of,omitempty"`
}

func (n OrRecognitionParam) isRecognitionParam() {}

// RecOr creates an OR recognition that succeeds if any sub-recognition succeeds.
func RecOr(anyOf ...SubRecognitionItem) *Recognition {
	param := &OrRecognitionParam{
		AnyOf: slices.Clone(anyOf),
	}
	return &Recognition{
		Type:  RecognitionTypeOr,
		Param: param,
	}
}

// CustomRecognitionParam defines parameters for custom recognition handlers.
type CustomRecognitionParam struct {
	// ROI specifies the region of interest for recognition.
	ROI Target `json:"roi,omitzero"`
	// ROIOffset specifies the offset applied to ROI.
	ROIOffset Rect `json:"roi_offset,omitzero"`
	// CustomRecognition specifies the recognizer name registered via MaaResourceRegisterCustomRecognition. Required.
	CustomRecognition string `json:"custom_recognition,omitempty"`
	// CustomRecognitionParam specifies custom parameters passed to the recognition callback.
	CustomRecognitionParam any `json:"custom_recognition_param,omitempty"`
}

func (n CustomRecognitionParam) isRecognitionParam() {}

// RecCustom creates a Custom recognition with the given parameters.
func RecCustom(p CustomRecognitionParam) *Recognition {
	param := p
	return &Recognition{
		Type:  RecognitionTypeCustom,
		Param: &param,
	}
}
