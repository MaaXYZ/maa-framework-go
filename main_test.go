package maa

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	Init(
		WithLogDir("./test/debug"),
		WithSaveDraw(true),
		WithStdoutLevel(LoggingLevelAll),
	)

	os.Exit(m.Run())
}
