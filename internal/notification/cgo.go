//go:build !customenv

package notification

/*
#cgo !windows pkg-config: maa
#cgo windows CFLAGS: -IC:/maa/include
#cgo windows LDFLAGS: -LC:/maa/bin -lMaaFramework
*/
import "C"