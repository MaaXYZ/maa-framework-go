package maa

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubRecognitionItem_UnmarshalJSON_String(t *testing.T) {
	// Node name reference (C++ all_of/any_of element as string)
	data := []byte(`"SomeNodeName"`)
	var item SubRecognitionItem
	err := json.Unmarshal(data, &item)
	require.NoError(t, err)
	require.Equal(t, "SomeNodeName", item.NodeName)
	require.Nil(t, item.Inline)
}

func TestSubRecognitionItem_UnmarshalJSON_Object(t *testing.T) {
	// Inline recognition (C++ all_of/any_of element as object with type, param, sub_name)
	data := []byte(`{"sub_name":"MySub","type":"TemplateMatch","param":{"template":["a.png"],"threshold":[0.8]}}`)
	var item SubRecognitionItem
	err := json.Unmarshal(data, &item)
	require.NoError(t, err)
	require.Empty(t, item.NodeName)
	require.NotNil(t, item.Inline)
	require.Equal(t, "MySub", item.Inline.SubName)
	require.Equal(t, RecognitionTypeTemplateMatch, item.Inline.Type)
	require.IsType(t, (*TemplateMatchParam)(nil), item.Inline.Param)
}

func TestSubRecognitionItem_UnmarshalJSON_Invalid(t *testing.T) {
	data := []byte(`123`)
	var item SubRecognitionItem
	err := json.Unmarshal(data, &item)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected string or object")
}

func TestSubRecognitionItem_MarshalJSON(t *testing.T) {
	// Marshal node name ref
	ref := Ref("OtherNode")
	b, err := json.Marshal(ref)
	require.NoError(t, err)
	require.Equal(t, `"OtherNode"`, string(b))

	// Marshal inline
	inline := Inline(RecDirectHit(), "sub")
	b, err = json.Marshal(inline)
	require.NoError(t, err)
	var back SubRecognitionItem
	err = json.Unmarshal(b, &back)
	require.NoError(t, err)
	require.NotNil(t, back.Inline)
	require.Equal(t, "sub", back.Inline.SubName)
	require.Equal(t, RecognitionTypeDirectHit, back.Inline.Type)
}

func TestNodeAndRecognitionParam_Unmarshal_GetNodeDataStyle(t *testing.T) {
	// Simulates GetNodeData output: recognition.type + recognition.param { all_of, box_index }
	// all_of can be mix of string (node name) and object (inline)
	raw := `{
		"type": "And",
		"param": {
			"all_of": [
				"RefNodeA",
				{
					"sub_name": "InlineSub",
					"type": "DirectHit",
					"param": {}
				}
			],
			"box_index": 1
		}
	}`
	var reco Recognition
	err := json.Unmarshal([]byte(raw), &reco)
	require.NoError(t, err)
	require.Equal(t, RecognitionTypeAnd, reco.Type)
	andParam, ok := reco.Param.(*AndRecognitionParam)
	require.True(t, ok)
	require.Len(t, andParam.AllOf, 2)
	require.Equal(t, 1, andParam.BoxIndex)

	// First element: node name ref
	require.Equal(t, "RefNodeA", andParam.AllOf[0].NodeName)
	require.Nil(t, andParam.AllOf[0].Inline)

	// Second element: inline
	require.Empty(t, andParam.AllOf[1].NodeName)
	require.NotNil(t, andParam.AllOf[1].Inline)
	require.Equal(t, "InlineSub", andParam.AllOf[1].Inline.SubName)
	require.Equal(t, RecognitionTypeDirectHit, andParam.AllOf[1].Inline.Type)
}

func TestNodeOrRecognitionParam_Unmarshal_GetNodeDataStyle(t *testing.T) {
	// Simulates GetNodeData output for Or: any_of as string | object
	raw := `{
		"type": "Or",
		"param": {
			"any_of": [
				"RefNodeB",
				{
					"sub_name": "",
					"type": "ColorMatch",
					"param": {"lower":[[0,0,0]],"upper":[[255,255,255]],"count":1}
				}
			]
		}
	}`
	var reco Recognition
	err := json.Unmarshal([]byte(raw), &reco)
	require.NoError(t, err)
	require.Equal(t, RecognitionTypeOr, reco.Type)
	orParam, ok := reco.Param.(*OrRecognitionParam)
	require.True(t, ok)
	require.Len(t, orParam.AnyOf, 2)

	require.Equal(t, "RefNodeB", orParam.AnyOf[0].NodeName)
	require.Nil(t, orParam.AnyOf[0].Inline)

	require.NotNil(t, orParam.AnyOf[1].Inline)
	require.Equal(t, RecognitionTypeColorMatch, orParam.AnyOf[1].Inline.Type)
}

func TestRecAnd_RecOr_MarshalRoundtrip(t *testing.T) {
	// Build And with mix of ref and inline
	andRec := RecAnd(Ref("NodeRef"), Inline(RecDirectHit(), "sub1")).WithBoxIndex(2)
	b, err := json.Marshal(andRec)
	require.NoError(t, err)
	var reco Recognition
	err = json.Unmarshal(b, &reco)
	require.NoError(t, err)
	require.Equal(t, RecognitionTypeAnd, reco.Type)
	andParam := reco.Param.(*AndRecognitionParam)
	require.Len(t, andParam.AllOf, 2)
	require.Equal(t, 2, andParam.BoxIndex)
	require.Equal(t, "NodeRef", andParam.AllOf[0].NodeName)
	require.Equal(t, "sub1", andParam.AllOf[1].Inline.SubName)

	// Build Or with variadic Inline (no empty sub_name)
	orRec := RecOr(Inline(RecTemplateMatch(TemplateMatchParam{Template: []string{"t.png"}})))
	b, err = json.Marshal(orRec)
	require.NoError(t, err)
	err = json.Unmarshal(b, &reco)
	require.NoError(t, err)
	require.Equal(t, RecognitionTypeOr, reco.Type)
	orParam := reco.Param.(*OrRecognitionParam)
	require.Len(t, orParam.AnyOf, 1)
	require.NotNil(t, orParam.AnyOf[0].Inline)
	require.Equal(t, RecognitionTypeTemplateMatch, orParam.AnyOf[0].Inline.Type)
}

func TestNodeAndRecognitionParam_UnmarshalJSON(t *testing.T) {
	const nodeName = "CreditShoppingBuyFirst"

	const pipelineJSON = `
{
    "CreditShoppingBuyFirst": {
        "doc": "优先购买",
        "recognition": "And",
        "box_index": 3,
        "all_of": [
            {
                "sub_name": "CreditIcon",
                "recognition": "TemplateMatch",
                "template": "CreditShopping/CreditIcon.png",
                "roi": [
                    74,
                    103,
                    1131,
                    538
                ],
                "order_by": "vertical"
            },
            {
                "doc": "非售罄",
                "sub_name": "NotSoldOut",
                "recognition": "ColorMatch",
                "method": 6,
                "roi": "CreditIcon",
                "roi_offset": [
                    -95,
                    -145,
                    150,
                    200
                ],
                "lower": [
                    120
                ],
                "upper": [
                    255
                ],
                "count": 20000
            },
            {
                "doc": "按用户指定顺序购买",
                "sub_name": "BuyFirstOCR",
                "recognition": "OCR",
                "roi": "NotSoldOut",
                "roi_offset": [
                    0,
                    170,
                    0,
                    -160
                ],
                "expected": "^(嵌晶玉|武库配额)$",
                "order_by": "Expected"
            },
            {
                "doc": "买得起（价格不是红色）",
                "sub_name": "Affordable",
                "recognition": "ColorMatch",
                "roi": "BuyFirstOCR",
                "roi_offset": [
                    65,
                    -40,
                    -20,
                    1
                ],
                "lower": [
                    76,
                    76,
                    76
                ],
                "upper": [
                    108,
                    108,
                    108
                ],
                "count": 20,
                "order_by": "vertical"
            }
        ],
        "next": [
            "CreditShoppingBuyFirstItem"
        ]
    }
}
`

	resource := createResource(t)
	defer resource.Destroy()

	// Setup test pipeline
	err := resource.overridePipeline(pipelineJSON)
	require.NoError(t, err, "should successfully set pipeline")

	// Get node JSON
	nodeJSON, err := resource.GetNodeJSON(nodeName)
	require.NoError(t, err, "should successfully get node JSON")
	require.NotEmpty(t, nodeJSON, "node JSON should not be empty")

	// Unmarshal node
	var parsedNode Node
	err = json.Unmarshal([]byte(nodeJSON), &parsedNode)
	require.NoError(t, err, "should successfully unmarshal node")

	// Verify recognition configuration
	require.NotNil(t, parsedNode.Recognition, "recognition should not be nil")
	require.Equal(t, RecognitionTypeAnd, parsedNode.Recognition.Type, "recognition type should be And")

	// Verify And recognition parameters
	andParam, ok := parsedNode.Recognition.Param.(*AndRecognitionParam)
	require.True(t, ok, "param should be of type *NodeAndRecognitionParam")
	require.NotNil(t, andParam, "And recognition param should not be nil")
	require.Equal(t, 3, andParam.BoxIndex, "box_index should be 3")
	require.Len(t, andParam.AllOf, 4, "all_of should contain 4 sub-recognition items")

	// Verify first sub-recognition item (TemplateMatch) — inline object from GetNodeData
	templateMatchItem := andParam.AllOf[0].Inline
	require.NotNil(t, templateMatchItem, "first sub-item should not be nil")
	require.Equal(t, "CreditIcon", templateMatchItem.SubName, "first sub-item name should be CreditIcon")
	require.Equal(t, RecognitionTypeTemplateMatch, templateMatchItem.Type, "first sub-item type should be TemplateMatch")
	templateParam, ok := templateMatchItem.Param.(*TemplateMatchParam)
	require.True(t, ok, "first sub-item param should be of type *NodeTemplateMatchParam")
	require.NotEmpty(t, templateParam.Template, "template path should not be empty")
	require.Equal(t, TemplateMatchOrderByVertical, templateParam.OrderBy, "order_by should be Vertical")

	// Verify second sub-recognition item (ColorMatch)
	colorMatchItem := andParam.AllOf[1].Inline
	require.NotNil(t, colorMatchItem, "second sub-item should not be nil")
	require.Equal(t, "NotSoldOut", colorMatchItem.SubName, "second sub-item name should be NotSoldOut")
	require.Equal(t, RecognitionTypeColorMatch, colorMatchItem.Type, "second sub-item type should be ColorMatch")
	colorParam, ok := colorMatchItem.Param.(*ColorMatchParam)
	require.True(t, ok, "second sub-item param should be of type *NodeColorMatchParam")
	require.Equal(t, ColorMatchMethodGRAY, colorParam.Method, "color match method should be GRAY")
	require.Equal(t, 20000, colorParam.Count, "pixel count threshold should be 20000")

	// Verify third sub-recognition item (OCR)
	ocrItem := andParam.AllOf[2].Inline
	require.NotNil(t, ocrItem, "third sub-item should not be nil")
	require.Equal(t, "BuyFirstOCR", ocrItem.SubName, "third sub-item name should be BuyFirstOCR")
	require.Equal(t, RecognitionTypeOCR, ocrItem.Type, "third sub-item type should be OCR")
	ocrParam, ok := ocrItem.Param.(*OCRParam)
	require.True(t, ok, "third sub-item param should be of type *NodeOCRParam")
	require.NotEmpty(t, ocrParam.Expected, "OCR expected text should not be empty")

	// Verify fourth sub-recognition item (ColorMatch)
	affordableItem := andParam.AllOf[3].Inline
	require.NotNil(t, affordableItem, "fourth sub-item should not be nil")
	require.Equal(t, "Affordable", affordableItem.SubName, "fourth sub-item name should be Affordable")
	require.Equal(t, RecognitionTypeColorMatch, affordableItem.Type, "fourth sub-item type should be ColorMatch")
	affordableParam, ok := affordableItem.Param.(*ColorMatchParam)
	require.True(t, ok, "fourth sub-item param should be of type *NodeColorMatchParam")
	require.Equal(t, 20, affordableParam.Count, "pixel count threshold should be 20")
	require.Equal(t, ColorMatchOrderByVertical, affordableParam.OrderBy, "order_by should be Vertical")
}
