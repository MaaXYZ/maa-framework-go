package test

import (
	"os"
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v3"
)

func TestMain(m *testing.M) {
	maa.Init(
		maa.WithLogDir("./debug"),
		maa.WithSaveDraw(true),
		maa.WithStdoutLevel(maa.LoggingLevelInfo),
	)

	os.Exit(m.Run())
}
