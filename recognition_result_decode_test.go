package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// These tests pin the decode-level contract of the recognition result union
// type with synthetic JSON, independently of the native round-trip in
// recognition_result_test.go.

func TestParseRecognitionResults_DecodeBranches(t *testing.T) {
	t.Run("DirectHitReturnsNil", func(t *testing.T) {
		results, err := parseRecognitionResults(string(RecognitionTypeDirectHit), []byte(`{"all":[]}`))
		require.NoError(t, err)
		require.Nil(t, results)
	})

	t.Run("EmptyDetailVariantsYieldEmptyNonNilContainers", func(t *testing.T) {
		for _, detail := range []string{"", "   ", "{}", "null"} {
			results, err := parseRecognitionResults(string(RecognitionTypeOCR), []byte(detail))
			require.NoError(t, err)
			require.NotNil(t, results)
			require.NotNil(t, results.All)
			require.Empty(t, results.All)
			require.NotNil(t, results.Filtered)
			require.Empty(t, results.Filtered)
			require.Nil(t, results.Best)
		}
	})

	t.Run("BestNullAndMissing", func(t *testing.T) {
		results, err := parseRecognitionResults(string(RecognitionTypeOCR),
			[]byte(`{"all":[{"box":[1,2,3,4],"text":"a","score":0.5}],"best":null,"filtered":[]}`))
		require.NoError(t, err)
		require.Nil(t, results.Best)
		require.Len(t, results.All, 1)

		results, err = parseRecognitionResults(string(RecognitionTypeOCR), []byte(`{"all":[],"filtered":[]}`))
		require.NoError(t, err)
		require.Nil(t, results.Best)
	})

	t.Run("AllAsSingleObject", func(t *testing.T) {
		results, err := parseRecognitionResults(string(RecognitionTypeOCR),
			[]byte(`{"all":{"box":[1,2,3,4],"text":"a","score":0.5},"best":null,"filtered":[]}`))
		require.NoError(t, err)
		require.Len(t, results.All, 1)
		val, ok := results.All[0].AsOCR()
		require.True(t, ok)
		require.Equal(t, Rect{1, 2, 3, 4}, val.Box)
		require.Equal(t, "a", val.Text)
		require.Equal(t, 0.5, val.Score)
	})

	t.Run("BestNonObjectIsSkipped", func(t *testing.T) {
		for _, best := range []string{"42", `"a"`, "[1]"} {
			results, err := parseRecognitionResults(string(RecognitionTypeOCR),
				[]byte(`{"all":[],"best":`+best+`,"filtered":[]}`))
			require.NoError(t, err)
			require.Nil(t, results.Best)
		}
	})

	t.Run("AllScalarYieldsEmptyWithoutError", func(t *testing.T) {
		for _, all := range []string{"42", `"a"`, "true"} {
			results, err := parseRecognitionResults(string(RecognitionTypeOCR),
				[]byte(`{"all":`+all+`,"best":null,"filtered":[]}`))
			require.NoError(t, err)
			require.NotNil(t, results)
			require.Empty(t, results.All)
		}
	})

	t.Run("UnknownAlgorithmErrors", func(t *testing.T) {
		results, err := parseRecognitionResults("Future", []byte(`{"all":[{}]}`))
		require.Error(t, err)
		require.Nil(t, results)
	})
}

func TestParseCustomRecognitionResult_DetailForms(t *testing.T) {
	t.Run("StringDetail", func(t *testing.T) {
		result, err := parseRecognitionResult(string(RecognitionTypeCustom),
			[]byte(`{"box":[1,2,3,4],"detail":"hello"}`))
		require.NoError(t, err)
		val, ok := result.AsCustom()
		require.True(t, ok)
		require.Equal(t, Rect{1, 2, 3, 4}, val.Box)
		require.Equal(t, "hello", val.Detail)
	})

	t.Run("ObjectDetailPassesThroughVerbatim", func(t *testing.T) {
		result, err := parseRecognitionResult(string(RecognitionTypeCustom),
			[]byte(`{"box":[1,2,3,4],"detail":{"k": 1,"v":[true,null]}}`))
		require.NoError(t, err)
		val, ok := result.AsCustom()
		require.True(t, ok)
		require.Equal(t, `{"k": 1,"v":[true,null]}`, val.Detail)
	})

	t.Run("NumberAndArrayDetailPassThrough", func(t *testing.T) {
		result, err := parseRecognitionResult(string(RecognitionTypeCustom),
			[]byte(`{"box":[0,0,0,0],"detail":42}`))
		require.NoError(t, err)
		val, ok := result.AsCustom()
		require.True(t, ok)
		require.Equal(t, "42", val.Detail)

		result, err = parseRecognitionResult(string(RecognitionTypeCustom),
			[]byte(`{"box":[0,0,0,0],"detail":[1,2]}`))
		require.NoError(t, err)
		val, ok = result.AsCustom()
		require.True(t, ok)
		require.Equal(t, "[1,2]", val.Detail)
	})

	t.Run("MissingDetailIsEmptyString", func(t *testing.T) {
		result, err := parseRecognitionResult(string(RecognitionTypeCustom),
			[]byte(`{"box":[1,2,3,4]}`))
		require.NoError(t, err)
		val, ok := result.AsCustom()
		require.True(t, ok)
		require.Equal(t, "", val.Detail)
	})

	t.Run("MalformedBoxErrors", func(t *testing.T) {
		result, err := parseRecognitionResult(string(RecognitionTypeCustom),
			[]byte(`{"box":"bad"}`))
		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("ThroughParseRecognitionResults", func(t *testing.T) {
		results, err := parseRecognitionResults(string(RecognitionTypeCustom),
			[]byte(`{"all":[{"box":[1,2,3,4],"detail":"x"}],"best":null,"filtered":[]}`))
		require.NoError(t, err)
		require.Len(t, results.All, 1)
		val, ok := results.All[0].AsCustom()
		require.True(t, ok)
		require.Equal(t, "x", val.Detail)
	})
}

func TestParseRecognitionResults_FailureLeavesNoPartialResult(t *testing.T) {
	t.Run("MalformedItemInAll", func(t *testing.T) {
		results, err := parseRecognitionResults(string(RecognitionTypeOCR),
			[]byte(`{"all":[{"box":[1,2,3,4],"text":"a","score":"high"}],"best":null,"filtered":[]}`))
		require.Error(t, err)
		require.Nil(t, results)
	})

	t.Run("MalformedBest", func(t *testing.T) {
		results, err := parseRecognitionResults(string(RecognitionTypeOCR),
			[]byte(`{"all":[],"best":{"score":"high"},"filtered":[]}`))
		require.Error(t, err)
		require.Nil(t, results)
	})

	t.Run("MalformedFiltered", func(t *testing.T) {
		results, err := parseRecognitionResults(string(RecognitionTypeOCR),
			[]byte(`{"all":[],"best":null,"filtered":[{"box":"bad"}]}`))
		require.Error(t, err)
		require.Nil(t, results)
	})

	t.Run("CombinedMalformedItemDetail", func(t *testing.T) {
		combined, err := parseCombinedResult(
			[]byte(`[{"algorithm":"OCR","reco_id":1,"detail":{"all":[{"score":"x"}]}}]`))
		require.Error(t, err)
		require.Empty(t, combined)
	})

	t.Run("CombinedUnknownAlgorithmWithDetail", func(t *testing.T) {
		combined, err := parseCombinedResult(
			[]byte(`[{"algorithm":"Future","reco_id":1,"detail":{"all":[{}]}}]`))
		require.Error(t, err)
		require.Empty(t, combined)
	})

	t.Run("CombinedNullDetailTolerated", func(t *testing.T) {
		combined, err := parseCombinedResult(
			[]byte(`[{"algorithm":"Future","reco_id":1,"name":"n","detail":null}]`))
		require.NoError(t, err)
		require.Len(t, combined, 1)
		require.Equal(t, "Future", combined[0].Algorithm)
		require.Nil(t, combined[0].Results)
		require.Nil(t, combined[0].CombinedResult)
	})
}
