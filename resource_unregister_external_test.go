package maa

import (
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

func TestResource_UnregisterWithoutGoCallbackID(t *testing.T) {
	testCases := []struct {
		name               string
		registerExternal   func(uintptr, string) bool
		unregister         func(*Resource, string) error
		list               func(*Resource) ([]string, error)
		nativeList         *func(uintptr, uintptr) bool
		unregisterExternal func(uintptr, string) bool
	}{
		{
			name: "Recognition",
			registerExternal: func(handle uintptr, name string) bool {
				return native.MaaResourceRegisterCustomRecognition(handle, name, _MaaCustomRecognitionCallbackAgent, 0)
			},
			unregister:         (*Resource).UnregisterCustomRecognition,
			list:               (*Resource).GetCustomRecognitionList,
			nativeList:         &native.MaaResourceGetCustomRecognitionList,
			unregisterExternal: native.MaaResourceUnregisterCustomRecognition,
		},
		{
			name: "Action",
			registerExternal: func(handle uintptr, name string) bool {
				return native.MaaResourceRegisterCustomAction(handle, name, _MaaCustomActionCallbackAgent, 0)
			},
			unregister:         (*Resource).UnregisterCustomAction,
			list:               (*Resource).GetCustomActionList,
			nativeList:         &native.MaaResourceGetCustomActionList,
			unregisterExternal: native.MaaResourceUnregisterCustomAction,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := createResource(t)
			t.Cleanup(func() { require.NoError(t, res.Destroy()) })

			t.Run("UnknownName", func(t *testing.T) {
				require.NoError(t, tc.unregister(res, "MissingExternalCallback"))
			})

			t.Run("ListFailure", func(t *testing.T) {
				replaceNativeForTest(t, tc.nativeList, func(uintptr, uintptr) bool { return false })
				require.Error(t, tc.unregister(res, "UnconfirmedExternalCallback"))
			})

			t.Run("ExternalRegistration", func(t *testing.T) {
				name := "External" + tc.name
				// Register directly in native code, without a Go callback ID.
				// This test never invokes the callback.
				require.True(t, tc.registerExternal(res.handle, name))
				t.Cleanup(func() { require.True(t, tc.unregisterExternal(res.handle, name)) })
				names, err := tc.list(res)
				require.NoError(t, err)
				require.Contains(t, names, name)

				require.Error(t, tc.unregister(res, name))

				names, err = tc.list(res)
				require.NoError(t, err)
				require.Contains(t, names, name)
			})
		})
	}
}
