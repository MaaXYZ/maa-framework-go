package ios

import (
	"testing"
	"unsafe"
)

func TestIOSMethodABI(t *testing.T) {
	if unsafe.Sizeof(ScreencapDefault) != 8 || unsafe.Sizeof(ConnectionAuto) != 4 {
		t.Fatal("iOS capture flags must be uint64 and connection selection int32")
	}
	if ScreencapNone != 0 || ScreencapStream != 1 || ScreencapScreenshotService != 2 || uint64(ScreencapAll) != ^uint64(0) || ScreencapDefault != ScreencapAll {
		t.Fatal("capture flags differ from the native definitions")
	}
	if ConnectionAuto != 0 || ConnectionUSB != 1 || ConnectionWiFi != 2 {
		t.Fatal("connection values differ from the native definitions")
	}
}
