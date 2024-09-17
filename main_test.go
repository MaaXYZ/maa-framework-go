package maa

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	loggingDir := "./test/debug"
	SetLogDir(loggingDir)
	SetSaveDraw(true)
	SetStdoutLevel(LoggingLevelInfo)

	os.Exit(m.Run())
}
