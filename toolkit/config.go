package toolkit

/*
#include <stdlib.h>
#include <MaaToolkit/MaaToolkitAPI.h>
*/
import "C"
import "unsafe"

// ConfigInitOption inits the toolkit config option.
func ConfigInitOption(userPath, defaultJson string) bool {
	cUserPath := C.CString(userPath)
	defer C.free(unsafe.Pointer(cUserPath))
	cDefaultJson := C.CString(defaultJson)
	defer C.free(unsafe.Pointer(cDefaultJson))

	return C.MaaToolkitConfigInitOption(cUserPath, cDefaultJson) != 0
}
