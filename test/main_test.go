package test

import (
	"os"
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4"
)

func TestMain(m *testing.M) {
	maa.Init(
		maa.WithLogDir("./debug"),
		maa.WithSaveDraw(true),
		maa.WithStdoutLevel(maa.LoggingLevelOff),
	)

	os.Exit(m.Run())
}
