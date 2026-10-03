package checker

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// printReport writes the report to standard output. It is retained as a thin
// compatibility wrapper around writeReport.
func printReport(report []string, issues []issue) int {
	return writeReport(os.Stdout, report, issues)
}

// writeReport renders the summary and issues to w and returns the process exit
// code (0 without issues, 1 otherwise). Known sections follow sectionOrder and
// any other section is appended in sorted order, so no issue is silently
// dropped and repeated runs produce identical output.
func writeReport(w io.Writer, report []string, issues []issue) int {
	fmt.Fprintln(w, "== API Consistency Check ==")
	for _, line := range report {
		fmt.Fprintf(w, "- %s\n", line)
	}
	fmt.Fprintln(w)

	if len(issues) == 0 {
		fmt.Fprintln(w, "PASS: no inconsistencies found.")
		return 0
	}

	grouped := make(map[string][]string, len(sectionOrder))
	known := make(map[string]struct{}, len(sectionOrder))
	for _, section := range sectionOrder {
		known[section] = struct{}{}
	}
	for _, it := range issues {
		grouped[it.section] = append(grouped[it.section], it.message)
	}

	sections := make([]string, 0, len(grouped))
	sections = append(sections, sectionOrder...)
	extra := make([]string, 0, len(grouped))
	for section := range grouped {
		if _, ok := known[section]; !ok {
			extra = append(extra, section)
		}
	}
	sort.Strings(extra)
	sections = append(sections, extra...)

	for _, section := range sections {
		msgs := grouped[section]
		if len(msgs) == 0 {
			continue
		}
		sort.Strings(msgs)
		fmt.Fprintf(w, "## %s\n", section)
		for _, msg := range msgs {
			writeIssue(w, msg)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "FAIL: found %d inconsistency(s).\n", len(issues))
	return 1
}

// writeIssue writes one issue message, indenting continuation lines so
// multiline messages stay attached to their bullet.
func writeIssue(w io.Writer, msg string) {
	lines := strings.Split(msg, "\n")
	fmt.Fprintf(w, "- %s\n", lines[0])
	for _, line := range lines[1:] {
		fmt.Fprintf(w, "  %s\n", line)
	}
}
