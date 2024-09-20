package maa

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func createResource(t *testing.T, notify Notification) *Resource {
	res := NewResource(notify)
	require.NotNil(t, res)
	return res
}
