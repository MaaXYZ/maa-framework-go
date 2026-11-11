package maa

import (
	"errors"
	"strings"

	"github.com/MaaXYZ/maa-framework-go/v4/controller/ios"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
)

// NewIOSController creates an unconnected iOS controller backed by a running
// XDeviceBridge server. udid must be nonempty and contain no NUL. Call
// PostConnect and wait for its result before using the controller.
//
// config is a JSON object string; an empty string selects defaults. Supported
// fields include server_port, wifi_address, wifi_port, shared_control,
// max_bitrate, max_fps, and hardware_decode. XDeviceBridge owns pairing and DDI
// preparation. Stream prefers video frames; ScreenshotService permits lossless
// screenshot fallback. This controller works independently of the host OS.
func NewIOSController(udid string, screencap ios.ScreencapMethod, connection ios.ConnectionMethod, config string) (*Controller, error) {
	if udid == "" || strings.IndexByte(udid, 0) >= 0 {
		return nil, errors.New("iOS device UDID must be nonempty and contain no NUL")
	}
	if strings.IndexByte(config, 0) >= 0 {
		return nil, errors.New("iOS controller configuration must contain no NUL")
	}
	handle := native.MaaIOSControllerCreate(udid, uint64(screencap), int32(connection), config)
	if handle == 0 {
		return nil, errors.New("failed to create iOS controller")
	}
	initControllerStore(handle)
	return newOwnedController(handle), nil
}
