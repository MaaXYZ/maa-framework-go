//go:build !customenv

package buffer

/*
#cgo !windows pkg-config: maa
#cgo windows CFLAGS: -IC:/maa/include
#cgo windows LDFLAGS: -LC:/maa/bin -lMaaFramework
*/
import "C"