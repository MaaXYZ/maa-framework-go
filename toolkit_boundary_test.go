package maa

import (
	"sync/atomic"
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

func TestToolkit_PortalHelperBoundary(t *testing.T) {
	var createCalls, destroyCalls, otherCalls atomic.Int64
	count := func() { otherCalls.Add(1) }

	// A fake create/destroy pair keeps the helper owned by the test while
	// every other native portal entry point is replaced by a call counter.
	replaceNativeForTest(t, &native.MaaToolkitPortalHelperCreate, func() uintptr {
		createCalls.Add(1)
		return 42
	})
	replaceNativeForTest(t, &native.MaaToolkitPortalHelperDestroy, func(uintptr) {
		destroyCalls.Add(1)
	})
	replaceNativeForTest(t, &native.MaaToolkitPortalHelperOpenStream, func(uintptr) bool {
		count()
		return true
	})
	replaceNativeForTest(t, &native.MaaToolkitPortalHelperGetPersist, func(uintptr) bool {
		count()
		return true
	})
	replaceNativeForTest(t, &native.MaaToolkitPortalHelperSetPersist, func(uintptr, bool) {
		count()
	})
	replaceNativeForTest(t, &native.MaaToolkitPortalHelperGetPipeWireFD, func(uintptr) int32 {
		count()
		return 17
	})
	replaceNativeForTest(t, &native.MaaToolkitPortalHelperGetPipeWireNodeID, func(uintptr) uint32 {
		count()
		return 23
	})
	replaceNativeForTest(t, &native.MaaToolkitPortalHelperGetRestoreToken, func(uintptr) string {
		count()
		return "token"
	})
	replaceNativeForTest(t, &native.MaaToolkitPortalHelperSetRestoreToken, func(uintptr, string) {
		count()
	})

	before := liveNativeObjects.Load()
	helper, err := NewPortalHelper()
	require.NoError(t, err)
	require.NotNil(t, helper)
	require.EqualValues(t, 1, createCalls.Load())
	require.Equal(t, before+1, liveNativeObjects.Load())

	helper.Destroy()
	require.EqualValues(t, 1, destroyCalls.Load())
	require.Equal(t, before, liveNativeObjects.Load())

	for _, tc := range []struct {
		name   string
		helper *PortalHelper
	}{
		{name: "destroyed", helper: helper},
		{name: "nil"},
		{name: "zero value", helper: &PortalHelper{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Error(t, tc.helper.OpenStream())
			require.False(t, tc.helper.Persist())
			require.Equal(t, -1, tc.helper.PipeWireFD())
			require.Zero(t, tc.helper.PipeWireNodeID())
			require.Empty(t, tc.helper.RestoreToken())
			tc.helper.SetPersist(true)
			tc.helper.SetRestoreToken("restored-token")
			tc.helper.Destroy()
			require.Zero(t, otherCalls.Load(), "closed helpers must not reach native code")
			require.EqualValues(t, 1, destroyCalls.Load(), "Destroy after closure must not destroy again")
		})
	}

	require.Zero(t, otherCalls.Load(), "closed helpers must not reach native code")
	require.EqualValues(t, 1, createCalls.Load())
	require.EqualValues(t, 1, destroyCalls.Load())
	require.Equal(t, before, liveNativeObjects.Load())
}
