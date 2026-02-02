package maa

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	Init(
		WithLogDir("./test/debug"),
		WithStdoutLevel(LoggingLevelOff),
	)

	os.Exit(m.Run())
}
