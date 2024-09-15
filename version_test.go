package maa

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVersion(t *testing.T) {
	got := Version()
	require.NotEmpty(t, got)
	t.Log(got)
}
