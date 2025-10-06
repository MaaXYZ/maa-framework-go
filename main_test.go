package maa

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	Init(
		WithLogDir("./test/debug"),
		WithSaveDraw(true),
		WithStdoutLevel(LoggingLevelInfo),
	)

	os.Exit(m.Run())
}
