// This file provides BlankController, a pure-Go no-op custom controller
// implementation (via NewCustomController) for testing and development
// purposes.
//
// NOTE: This is NOT a binding to MaaDbgControllerCreate from the C API.
// MaaDbgControllerCreate is intentionally excluded from the Go binding.
// Do NOT add a NewDbgController or any wrapper for MaaDbgControllerCreate here.
// Use BlankController as a Go-native no-op stub alternative.
// For image-based testing in Go, use NewReplayController or a CustomController
// with a custom Screencap implementation; MaaDbgControllerCreate serves images
// from a directory and is available only through the C API.

package maa

import (
	"image"
)

// BlankController is a no-op CustomController for testing: every method
// reports success without performing any action. It is designed to be
// embedded in test controller types that override only the methods under
// test. Observable behavior worth knowing:
//   - Screencap returns a 1280x720 image; after the native round-trip it is
//     opaque black (alpha is dropped on the RGBA-to-BGR conversion and
//     forced back to 255 on decoding).
//   - RequestUUID returns "blank-controller" for every instance, so
//     controllers are not distinguishable by UUID.
//   - Connected reports true unconditionally, even before Connect.
//   - GetInfo returns {"type":"blank"}, but the framework overwrites the
//     "type" key with "custom" for every custom controller, so
//     Controller.GetInfo always reports type "custom".
type BlankController struct{}

var _ CustomController = (*BlankController)(nil)

// NewBlankController creates a controller whose operations all succeed as
// no-ops. Use it to test framework features (resource binding, tasker
// initialization, etc.) without any real controller behavior. It returns an
// error only when native controller creation fails.
func NewBlankController() (*Controller, error) {
	return NewCustomController(&BlankController{})
}

// Click implements CustomController.
func (c *BlankController) Click(x int32, y int32) bool {
	return true
}

// ClickKey implements CustomController.
func (c *BlankController) ClickKey(keycode int32) bool {
	return true
}

// Connect implements CustomController.
func (c *BlankController) Connect() bool {
	return true
}

// Connected implements CustomController.
func (c *BlankController) Connected() bool {
	return true
}

// GetFeature implements CustomController.
func (c *BlankController) GetFeature() ControllerFeature {
	return ControllerFeatureNone
}

// InputText implements CustomController.
func (c *BlankController) InputText(text string) bool {
	return true
}

// KeyDown implements CustomController.
func (c *BlankController) KeyDown(keycode int32) bool {
	return true
}

// KeyUp implements CustomController.
func (c *BlankController) KeyUp(keycode int32) bool {
	return true
}

// RequestUUID implements CustomController.
func (c *BlankController) RequestUUID() (string, bool) {
	return "blank-controller", true
}

// Screencap implements CustomController. It returns a blank 1280x720 image,
// opaque black after the native round-trip.
func (c *BlankController) Screencap() (image.Image, bool) {
	return image.NewRGBA(image.Rect(0, 0, 1280, 720)), true
}

// StartApp implements CustomController.
func (c *BlankController) StartApp(intent string) bool {
	return true
}

// StopApp implements CustomController.
func (c *BlankController) StopApp(intent string) bool {
	return true
}

// Swipe implements CustomController.
func (c *BlankController) Swipe(x1 int32, y1 int32, x2 int32, y2 int32, duration int32) bool {
	return true
}

// TouchDown implements CustomController.
func (c *BlankController) TouchDown(contact int32, x int32, y int32, pressure int32) bool {
	return true
}

// TouchMove implements CustomController.
func (c *BlankController) TouchMove(contact int32, x int32, y int32, pressure int32) bool {
	return true
}

// TouchUp implements CustomController.
func (c *BlankController) TouchUp(contact int32) bool {
	return true
}

// Scroll implements CustomController.
func (c *BlankController) Scroll(dx int32, dy int32) bool {
	return true
}

// RelativeMove implements CustomController.
func (c *BlankController) RelativeMove(dx int32, dy int32) bool {
	return true
}

// Shell implements CustomController.
func (c *BlankController) Shell(cmd string, timeout int64) (string, bool) {
	return "", true
}

// Inactive implements CustomController.
func (c *BlankController) Inactive() bool {
	return true
}

// GetInfo implements CustomController.
func (c *BlankController) GetInfo() (string, bool) {
	info := map[string]any{
		"type": "blank",
	}
	data, err := marshalJSON(info)
	if err != nil {
		return "", false
	}
	return string(data), true
}
