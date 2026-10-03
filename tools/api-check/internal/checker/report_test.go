package checker

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestWriteReportPass(t *testing.T) {
	var buf bytes.Buffer
	code := writeReport(&buf, []string{"repo_root: /repo", "blacklist_size: 0"}, nil)
	out := buf.String()

	if code != 0 {
		t.Fatalf("writeReport() = %d, want 0", code)
	}
	for _, want := range []string{
		"== API Consistency Check ==",
		"- repo_root: /repo\n",
		"- blacklist_size: 0\n",
		"PASS: no inconsistencies found.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "FAIL") || strings.Contains(out, "## ") {
		t.Fatalf("passing output should not contain sections or FAIL:\n%s", out)
	}
}

func TestWriteReportDeterministicSections(t *testing.T) {
	issues := []issue{
		{section: "Zebra Section", message: "z2"},
		{section: sectionController, message: "b"},
		{section: "Alpha Section", message: "a2"},
		{section: sectionNativeAPI, message: "n2"},
		{section: sectionController, message: "a"},
		{section: sectionNativeAPI, message: "n1"},
		{section: sectionPipeline, message: "p1"},
		{section: sectionControllerMethod, message: "m1"},
		{section: "Zebra Section", message: "z1"},
		{section: "Alpha Section", message: "a1"},
	}

	var first, second bytes.Buffer
	if code := writeReport(&first, []string{"config: <none> (using defaults)"}, issues); code != 1 {
		t.Fatalf("writeReport() = %d, want 1", code)
	}
	if code := writeReport(&second, []string{"config: <none> (using defaults)"}, issues); code != 1 {
		t.Fatalf("second writeReport() = %d, want 1", code)
	}
	if first.String() != second.String() {
		t.Fatalf("report output is not deterministic:\n--- first ---\n%s\n--- second ---\n%s", first.String(), second.String())
	}

	out := first.String()
	order := []string{
		sectionNativeAPI,
		sectionController,
		sectionControllerMethod,
		sectionPipeline,
		"Alpha Section",
		"Zebra Section",
	}
	previous := -1
	for _, section := range order {
		at := strings.Index(out, "## "+section+"\n")
		if at < 0 {
			t.Fatalf("section %q missing from report:\n%s", section, out)
		}
		if at < previous {
			t.Fatalf("section %q out of order:\n%s", section, out)
		}
		previous = at
	}

	if strings.Index(out, "- a\n") > strings.Index(out, "- b\n") {
		t.Fatalf("messages within a section are not sorted:\n%s", out)
	}
	if !strings.Contains(out, "FAIL: found 10 inconsistency(s).") {
		t.Fatalf("issue count missing:\n%s", out)
	}
	for _, msg := range []string{"a1", "a2", "b", "m1", "n1", "n2", "p1", "z1", "z2"} {
		if !strings.Contains(out, "- "+msg+"\n") {
			t.Fatalf("issue %q silently omitted:\n%s", msg, out)
		}
	}
}

func TestWriteReportMultilineMessage(t *testing.T) {
	var buf bytes.Buffer
	writeReport(&buf, nil, []issue{{
		section: sectionNativeAPI,
		message: "first line\nsecond line\nthird line",
	}})

	want := "## " + sectionNativeAPI + "\n- first line\n  second line\n  third line\n"
	if !strings.Contains(buf.String(), want) {
		t.Fatalf("multiline issue not indented, want:\n%s\ngot:\n%s", want, buf.String())
	}
}

func TestWriteReportUnknownSectionKept(t *testing.T) {
	var buf bytes.Buffer
	writeReport(&buf, nil, []issue{
		{section: "Unmapped Section", message: "unknown issue"},
		{section: sectionPipeline, message: "known issue"},
	})

	out := buf.String()
	if !strings.Contains(out, "## Unmapped Section\n- unknown issue\n") {
		t.Fatalf("unknown section missing:\n%s", out)
	}
	if strings.Index(out, "## "+sectionPipeline) > strings.Index(out, "## Unmapped Section") {
		t.Fatalf("known section should precede fallback sections:\n%s", out)
	}
}

func TestPrintReportWrapperWritesStdout(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = original }()

	code := printReport([]string{"line"}, nil)
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = original

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Fatalf("printReport() = %d, want 0", code)
	}
	if !strings.Contains(string(data), "- line\n") || !strings.Contains(string(data), "PASS") {
		t.Fatalf("printReport() output = %q", data)
	}
}
