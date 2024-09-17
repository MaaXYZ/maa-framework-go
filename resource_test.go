package maa

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func createResource(t *testing.T) *Resource {
	res := NewResource(nil)
	require.NotNil(t, res)
	return res
}
