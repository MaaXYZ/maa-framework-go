package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"

// Version returns the version of the maa framework.
func Version() string {
	return C.GoString(C.MaaVersion())
}
