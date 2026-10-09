package maa

import (
	"testing"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

// resourceOptionCall records one native MaaResourceSetOption invocation:
// the option key, the int32 value passed through the value pointer, and the
// declared value size in bytes.
type resourceOptionCall struct {
	key     native.MaaResOption
	value   int32
	valSize uint64
}

// stubResourceSetOption replaces native.MaaResourceSetOption with a recorder
// returning ok, appending every call to the slice pointed to by calls.
func stubResourceSetOption(t *testing.T, calls *[]resourceOptionCall, ok bool) {
	t.Helper()
	replaceNativeForTest(t, &native.MaaResourceSetOption, func(_ uintptr, key native.MaaResOption, value unsafe.Pointer, valSize uint64) bool {
		var recorded int32
		if value != nil && valSize == uint64(unsafe.Sizeof(recorded)) {
			recorded = *(*int32)(value)
		}
		*calls = append(*calls, resourceOptionCall{key: key, value: recorded, valSize: valSize})
		return ok
	})
}

// inferenceOptionValSize is the ABI value size every inference option must
// report: both MaaInferenceDevice and MaaInferenceExecutionProvider are
// int32-sized in MaaDef.h.
var inferenceOptionValSize = uint64(unsafe.Sizeof(native.MaaInferenceDevice_Auto))

func TestResource_UseExecutionProviderOptionSequence(t *testing.T) {
	testCases := []struct {
		name string
		use  func(*Resource) error
		want []resourceOptionCall
	}{
		{
			name: "UseCPU",
			use:  func(r *Resource) error { return r.UseCPU() },
			want: []resourceOptionCall{
				{
					key:     native.MaaResOption_InferenceExecutionProvider,
					value:   int32(native.MaaInferenceExecutionProvider_CPU),
					valSize: inferenceOptionValSize,
				},
				{
					key:     native.MaaResOption_InferenceDevice,
					value:   int32(native.MaaInferenceDevice_CPU),
					valSize: inferenceOptionValSize,
				},
			},
		},
		{
			name: "UseDirectml",
			use:  func(r *Resource) error { return r.UseDirectml(InferenceDevice0) },
			want: []resourceOptionCall{
				{
					key:     native.MaaResOption_InferenceExecutionProvider,
					value:   int32(native.MaaInferenceExecutionProvider_DirectML),
					valSize: inferenceOptionValSize,
				},
				{
					key:     native.MaaResOption_InferenceDevice,
					value:   int32(InferenceDevice0),
					valSize: inferenceOptionValSize,
				},
			},
		},
		{
			name: "UseCoreml",
			use:  func(r *Resource) error { return r.UseCoreml(InferenceDevice1) },
			want: []resourceOptionCall{
				{
					key:     native.MaaResOption_InferenceExecutionProvider,
					value:   int32(native.MaaInferenceExecutionProvider_CoreML),
					valSize: inferenceOptionValSize,
				},
				{
					key:     native.MaaResOption_InferenceDevice,
					value:   int32(InferenceDevice1),
					valSize: inferenceOptionValSize,
				},
			},
		},
		{
			name: "UseAutoExecutionProvider",
			use:  func(r *Resource) error { return r.UseAutoExecutionProvider() },
			want: []resourceOptionCall{
				{
					key:     native.MaaResOption_InferenceExecutionProvider,
					value:   int32(native.MaaInferenceExecutionProvider_Auto),
					valSize: inferenceOptionValSize,
				},
				{
					key:     native.MaaResOption_InferenceDevice,
					value:   int32(native.MaaInferenceDevice_Auto),
					valSize: inferenceOptionValSize,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := createResource(t)
			defer res.Destroy()

			var calls []resourceOptionCall
			stubResourceSetOption(t, &calls, true)

			require.NoError(t, tc.use(res))
			require.Equal(t, tc.want, calls)
		})
	}
}

// TestInferenceDeviceConstantsABI pins the inference device ABI to the values
// declared in deps/include/MaaFramework/MaaDef.h (MaaInferenceDeviceEnum) so
// accidental renumbering of the Go wrappers breaks the build here first.
func TestInferenceDeviceConstantsABI(t *testing.T) {
	require.EqualValues(t, -1, InferenceDeviceAuto)
	require.EqualValues(t, 0, InferenceDevice0)
	require.EqualValues(t, 1, InferenceDevice1)
	require.EqualValues(t, -2, native.MaaInferenceDevice_CPU)
	require.Equal(t, uint64(4), inferenceOptionValSize, "MaaInferenceDevice must stay int32-sized")
}

func TestResource_UseExecutionProviderNativeFailure(t *testing.T) {
	res := createResource(t)
	defer res.Destroy()

	var calls []resourceOptionCall
	stubResourceSetOption(t, &calls, false)

	require.Error(t, res.UseCPU())
	require.NotEmpty(t, calls)
}
