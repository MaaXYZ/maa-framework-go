package maa

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const adbFixtureRecordEnv = "MAA_GO_ADB_FIXTURE_RECORD"
const adbFixtureSerial = "maa-fixture-serial"

type adbFixtureCall struct {
	Args  []string `json:"args"`
	Error string   `json:"error,omitempty"`
}

func setupAdbFixture(t *testing.T) (adbPath, recordPath string) {
	t.Helper()
	adbPath, err := os.Executable()
	require.NoError(t, err)
	adbPath = filepath.ToSlash(adbPath)
	recordPath = filepath.Join(t.TempDir(), "adb-calls.jsonl")
	t.Setenv(adbFixtureRecordEnv, recordPath)
	return adbPath, recordPath
}

func runAdbFixture(recordPath string, args []string) int {
	output, monitor, err := adbFixtureResponse(args)
	call := adbFixtureCall{Args: args}
	if err != nil {
		call.Error = err.Error()
	}
	record, recordErr := json.Marshal(call)
	if recordErr != nil {
		return 2
	}
	file, recordErr := os.OpenFile(recordPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if recordErr != nil {
		return 2
	}
	_, recordErr = file.Write(append(record, '\n'))
	closeErr := file.Close()
	if err != nil || recordErr != nil || closeErr != nil {
		return 2
	}
	if monitor {
		// Native discovery owns this foreground monitor and terminates it when
		// destroying its temporary controller. No daemon or descendant is started.
		_, err = io.Copy(io.Discard, os.Stdin)
	} else {
		_, err = fmt.Fprint(os.Stdout, output)
	}
	if err != nil {
		return 2
	}
	return 0
}

func adbFixtureResponse(args []string) (output string, monitor bool, err error) {
	if len(args) == 1 && args[0] == "devices" {
		return "List of devices attached\n" + adbFixtureSerial + "\tdevice\noffline-serial\toffline\n", false, nil
	}
	if len(args) < 3 || args[0] != "-s" || args[1] != adbFixtureSerial {
		return "", false, fmt.Errorf("unexpected adb arguments: %q", args)
	}
	if len(args) == 3 && args[2] == "get-state" {
		return "device\n", false, nil
	}
	if len(args) == 4 && args[2] == "shell" {
		switch command := args[3]; {
		case command == "cat":
			return "", true, nil
		case command == "getprop sys.tencent.imei":
			return "\n", false, nil
		case strings.HasPrefix(command, "getprop "):
			return "fixture\n", false, nil
		}
	}
	return "", false, fmt.Errorf("unexpected adb arguments: %q", args)
}

func assertAdbFixtureCalls(t *testing.T, recordPath string) {
	t.Helper()
	data, err := os.ReadFile(recordPath)
	require.NoError(t, err)
	var calls [][]string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var call adbFixtureCall
		require.NoError(t, json.Unmarshal([]byte(line), &call))
		require.Empty(t, call.Error)
		calls = append(calls, call.Args)
	}
	require.Contains(t, calls, []string{"devices"})
	require.Contains(t, calls, []string{"-s", adbFixtureSerial, "get-state"})
	require.Contains(t, calls, []string{"-s", adbFixtureSerial, "shell", "cat"})
}
