package maa

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sync/atomic"
)

type CarouselImageController struct {
	path       string
	images     []image.Image
	imageIndex int
	resolution image.Point
	connected  atomic.Bool
}

func NewCarouselImageController(path string) *Controller {
	return NewCustomController(&CarouselImageController{
		path: path,
	})
}

// Click implements CustomController.
func (c *CarouselImageController) Click(x int32, y int32) bool {
	return true
}

// ClickKey implements CustomController.
func (c *CarouselImageController) ClickKey(keycode int32) bool {
	return true
}

// Connect implements CustomController.
func (c *CarouselImageController) Connect() bool {
	// reset connection state first (reconnect-safe)
	c.connected.Store(false)

	if c.path == "" {
		return false
	}

	info, err := os.Stat(c.path)
	if err != nil {
		return false
	}

	// reset any previous cache
	c.images = nil
	c.imageIndex = 0

	// helper to try decode an image file and append on success
	tryDecode := func(p string) {
		f, err := os.Open(p)
		if err != nil {
			return
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			return
		}
		c.images = append(c.images, img)
	}

	if info.IsDir() {
		// walk directory recursively and try decode files
		_ = filepath.Walk(c.path, func(p string, fi os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if fi.IsDir() {
				return nil
			}
			tryDecode(p)
			return nil
		})
	} else {
		tryDecode(c.path)
	}

	if len(c.images) == 0 {
		return false
	}

	c.resolution = c.images[0].Bounds().Size()
	c.connected.Store(true)
	return true
}

// Connected implements CustomController.
func (c *CarouselImageController) Connected() bool {
	return c.connected.Load()
}

// GetFeature implements CustomController.
func (c *CarouselImageController) GetFeature() ControllerFeature {
	return ControllerFeatureNone
}

// InputText implements CustomController.
func (c *CarouselImageController) InputText(text string) bool {
	return true
}

// KeyDown implements CustomController.
func (c *CarouselImageController) KeyDown(keycode int32) bool {
	return true
}

// KeyUp implements CustomController.
func (c *CarouselImageController) KeyUp(keycode int32) bool {
	return true
}

// RequestUUID implements CustomController.
func (c *CarouselImageController) RequestUUID() (string, bool) {
	return c.path, true
}

// Screencap implements CustomController.
func (c *CarouselImageController) Screencap() (image.Image, bool) {
	if len(c.images) == 0 {
		return nil, false
	}

	if c.imageIndex >= len(c.images) {
		c.imageIndex = 0
	}

	img := c.images[c.imageIndex]
	c.imageIndex++
	return img, true
}

// StartApp implements CustomController.
func (c *CarouselImageController) StartApp(intent string) bool {
	return true
}

// StopApp implements CustomController.
func (c *CarouselImageController) StopApp(intent string) bool {
	return true
}

// Swipe implements CustomController.
func (c *CarouselImageController) Swipe(x1 int32, y1 int32, x2 int32, y2 int32, duration int32) bool {
	return true
}

// TouchDown implements CustomController.
func (c *CarouselImageController) TouchDown(contact int32, x int32, y int32, pressure int32) bool {
	return true
}

// TouchMove implements CustomController.
func (c *CarouselImageController) TouchMove(contact int32, x int32, y int32, pressure int32) bool {
	return true
}

// TouchUp implements CustomController.
func (c *CarouselImageController) TouchUp(contact int32) bool {
	return true
}

// Scroll implements CustomController.
func (c *CarouselImageController) Scroll(dx int32, dy int32) bool {
	return true
}

type BlankController struct{}

func NewBlankController() *Controller {
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

// Screencap implements CustomController.
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
