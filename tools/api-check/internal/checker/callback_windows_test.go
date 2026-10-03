//go:build windows

package checker

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestCallbackABIWindowsRuntime compares the checker with actual registrations
// through the repository's purego dependency without loading MaaFramework.
func TestCallbackABIWindowsRuntime(t *testing.T) {
	repoRoot, err := detectRepoRoot()
	if err != nil {
		t.Fatalf("locate root module for purego callback oracle: %v", err)
	}
	sourcePath := filepath.Join(t.TempDir(), "callback_oracle.go")
	if err := os.WriteFile(sourcePath, []byte(windowsCallbackOracleSource), 0o600); err != nil {
		t.Fatalf("write Windows callback oracle: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "run", "-mod=readonly", sourcePath)
	command.Dir = repoRoot
	command.Env = append(os.Environ(), "GOWORK=off")
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		t.Fatalf("run Windows callback oracle: %v\nstdout:\n%s\nstderr:\n%s", err, output, stderr.String())
	}

	var results []struct {
		Name       string
		Registered bool
		Panicked   bool
		Panic      string
	}
	if err := json.Unmarshal(output, &results); err != nil {
		t.Fatalf("decode Windows callback oracle results: %v\nstdout:\n%s", err, output)
	}
	const pointerSizedResultPanic = "compileCallback: expected function with one uintptr-sized result"
	tests := []struct {
		name      string
		goReturns []string
		cReturns  []string
		want      bool
		wantPanic string
	}{
		{name: "uintptr", goReturns: []string{"ptr"}, cReturns: []string{"bool"}, want: true},
		{name: "int64", goReturns: []string{"int64"}, cReturns: []string{"int64"}, want: true},
		{name: "uint64", goReturns: []string{"uint64"}, cReturns: []string{"uint64"}, want: true},
		{name: "int32", goReturns: []string{"int32"}, cReturns: []string{"int32"}, want: false, wantPanic: pointerSizedResultPanic},
		{name: "float64", goReturns: []string{"float64"}, cReturns: []string{"float64"}, want: false, wantPanic: "compileCallback: float results not supported"},
		{name: "bool", goReturns: []string{"bool"}, cReturns: []string{"bool"}, want: false, wantPanic: pointerSizedResultPanic},
		{name: "void", want: false, wantPanic: pointerSizedResultPanic},
	}
	if len(results) != len(tests) {
		t.Fatalf("Windows callback oracle returned %d results, want %d: %s", len(results), len(tests), output)
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := results[i]
			if result.Name != tt.name {
				t.Fatalf("Windows callback oracle result name = %q, want %q", result.Name, tt.name)
			}
			if result.Registered != tt.want || result.Panicked != !tt.want {
				t.Fatalf("Windows callback registration = %+v, want registered=%t and panicked=%t", result, tt.want, !tt.want)
			}
			if result.Panic != tt.wantPanic {
				t.Fatalf("Windows callback panic = %q, want %q", result.Panic, tt.wantPanic)
			}
			goSig := methodSig{params: []string{"ptr"}, returns: tt.goReturns}
			cSig := methodSig{params: []string{"ptr"}, returns: tt.cReturns}
			if got := callbackABIMatches(goSig, cSig); got != result.Registered {
				t.Fatalf("callbackABIMatches(%v -> %v, %v -> %v) = %t, Windows registration accepted=%t", goSig.params, goSig.returns, cSig.params, cSig.returns, got, result.Registered)
			}
		})
	}
}

const windowsCallbackOracleSource = `package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ebitengine/purego"
)

type registrationResult struct {
	Name       string
	Registered bool
	Panicked   bool
	Panic      string
}

func register(name string, callback any) (result registrationResult) {
	result.Name = name
	defer func() {
		if problem := recover(); problem != nil {
			result.Panicked = true
			result.Panic = fmt.Sprint(problem)
		}
	}()
	result.Registered = purego.NewCallback(callback) != 0
	return result
}

func main() {
	results := []registrationResult{
		register("uintptr", func(handle uintptr) uintptr { return 0 }),
		register("int64", func(handle uintptr) int64 { return 0 }),
		register("uint64", func(handle uintptr) uint64 { return 0 }),
		register("int32", func(handle uintptr) int32 { return 0 }),
		register("float64", func(handle uintptr) float64 { return 0 }),
		register("bool", func(handle uintptr) bool { return false }),
		register("void", func(handle uintptr) {}),
	}
	if err := json.NewEncoder(os.Stdout).Encode(results); err != nil {
		panic(err)
	}
}
`
