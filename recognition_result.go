package maa

import (
	"encoding/json"
)

type RecognitionResult struct {
	tp  NodeRecognitionType
	val any
}

// Type returns the recognition type of the result.
func (r *RecognitionResult) Type() NodeRecognitionType {
	return r.tp
}

// Value returns the underlying value of the result.
func (r *RecognitionResult) Value() any {
	return r.val
}

// AsTemplateMatch returns the result as TemplateMatchResult if the type matches.
func (r *RecognitionResult) AsTemplateMatch() (*TemplateMatchResult, bool) {
	if r.tp != NodeRecognitionTypeTemplateMatch {
		return nil, false
	}
	val, ok := r.val.(*TemplateMatchResult)
	return val, ok
}

// AsFeatureMatch returns the result as FeatureMatchResult if the type matches.
func (r *RecognitionResult) AsFeatureMatch() (*FeatureMatchResult, bool) {
	if r.tp != NodeRecognitionTypeFeatureMatch {
		return nil, false
	}
	val, ok := r.val.(*FeatureMatchResult)
	return val, ok
}

// AsColorMatch returns the result as ColorMatchResult if the type matches.
func (r *RecognitionResult) AsColorMatch() (*ColorMatchResult, bool) {
	if r.tp != NodeRecognitionTypeColorMatch {
		return nil, false
	}
	val, ok := r.val.(*ColorMatchResult)
	return val, ok
}

// AsOCR returns the result as OCRResult if the type matches.
func (r *RecognitionResult) AsOCR() (*OCRResult, bool) {
	if r.tp != NodeRecognitionTypeOCR {
		return nil, false
	}
	val, ok := r.val.(*OCRResult)
	return val, ok
}

// AsNeuralNetworkClassify returns the result as NeuralNetworkClassifyResult if the type matches.
func (r *RecognitionResult) AsNeuralNetworkClassify() (*NeuralNetworkClassifyResult, bool) {
	if r.tp != NodeRecognitionTypeNeuralNetworkClassify {
		return nil, false
	}
	val, ok := r.val.(*NeuralNetworkClassifyResult)
	return val, ok
}

// AsNeuralNetworkDetect returns the result as NeuralNetworkDetectResult if the type matches.
func (r *RecognitionResult) AsNeuralNetworkDetect() (*NeuralNetworkDetectResult, bool) {
	if r.tp != NodeRecognitionTypeNeuralNetworkDetect {
		return nil, false
	}
	val, ok := r.val.(*NeuralNetworkDetectResult)
	return val, ok
}

func (r *RecognitionResult) AsCustom() (*CustomRecognitionResult, bool) {
	if r.tp != NodeRecognitionTypeCustom {
		return nil, false
	}

	val, ok := r.val.(*CustomRecognitionResult)
	return val, ok
}

type TemplateMatchResult struct {
	Box   Rect    `json:"box"`
	Score float64 `json:"score"`
}

type FeatureMatchResult struct {
	Box   Rect `json:"box"`
	Count int  `json:"count"`
}

type ColorMatchResult struct {
	Box   Rect `json:"box"`
	Count int  `json:"count"`
}

type OCRResult struct {
	Box   Rect    `json:"box"`
	Text  string  `json:"text"`
	Score float64 `json:"score"`
}

type NeuralNetworkClassifyResult struct {
	Box      Rect      `json:"box"`
	ClsIndex uint64    `json:"cls_index"`
	Label    string    `json:"label"`
	Raw      []float64 `json:"raw"`
	Probs    []float64 `json:"probs"`
	Score    float64   `json:"score"`
}

type NeuralNetworkDetectResult struct {
	Box      Rect    `json:"box"`
	ClsIndex uint64  `json:"cls_index"`
	Label    string  `json:"label"`
	Score    float64 `json:"score"`
}

// RecognitionResults contains all, best, and filtered recognition results.
// Detail JSON format: {"all": [Result...], "best": [Result...], "filtered": [Result...]}
// if algorithm is direct hit, Results is nil
type RecognitionResults struct {
	All      []*RecognitionResult `json:"all"`
	Best     []*RecognitionResult `json:"best"`
	Filtered []*RecognitionResult `json:"filtered"`
}

// parseRecognitionResult parses a single result JSON based on the algorithm type.
// Returns nil if parsing fails or the algorithm type is unknown.
func parseRecognitionResult(algorithm string, resultJson []byte) *RecognitionResult {
	algorithmType := NodeRecognitionType(algorithm)

	var resultVal any
	var err error

	switch algorithmType {
	case NodeRecognitionTypeTemplateMatch:
		resultVal = &TemplateMatchResult{}
		err = json.Unmarshal(resultJson, resultVal)
	case NodeRecognitionTypeFeatureMatch:
		resultVal = &FeatureMatchResult{}
		err = json.Unmarshal(resultJson, resultVal)
	case NodeRecognitionTypeColorMatch:
		resultVal = &ColorMatchResult{}
		err = json.Unmarshal(resultJson, resultVal)
	case NodeRecognitionTypeOCR:
		resultVal = &OCRResult{}
		err = json.Unmarshal(resultJson, resultVal)
	case NodeRecognitionTypeNeuralNetworkClassify:
		resultVal = &NeuralNetworkClassifyResult{}
		err = json.Unmarshal(resultJson, resultVal)
	case NodeRecognitionTypeNeuralNetworkDetect:
		resultVal = &NeuralNetworkDetectResult{}
		err = json.Unmarshal(resultJson, resultVal)
	case NodeRecognitionTypeCustom:
		resultVal = &CustomRecognitionResult{}
		err = json.Unmarshal(resultJson, resultVal)
	default:
		return nil
	}

	if err != nil {
		return nil
	}

	return &RecognitionResult{
		tp:  algorithmType,
		val: resultVal,
	}
}

// parseRecognitionResults parses detailJson and returns RecognitionResults containing all, best, and filtered results.
// Detail JSON format: {"all": [Result...], "best": [Result...], "filtered": [Result...]}
func parseRecognitionResults(algorithm, detailJson string) (*RecognitionResults, error) {
	if algorithm == string(NodeRecognitionTypeDirectHit) {
		return nil, nil
	}

	// Handle empty or invalid JSON
	if detailJson == "" || detailJson == "{}" {
		return &RecognitionResults{
			All:      []*RecognitionResult{},
			Best:     []*RecognitionResult{},
			Filtered: []*RecognitionResult{},
		}, nil
	}

	var raw struct {
		All      json.RawMessage `json:"all"`
		Best     json.RawMessage `json:"best"`
		Filtered json.RawMessage `json:"filtered"`
	}

	if err := json.Unmarshal([]byte(detailJson), &raw); err != nil {
		return nil, err
	}

	results := &RecognitionResults{
		All:      make([]*RecognitionResult, 0),
		Best:     make([]*RecognitionResult, 0),
		Filtered: make([]*RecognitionResult, 0),
	}

	// Parse all results
	var allItems []json.RawMessage
	if len(raw.All) > 0 {
		if err := json.Unmarshal(raw.All, &allItems); err == nil {
			for _, item := range allItems {
				if result := parseRecognitionResult(algorithm, item); result != nil {
					results.All = append(results.All, result)
				}
			}
		}
	}

	// Parse best results
	var bestItems []json.RawMessage
	if len(raw.Best) > 0 {
		if err := json.Unmarshal(raw.Best, &bestItems); err == nil {
			for _, item := range bestItems {
				if result := parseRecognitionResult(algorithm, item); result != nil {
					results.Best = append(results.Best, result)
				}
			}
		}
	}

	// Parse filtered results
	var filteredItems []json.RawMessage
	if len(raw.Filtered) > 0 {
		if err := json.Unmarshal(raw.Filtered, &filteredItems); err == nil {
			for _, item := range filteredItems {
				if result := parseRecognitionResult(algorithm, item); result != nil {
					results.Filtered = append(results.Filtered, result)
				}
			}
		}
	}

	return results, nil
}

// combinedResultItem represents a single item in the CombinedResult array for And/Or recognition.
type combinedResultItem struct {
	Algorithm string          `json:"algorithm"`
	Box       Rect            `json:"box"`
	Detail    json.RawMessage `json:"detail"`
	Name      string          `json:"name"`
	RecoID    int64           `json:"reco_id"`
}

// isCombinedRecognition returns true if the algorithm is a combined recognition (And or Or).
func isCombinedRecognition(algorithm string) bool {
	switch NodeRecognitionType(algorithm) {
	case NodeRecognitionTypeAnd, NodeRecognitionTypeOr:
		return true
	default:
		return false
	}
}

// parseCombinedResult parses detailJson for And/Or recognition and returns an array of RecognitionDetail.
// Detail JSON format: [{"algorithm": "...", "box": [...], "detail": {...}, "name": "...", "reco_id": ...}, ...]
func parseCombinedResult(detailJson string) ([]*RecognitionDetail, error) {
	if detailJson == "" || detailJson == "[]" {
		return []*RecognitionDetail{}, nil
	}

	var items []combinedResultItem
	if err := json.Unmarshal([]byte(detailJson), &items); err != nil {
		return []*RecognitionDetail{}, err
	}

	combinedResult := make([]*RecognitionDetail, 0, len(items))
	for _, item := range items {
		var err error
		var detail *RecognitionResults
		var combined []*RecognitionDetail

		if len(item.Detail) > 0 {
			if isCombinedRecognition(item.Algorithm) {
				combined, err = parseCombinedResult(string(item.Detail))
				if err != nil {
					return []*RecognitionDetail{}, err
				}
			} else {
				detail, err = parseRecognitionResults(item.Algorithm, string(item.Detail))
				if err != nil {
					return []*RecognitionDetail{}, err
				}
			}
		}

		combinedResult = append(combinedResult, &RecognitionDetail{
			ID:             item.RecoID,
			Name:           item.Name,
			Algorithm:      item.Algorithm,
			Box:            item.Box,
			DetailJson:     string(item.Detail),
			Results:        detail,
			CombinedResult: combined,
		})
	}

	return combinedResult, nil
}
