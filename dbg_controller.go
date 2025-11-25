package maa

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
)

type CarouselImageController struct {
	path       string
	images     []image.Image
	imageIndex int
	resolution image.Point
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
	return true
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

var _ CustomController = (*CarouselImageController)(nil)
