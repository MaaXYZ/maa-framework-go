package maa

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

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
	require.Equal(t, NodeRecognitionTypeAnd, parsedNode.Recognition.Type, "recognition type should be And")

	// Verify And recognition parameters
	andParam, ok := parsedNode.Recognition.Param.(*NodeAndRecognitionParam)
	require.True(t, ok, "param should be of type *NodeAndRecognitionParam")
	require.NotNil(t, andParam, "And recognition param should not be nil")
	require.Equal(t, 3, andParam.BoxIndex, "box_index should be 3")
	require.Len(t, andParam.AllOf, 4, "all_of should contain 4 sub-recognition items")

	// Verify first sub-recognition item (TemplateMatch)
	templateMatchItem := andParam.AllOf[0]
	require.NotNil(t, templateMatchItem, "first sub-item should not be nil")
	require.Equal(t, "CreditIcon", templateMatchItem.SubName, "first sub-item name should be CreditIcon")
	require.Equal(t, NodeRecognitionTypeTemplateMatch, templateMatchItem.Type, "first sub-item type should be TemplateMatch")
	templateParam, ok := templateMatchItem.Param.(*NodeTemplateMatchParam)
	require.True(t, ok, "first sub-item param should be of type *NodeTemplateMatchParam")
	require.NotEmpty(t, templateParam.Template, "template path should not be empty")
	require.Equal(t, NodeTemplateMatchOrderByVertical, templateParam.OrderBy, "order_by should be Vertical")

	// Verify second sub-recognition item (ColorMatch)
	colorMatchItem := andParam.AllOf[1]
	require.NotNil(t, colorMatchItem, "second sub-item should not be nil")
	require.Equal(t, "NotSoldOut", colorMatchItem.SubName, "second sub-item name should be NotSoldOut")
	require.Equal(t, NodeRecognitionTypeColorMatch, colorMatchItem.Type, "second sub-item type should be ColorMatch")
	colorParam, ok := colorMatchItem.Param.(*NodeColorMatchParam)
	require.True(t, ok, "second sub-item param should be of type *NodeColorMatchParam")
	require.Equal(t, NodeColorMatchMethodGRAY, colorParam.Method, "color match method should be GRAY")
	require.Equal(t, 20000, colorParam.Count, "pixel count threshold should be 20000")

	// Verify third sub-recognition item (OCR)
	ocrItem := andParam.AllOf[2]
	require.NotNil(t, ocrItem, "third sub-item should not be nil")
	require.Equal(t, "BuyFirstOCR", ocrItem.SubName, "third sub-item name should be BuyFirstOCR")
	require.Equal(t, NodeRecognitionTypeOCR, ocrItem.Type, "third sub-item type should be OCR")
	ocrParam, ok := ocrItem.Param.(*NodeOCRParam)
	require.True(t, ok, "third sub-item param should be of type *NodeOCRParam")
	require.NotEmpty(t, ocrParam.Expected, "OCR expected text should not be empty")

	// Verify fourth sub-recognition item (ColorMatch)
	affordableItem := andParam.AllOf[3]
	require.NotNil(t, affordableItem, "fourth sub-item should not be nil")
	require.Equal(t, "Affordable", affordableItem.SubName, "fourth sub-item name should be Affordable")
	require.Equal(t, NodeRecognitionTypeColorMatch, affordableItem.Type, "fourth sub-item type should be ColorMatch")
	affordableParam, ok := affordableItem.Param.(*NodeColorMatchParam)
	require.True(t, ok, "fourth sub-item param should be of type *NodeColorMatchParam")
	require.Equal(t, 20, affordableParam.Count, "pixel count threshold should be 20")
	require.Equal(t, NodeColorMatchOrderByVertical, affordableParam.OrderBy, "order_by should be Vertical")
}
