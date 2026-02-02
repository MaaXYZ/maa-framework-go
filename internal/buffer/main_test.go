package buffer

import (
	"os"
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
)

func TestMain(m *testing.M) {
	native.Init("")

	os.Exit(m.Run())
}
