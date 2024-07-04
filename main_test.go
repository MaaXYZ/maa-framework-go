package maa

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	loggingDir := "./debug"
	SetLogDir(loggingDir)
	SetSaveDraw(true)
	SetStdoutLevel(LoggingLevelInfo)

	os.Exit(m.Run())
}
