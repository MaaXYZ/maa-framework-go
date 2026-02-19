package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func checkNativeAPICoverage(headerDir string, blacklist map[string]struct{}) ([]issue, error) {
	goRegistered, err := parseGoRegistrations()
	if err != nil {
		return nil, fmt.Errorf("parse Go registrations: %w", err)
	}

	headerFuncs, err := parseHeaderFunctions(headerDir)
	if err != nil {
		return nil, fmt.Errorf("parse C headers: %w", err)
	}

	issues := make([]issue, 0)
	for _, module := range moduleOrder {
		goSet := goRegistered[module]
		headerSet := headerFuncs[module]

		headerOnly := setDiff(headerSet, goSet, blacklist)
		goOnly := setDiff(goSet, headerSet, blacklist)

		for _, fn := range headerOnly {
			issues = append(issues, issue{
				section: sectionNativeAPI,
				message: fmt.Sprintf("[%s] header has function but Go is not registering it: %s", module, fn),
			})
		}
		for _, fn := range goOnly {
			issues = append(issues, issue{
				section: sectionNativeAPI,
				message: fmt.Sprintf("[%s] Go registers function not found in headers: %s", module, fn),
			})
		}
	}

	return issues, nil
}

func parseGoRegistrations() (map[string]map[string]struct{}, error) {
	result := map[string]map[string]struct{}{}
	re := regexp.MustCompile(`RegisterLibFunc\(&[A-Za-z0-9_]+,\s*[A-Za-z0-9_]+,\s*"([A-Za-z0-9_]+)"\)`)

	for module, files := range nativeFilesByModule {
		if _, ok := result[module]; !ok {
			result[module] = map[string]struct{}{}
		}
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", file, err)
			}
			matches := re.FindAllStringSubmatch(string(data), -1)
			for _, m := range matches {
				if len(m) < 2 {
					continue
				}
				result[module][m[1]] = struct{}{}
			}
		}
	}

	return result, nil
}

func parseHeaderFunctions(headerDir string) (map[string]map[string]struct{}, error) {
	result := map[string]map[string]struct{}{
		"framework":    {},
		"toolkit":      {},
		"agent_server": {},
		"agent_client": {},
	}

	err := filepath.WalkDir(headerDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(path) != ".h" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		chunks := strings.Split(content, ";")
		for _, chunk := range chunks {
			stmt := strings.TrimSpace(chunk)
			if stmt == "" || !strings.Contains(stmt, "MAA_") {
				continue
			}
			if !strings.Contains(stmt, "_API") {
				continue
			}
			if strings.Contains(stmt, "MAA_DEPRECATED") {
				continue
			}

			module := moduleFromHeaderStmt(stmt)
			if module == "" {
				continue
			}
			name := extractFuncNameFromDecl(stmt)
			if name == "" {
				continue
			}
			result[module][name] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func moduleFromHeaderStmt(stmt string) string {
	switch {
	case strings.Contains(stmt, "MAA_FRAMEWORK_API"):
		return "framework"
	case strings.Contains(stmt, "MAA_TOOLKIT_API"):
		return "toolkit"
	case strings.Contains(stmt, "MAA_AGENT_SERVER_API"):
		return "agent_server"
	case strings.Contains(stmt, "MAA_AGENT_CLIENT_API"):
		return "agent_client"
	default:
		return ""
	}
}

func extractFuncNameFromDecl(stmt string) string {
	re := regexp.MustCompile(`\b(Maa[A-Za-z0-9_]+)\s*\(`)
	matches := re.FindAllStringSubmatch(stmt, -1)
	if len(matches) == 0 {
		return ""
	}
	last := matches[len(matches)-1]
	if len(last) < 2 {
		return ""
	}
	return strings.TrimSpace(last[1])
}

func setDiff(left, right map[string]struct{}, blacklist map[string]struct{}) []string {
	out := make([]string, 0)
	for k := range left {
		if _, ignored := blacklist[k]; ignored {
			continue
		}
		if _, ok := right[k]; !ok {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
