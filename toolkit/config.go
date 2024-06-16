package toolkit

/*
#include <stdlib.h>
#include <MaaToolkit/MaaToolkitAPI.h>
*/
import "C"
import "unsafe"

// InitOption inits the toolkit option config.
func InitOption(userPath, defaultJson string) bool {
	cUserPath := C.CString(userPath)
	cDefaultJson := C.CString(defaultJson)
	defer func() {
		C.free(unsafe.Pointer(cUserPath))
		C.free(unsafe.Pointer(cDefaultJson))
	}()
	return C.MaaToolkitInitOptionConfig(cUserPath, cDefaultJson) != 0
}
