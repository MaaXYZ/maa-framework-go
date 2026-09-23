package maa

import (
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

func TestToolkit_ConfigInitOption(t *testing.T) {
	err := ConfigInitOption("./test", "{}")
	require.NoError(t, err)
}

func TestToolkit_FindAdbDevices(t *testing.T) {
	adbDevices, err := FindAdbDevices()
	require.NoError(t, err)
	require.NotNil(t, adbDevices)
}

func TestToolkit_FindDesktopWindows(t *testing.T) {
	desktopWindows, err := FindDesktopWindows()
	require.NoError(t, err)
	require.NotNil(t, desktopWindows)
}

func TestToolkit_FindGamescopeInstances(t *testing.T) {
	create := native.MaaToolkitGamescopeInstanceListCreate
	destroy := native.MaaToolkitGamescopeInstanceListDestroy
	findAll := native.MaaToolkitGamescopeInstanceFindAll
	size := native.MaaToolkitGamescopeInstanceListSize
	at := native.MaaToolkitGamescopeInstanceListAt
	displayNo := native.MaaToolkitGamescopeInstanceGetDisplayNo
	nodeID := native.MaaToolkitGamescopeInstanceGetPipeWireNodeId
	eisPath := native.MaaToolkitGamescopeInstanceGetEisSocketPath
	t.Cleanup(func() {
		native.MaaToolkitGamescopeInstanceListCreate = create
		native.MaaToolkitGamescopeInstanceListDestroy = destroy
		native.MaaToolkitGamescopeInstanceFindAll = findAll
		native.MaaToolkitGamescopeInstanceListSize = size
		native.MaaToolkitGamescopeInstanceListAt = at
		native.MaaToolkitGamescopeInstanceGetDisplayNo = displayNo
		native.MaaToolkitGamescopeInstanceGetPipeWireNodeId = nodeID
		native.MaaToolkitGamescopeInstanceGetEisSocketPath = eisPath
	})

	destroyed := 0
	native.MaaToolkitGamescopeInstanceListCreate = func() uintptr { return 1 }
	native.MaaToolkitGamescopeInstanceListDestroy = func(handle uintptr) {
		require.Equal(t, uintptr(1), handle)
		destroyed++
	}
	native.MaaToolkitGamescopeInstanceFindAll = func(buffer uintptr) bool { return buffer == 1 }
	native.MaaToolkitGamescopeInstanceListSize = func(list uintptr) uint64 { return 2 }
	native.MaaToolkitGamescopeInstanceListAt = func(list uintptr, index uint64) uintptr { return uintptr(index + 10) }
	native.MaaToolkitGamescopeInstanceGetDisplayNo = func(instance uintptr) uint32 { return uint32(instance - 10) }
	native.MaaToolkitGamescopeInstanceGetPipeWireNodeId = func(instance uintptr) uint32 { return uint32(instance + 100) }
	native.MaaToolkitGamescopeInstanceGetEisSocketPath = func(instance uintptr) string { return "socket" }

	instances, err := FindGamescopeInstances()
	require.NoError(t, err)
	require.Equal(t, []*GamescopeInstance{
		{DisplayNo: 0, PipeWireNodeID: 110, EisSocketPath: "socket"},
		{DisplayNo: 1, PipeWireNodeID: 111, EisSocketPath: "socket"},
	}, instances)
	require.Equal(t, 1, destroyed)

	native.MaaToolkitGamescopeInstanceFindAll = func(uintptr) bool { return false }
	instances, err = FindGamescopeInstances()
	require.Error(t, err)
	require.Nil(t, instances)
	require.Equal(t, 2, destroyed)
}

func TestToolkit_PortalHelper(t *testing.T) {
	create := native.MaaToolkitPortalHelperCreate
	destroy := native.MaaToolkitPortalHelperDestroy
	openStream := native.MaaToolkitPortalHelperOpenStream
	getPersist := native.MaaToolkitPortalHelperGetPersist
	setPersist := native.MaaToolkitPortalHelperSetPersist
	getFD := native.MaaToolkitPortalHelperGetPipeWireFD
	getNodeID := native.MaaToolkitPortalHelperGetPipeWireNodeID
	getToken := native.MaaToolkitPortalHelperGetRestoreToken
	setToken := native.MaaToolkitPortalHelperSetRestoreToken
	t.Cleanup(func() {
		native.MaaToolkitPortalHelperCreate = create
		native.MaaToolkitPortalHelperDestroy = destroy
		native.MaaToolkitPortalHelperOpenStream = openStream
		native.MaaToolkitPortalHelperGetPersist = getPersist
		native.MaaToolkitPortalHelperSetPersist = setPersist
		native.MaaToolkitPortalHelperGetPipeWireFD = getFD
		native.MaaToolkitPortalHelperGetPipeWireNodeID = getNodeID
		native.MaaToolkitPortalHelperGetRestoreToken = getToken
		native.MaaToolkitPortalHelperSetRestoreToken = setToken
	})

	destroyed := 0
	persist := false
	token := ""
	native.MaaToolkitPortalHelperCreate = func() uintptr { return 42 }
	native.MaaToolkitPortalHelperDestroy = func(handle uintptr) {
		require.Equal(t, uintptr(42), handle)
		destroyed++
	}
	native.MaaToolkitPortalHelperOpenStream = func(handle uintptr) bool { return handle == 42 }
	native.MaaToolkitPortalHelperGetPersist = func(uintptr) bool { return persist }
	native.MaaToolkitPortalHelperSetPersist = func(_ uintptr, enable bool) { persist = enable }
	native.MaaToolkitPortalHelperGetPipeWireFD = func(uintptr) int32 { return 17 }
	native.MaaToolkitPortalHelperGetPipeWireNodeID = func(uintptr) uint32 { return 23 }
	native.MaaToolkitPortalHelperGetRestoreToken = func(uintptr) string { return token }
	native.MaaToolkitPortalHelperSetRestoreToken = func(_ uintptr, value string) { token = value }

	helper, err := NewPortalHelper()
	require.NoError(t, err)
	helper.SetPersist(true)
	require.True(t, helper.Persist())
	helper.SetRestoreToken("saved-token")
	require.Equal(t, "saved-token", helper.RestoreToken())
	require.NoError(t, helper.OpenStream())
	require.Equal(t, 17, helper.PipeWireFD())
	require.Equal(t, uint32(23), helper.PipeWireNodeID())
	helper.Destroy()
	helper.Destroy()
	require.Equal(t, 1, destroyed)
	require.Error(t, helper.OpenStream())

	native.MaaToolkitPortalHelperCreate = func() uintptr { return 0 }
	helper, err = NewPortalHelper()
	require.Error(t, err)
	require.Nil(t, helper)
}
