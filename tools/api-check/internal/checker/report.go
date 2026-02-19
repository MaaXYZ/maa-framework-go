package checker

import (
	"fmt"
	"sort"
	"strings"
)

func printReport(report []string, issues []issue) int {
	fmt.Println("== API Consistency Check ==")
	for _, line := range report {
		fmt.Printf("- %s\n", line)
	}
	fmt.Println()

	if len(issues) == 0 {
		fmt.Println("PASS: no inconsistencies found.")
		return 0
	}

	grouped := map[string][]string{}
	for _, it := range issues {
		grouped[it.section] = append(grouped[it.section], it.message)
	}

	for _, section := range sectionOrder {
		msgs := grouped[section]
		if len(msgs) == 0 {
			continue
		}
		sort.Strings(msgs)
		fmt.Printf("## %s\n", section)
		for _, msg := range msgs {
			printIssue(msg)
		}
		fmt.Println()
	}

	fmt.Printf("FAIL: found %d inconsistency(s).\n", len(issues))
	return 1
}

func printIssue(msg string) {
	lines := strings.Split(msg, "\n")
	if len(lines) == 0 {
		return
	}
	fmt.Printf("- %s\n", lines[0])
	for _, line := range lines[1:] {
		fmt.Printf("  %s\n", line)
	}
}
