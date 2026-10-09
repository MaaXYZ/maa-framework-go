package maa

import (
	"testing"
	"time"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v4/controller/adb"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

func TestToolkit_ConfigInitOption(t *testing.T) {
	err := ConfigInitOption("./test", "{}")
	require.NoError(t, err)
}

func TestToolkit_ConfigInitOptionFailure(t *testing.T) {
	var gotUserPath, gotDefaultJSON string
	called := false
	replaceNativeForTest(t, &native.MaaToolkitConfigInitOption, func(userPath, defaultJson string) bool {
		called = true
		gotUserPath, gotDefaultJSON = userPath, defaultJson
		return false
	})

	err := ConfigInitOption("my-path", `{"logging":false}`)
	require.Error(t, err)
	require.True(t, called)
	require.Equal(t, "my-path", gotUserPath)
	require.Equal(t, `{"logging":false}`, gotDefaultJSON)
}

func TestToolkit_FindAdbDevicesStub(t *testing.T) {
	const listHandle = uintptr(7)

	replaceNativeForTest(t, &native.MaaToolkitAdbDeviceListSize, func(list uintptr) uint64 {
		require.Equal(t, listHandle, list)
		return 1
	})
	replaceNativeForTest(t, &native.MaaToolkitAdbDeviceListAt, func(list uintptr, index uint64) uintptr {
		require.Equal(t, listHandle, list)
		require.Equal(t, uint64(0), index)
		return 100
	})
	replaceNativeForTest(t, &native.MaaToolkitAdbDeviceGetName, func(uintptr) string { return "serial-emulator" })
	replaceNativeForTest(t, &native.MaaToolkitAdbDeviceGetAdbPath, func(uintptr) string { return "/usr/bin/adb" })
	replaceNativeForTest(t, &native.MaaToolkitAdbDeviceGetAddress, func(uintptr) string { return "127.0.0.1:5555" })
	replaceNativeForTest(t, &native.MaaToolkitAdbDeviceGetScreencapMethods, func(uintptr) uint64 { return 3 })
	replaceNativeForTest(t, &native.MaaToolkitAdbDeviceGetInputMethods, func(uintptr) uint64 { return 4 })
	replaceNativeForTest(t, &native.MaaToolkitAdbDeviceGetConfig, func(uintptr) string { return `{"extras":{}}` })

	wantDevice := []*AdbDevice{{
		Name:            "serial-emulator",
		AdbPath:         "/usr/bin/adb",
		Address:         "127.0.0.1:5555",
		ScreencapMethod: adb.ScreencapMethod(3),
		InputMethod:     adb.InputMethod(4),
		Config:          `{"extras":{}}`,
	}}

	// Each subtest installs its own destroy stub so it can run in isolation.
	destroyStub := func(t *testing.T) *int {
		destroyed := 0
		replaceNativeForTest(t, &native.MaaToolkitAdbDeviceListDestroy, func(handle uintptr) {
			require.Equal(t, listHandle, handle)
			destroyed++
		})
		return &destroyed
	}

	t.Run("auto discovery uses MaaToolkitAdbDeviceFind", func(t *testing.T) {
		destroyed := destroyStub(t)
		replaceNativeForTest(t, &native.MaaToolkitAdbDeviceListCreate, func() uintptr { return listHandle })
		replaceNativeForTest(t, &native.MaaToolkitAdbDeviceFind, func(buffer uintptr) bool {
			require.Equal(t, listHandle, buffer)
			return true
		})
		replaceNativeForTest(t, &native.MaaToolkitAdbDeviceFindSpecified, func(string, uintptr) bool {
			t.Error("FindSpecified must not be called without a specified adb path")
			return false
		})

		devices, err := FindAdbDevices()
		require.NoError(t, err)
		require.Equal(t, wantDevice, devices)
		require.Equal(t, 1, *destroyed)
	})

	t.Run("specified discovery forwards only the first adb path", func(t *testing.T) {
		destroyed := destroyStub(t)
		replaceNativeForTest(t, &native.MaaToolkitAdbDeviceListCreate, func() uintptr { return listHandle })
		replaceNativeForTest(t, &native.MaaToolkitAdbDeviceFind, func(uintptr) bool {
			t.Error("Find must not be called when a specified adb path is given")
			return false
		})
		var gotPath string
		replaceNativeForTest(t, &native.MaaToolkitAdbDeviceFindSpecified, func(adbPath string, buffer uintptr) bool {
			gotPath = adbPath
			require.Equal(t, listHandle, buffer)
			return true
		})

		devices, err := FindAdbDevices("first-adb", "second-adb")
		require.NoError(t, err)
		require.Equal(t, "first-adb", gotPath)
		require.Equal(t, wantDevice, devices)
		require.Equal(t, 1, *destroyed)
	})

	t.Run("empty result is an empty slice", func(t *testing.T) {
		destroyed := destroyStub(t)
		replaceNativeForTest(t, &native.MaaToolkitAdbDeviceListCreate, func() uintptr { return listHandle })
		replaceNativeForTest(t, &native.MaaToolkitAdbDeviceFind, func(uintptr) bool { return true })
		replaceNativeForTest(t, &native.MaaToolkitAdbDeviceListSize, func(uintptr) uint64 { return 0 })

		devices, err := FindAdbDevices()
		require.NoError(t, err)
		require.NotNil(t, devices)
		require.Empty(t, devices)
		require.Equal(t, 1, *destroyed)
	})

	t.Run("native failure surfaces an error", func(t *testing.T) {
		destroyed := destroyStub(t)
		replaceNativeForTest(t, &native.MaaToolkitAdbDeviceListCreate, func() uintptr { return listHandle })
		replaceNativeForTest(t, &native.MaaToolkitAdbDeviceFind, func(uintptr) bool { return false })

		devices, err := FindAdbDevices()
		require.Error(t, err)
		require.Nil(t, devices)
		require.Equal(t, 1, *destroyed)
	})
}

func TestToolkit_FindDesktopWindowsStub(t *testing.T) {
	const listHandle = uintptr(9)
	var fakeHandle int

	replaceNativeForTest(t, &native.MaaToolkitDesktopWindowListCreate, func() uintptr { return listHandle })
	replaceNativeForTest(t, &native.MaaToolkitDesktopWindowListDestroy, func(handle uintptr) {
		require.Equal(t, listHandle, handle)
	})
	replaceNativeForTest(t, &native.MaaToolkitDesktopWindowGetHandle, func(uintptr) unsafe.Pointer {
		return unsafe.Pointer(&fakeHandle)
	})
	replaceNativeForTest(t, &native.MaaToolkitDesktopWindowGetClassName, func(uintptr) string { return "WindowClass" })
	replaceNativeForTest(t, &native.MaaToolkitDesktopWindowGetWindowName, func(uintptr) string { return "Window Title" })

	t.Run("field mapping", func(t *testing.T) {
		replaceNativeForTest(t, &native.MaaToolkitDesktopWindowFindAll, func(buffer uintptr) bool {
			require.Equal(t, listHandle, buffer)
			return true
		})
		replaceNativeForTest(t, &native.MaaToolkitDesktopWindowListSize, func(uintptr) uint64 { return 1 })
		replaceNativeForTest(t, &native.MaaToolkitDesktopWindowListAt, func(_ uintptr, index uint64) uintptr {
			require.Equal(t, uint64(0), index)
			return 77
		})

		windows, err := FindDesktopWindows()
		require.NoError(t, err)
		require.Equal(t, []*DesktopWindow{{
			Handle:     unsafe.Pointer(&fakeHandle),
			ClassName:  "WindowClass",
			WindowName: "Window Title",
		}}, windows)
	})

	t.Run("empty result is an empty slice", func(t *testing.T) {
		replaceNativeForTest(t, &native.MaaToolkitDesktopWindowFindAll, func(uintptr) bool { return true })
		replaceNativeForTest(t, &native.MaaToolkitDesktopWindowListSize, func(uintptr) uint64 { return 0 })

		windows, err := FindDesktopWindows()
		require.NoError(t, err)
		require.NotNil(t, windows)
		require.Empty(t, windows)
	})

	t.Run("native failure surfaces an error", func(t *testing.T) {
		replaceNativeForTest(t, &native.MaaToolkitDesktopWindowFindAll, func(uintptr) bool { return false })

		windows, err := FindDesktopWindows()
		require.Error(t, err)
		require.Nil(t, windows)
	})
}

func TestToolkit_FindGamescopeInstancesGuards(t *testing.T) {
	t.Run("list create failure", func(t *testing.T) {
		replaceNativeForTest(t, &native.MaaToolkitGamescopeInstanceListCreate, func() uintptr { return 0 })

		instances, err := FindGamescopeInstances()
		require.Error(t, err)
		require.Nil(t, instances)
	})

	t.Run("empty result is an empty slice", func(t *testing.T) {
		replaceNativeForTest(t, &native.MaaToolkitGamescopeInstanceListCreate, func() uintptr { return 7 })
		replaceNativeForTest(t, &native.MaaToolkitGamescopeInstanceListDestroy, func(uintptr) {})
		replaceNativeForTest(t, &native.MaaToolkitGamescopeInstanceFindAll, func(uintptr) bool { return true })
		replaceNativeForTest(t, &native.MaaToolkitGamescopeInstanceListSize, func(uintptr) uint64 { return 0 })

		instances, err := FindGamescopeInstances()
		require.NoError(t, err)
		require.NotNil(t, instances)
		require.Empty(t, instances)
	})

	t.Run("null instance entry", func(t *testing.T) {
		replaceNativeForTest(t, &native.MaaToolkitGamescopeInstanceListCreate, func() uintptr { return 7 })
		replaceNativeForTest(t, &native.MaaToolkitGamescopeInstanceListDestroy, func(uintptr) {})
		replaceNativeForTest(t, &native.MaaToolkitGamescopeInstanceFindAll, func(uintptr) bool { return true })
		replaceNativeForTest(t, &native.MaaToolkitGamescopeInstanceListSize, func(uintptr) uint64 { return 1 })
		replaceNativeForTest(t, &native.MaaToolkitGamescopeInstanceListAt, func(uintptr, uint64) uintptr { return 0 })

		instances, err := FindGamescopeInstances()
		require.Error(t, err)
		require.Nil(t, instances)
	})
}

func TestToolkit_MacOSPermissionWrappers(t *testing.T) {
	var checked, requested, revealed []MacOSPermission
	replaceNativeForTest(t, &native.MaaToolkitMacOSCheckPermission, func(perm native.MaaMacOSPermission) bool {
		checked = append(checked, MacOSPermission(perm))
		return perm == native.MaaMacOSPermissionScreenCapture
	})
	replaceNativeForTest(t, &native.MaaToolkitMacOSRequestPermission, func(perm native.MaaMacOSPermission) bool {
		requested = append(requested, MacOSPermission(perm))
		return false
	})
	replaceNativeForTest(t, &native.MaaToolkitMacOSRevealPermissionSettings, func(perm native.MaaMacOSPermission) bool {
		revealed = append(revealed, MacOSPermission(perm))
		return true
	})

	require.True(t, MacOSCheckPermission(MacOSPermissionScreenCapture))
	require.False(t, MacOSCheckPermission(MacOSPermissionAccessibility))
	require.False(t, MacOSRequestPermission(MacOSPermissionAccessibility))
	require.True(t, MacOSRevealPermissionSettings(MacOSPermission(99)))

	require.Equal(t, []MacOSPermission{MacOSPermissionScreenCapture, MacOSPermissionAccessibility}, checked)
	require.Equal(t, []MacOSPermission{MacOSPermissionAccessibility}, requested)
	require.Equal(t, []MacOSPermission{MacOSPermission(99)}, revealed)
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

func TestToolkit_PortalHelperConcurrentDestroy(t *testing.T) {
	create, destroy := native.MaaToolkitPortalHelperCreate, native.MaaToolkitPortalHelperDestroy
	t.Cleanup(func() {
		native.MaaToolkitPortalHelperCreate = create
		native.MaaToolkitPortalHelperDestroy = destroy
	})

	native.MaaToolkitPortalHelperCreate = func() uintptr { return 42 }
	entered := make(chan uintptr, 2)
	releaseDestroy := make(chan struct{})
	native.MaaToolkitPortalHelperDestroy = func(handle uintptr) {
		entered <- handle
		<-releaseDestroy
	}

	before := liveNativeObjects.Load()
	helper, err := NewPortalHelper()
	require.NoError(t, err)
	done := make(chan struct{}, 2)
	go func() {
		helper.Destroy()
		done <- struct{}{}
	}()
	require.Equal(t, uintptr(42), <-entered)

	secondStarted := make(chan struct{})
	go func() {
		close(secondStarted)
		helper.Destroy()
		done <- struct{}{}
	}()
	<-secondStarted
	select {
	case handle := <-entered:
		close(releaseDestroy)
		<-done
		<-done
		t.Fatalf("native destroy called twice for handle %d", handle)
	case <-time.After(25 * time.Millisecond):
	}
	close(releaseDestroy)
	<-done
	<-done
	require.Empty(t, entered)
	require.Equal(t, before, liveNativeObjects.Load())
}
