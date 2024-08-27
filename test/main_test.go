package test

import (
	"github.com/MaaXYZ/maa-framework-go"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	loggingDir := "./debug"
	maa.SetLogDir(loggingDir)
	maa.SetSaveDraw(true)
	maa.SetStdoutLevel(maa.LoggingLevelInfo)

	os.Exit(m.Run())
}
