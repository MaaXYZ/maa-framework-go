package maa

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
)

type DbgController struct {
	path       string
	images     []image.Image
	imageIndex int
	resolution image.Point
}

func NewDbgController(path string) *Controller {
	return NewCustomController(&DbgController{
		path: path,
	})
}

// Click implements CustomController.
func (d *DbgController) Click(x int32, y int32) bool {
	return true
}

// ClickKey implements CustomController.
func (d *DbgController) ClickKey(keycode int32) bool {
	return true
}

// Connect implements CustomController.
func (d *DbgController) Connect() bool {

	if d.path == "" {
		return false
	}

	info, err := os.Stat(d.path)
	if err != nil {
		return false
	}

	// reset any previous cache
	d.images = nil
	d.imageIndex = 0

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
		d.images = append(d.images, img)
	}

	if info.IsDir() {
		// walk directory recursively and try decode files
		_ = filepath.Walk(d.path, func(p string, fi os.FileInfo, err error) error {
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
		tryDecode(d.path)
	}

	if len(d.images) == 0 {
		return false
	}

	d.resolution = d.images[0].Bounds().Size()
	return true
}

// GetFeature implements CustomController.
func (d *DbgController) GetFeature() ControllerFeature {
	return ControllerFeatureNone
}

// InputText implements CustomController.
func (d *DbgController) InputText(text string) bool {
	return true
}

// KeyDown implements CustomController.
func (d *DbgController) KeyDown(keycode int32) bool {
	return true
}

// KeyUp implements CustomController.
func (d *DbgController) KeyUp(keycode int32) bool {
	return true
}

// RequestUUID implements CustomController.
func (d *DbgController) RequestUUID() (string, bool) {
	return d.path, true
}

// Screencap implements CustomController.
func (d *DbgController) Screencap() (image.Image, bool) {
	if len(d.images) == 0 {
		return nil, false
	}

	if d.imageIndex >= len(d.images) {
		d.imageIndex = 0
	}

	img := d.images[d.imageIndex]
	d.imageIndex++
	return img, true
}

// StartApp implements CustomController.
func (d *DbgController) StartApp(intent string) bool {
	return true
}

// StopApp implements CustomController.
func (d *DbgController) StopApp(intent string) bool {
	return true
}

// Swipe implements CustomController.
func (d *DbgController) Swipe(x1 int32, y1 int32, x2 int32, y2 int32, duration int32) bool {
	return true
}

// TouchDown implements CustomController.
func (d *DbgController) TouchDown(contact int32, x int32, y int32, pressure int32) bool {
	return true
}

// TouchMove implements CustomController.
func (d *DbgController) TouchMove(contact int32, x int32, y int32, pressure int32) bool {
	return true
}

// TouchUp implements CustomController.
func (d *DbgController) TouchUp(contact int32) bool {
	return true
}

var _ CustomController = (*DbgController)(nil)
