//go:build !customenv

package maa

/*
#cgo !windows pkg-config: maa
#cgo windows CFLAGS: -IC:/maa/include
#cgo windows LDFLAGS: -LC:/maa/bin -lMaaFramework -lMaaToolkit
*/
import "C"
