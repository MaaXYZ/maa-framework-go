package maa

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if recordPath := os.Getenv(adbFixtureRecordEnv); recordPath != "" {
		os.Exit(runAdbFixture(recordPath, os.Args[1:]))
	}

	Init(
		WithLogDir("./test/debug"),
		WithStdoutLevel(LoggingLevelOff),
	)

	os.Exit(m.Run())
}
