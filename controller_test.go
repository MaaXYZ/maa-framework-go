package maa

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func createDbgController(t *testing.T) Controller {
	testingPath := "./test/data_set/PipelineSmoking/Screenshot"
	resultPath := "./test/data_set/debug"

	ctrl := NewDbgController(testingPath, resultPath, DbgControllerTypeCarouselImage, "{}", nil)
	require.NotNil(t, ctrl)
	return ctrl
}
