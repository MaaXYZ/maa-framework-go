package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	got := Version()
	require.NotEmpty(t, got)
	t.Log(got)
}
