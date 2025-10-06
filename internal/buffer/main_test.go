package buffer

import (
	"os"
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
)

func TestMain(m *testing.M) {
	maa.Init("")

	os.Exit(m.Run())
}
