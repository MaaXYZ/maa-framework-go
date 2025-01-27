package test

import (
	"os"
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v2"
)

func TestMain(m *testing.M) {
	loggingDir := "./debug"
	maa.SetLogDir(loggingDir)
	maa.SetSaveDraw(true)
	maa.SetStdoutLevel(maa.LoggingLevelInfo)

	os.Exit(m.Run())
}
