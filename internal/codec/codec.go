// Package codec implements on-disk format codecs for hensu.
package codec

import (
	"fmt"
	"sort"
	"strings"
)

// Codec converts between a body (without FORMAT header) and a flat map.
type Codec interface {
	Decode(body string) (map[string]string, error)
	Apply(body string, updates map[string]string) (string, error)
}

const formatDirectivePrefix = "format:"

// SplitHeader extracts # FORMAT: and the remaining body.
// No header → DOTENV and the full raw string as body.
func SplitHeader(raw string) (format string, body string, err error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "DOTENV", "", nil
	}
	firstLine, rest, ok := strings.Cut(trimmed, "\n")
	if !ok {
		firstLine = trimmed
		rest = ""
	}
	firstLine = strings.TrimSpace(firstLine)
	if !strings.HasPrefix(firstLine, "#") {
		return "DOTENV", raw, nil
	}
	f, found, err := ParseFormatDirective(firstLine)
	if err != nil {
		return "", "", err
	}
	if !found {
		return "DOTENV", raw, nil
	}
	return f, rest, nil
}

// ParseFormatDirective parses a comment line "# FORMAT: NAME".
func ParseFormatDirective(commentLine string) (format string, found bool, err error) {
	line := strings.TrimSpace(commentLine)
	if !strings.HasPrefix(line, "#") {
		return "", false, nil
	}
	line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
	if !strings.HasPrefix(strings.ToLower(line), formatDirectivePrefix) {
		return "", false, nil
	}
	rest := strings.TrimSpace(line[len(formatDirectivePrefix):])
	if rest == "" {
		return "", false, fmt.Errorf("FORMAT directive is empty")
	}
	name := strings.ToUpper(strings.TrimSpace(rest))
	switch name {
	case "DOTENV", "SHEXPORT", "YAML", "JSON":
		return name, true, nil
	default:
		return "", false, fmt.Errorf("FORMAT must be YAML, DOTENV, SHEXPORT, or JSON (got %q)", rest)
	}
}

// FirstLineIsFormatDirective reports whether the first line is a FORMAT header.
func FirstLineIsFormatDirective(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return false
	}
	first, _, _ := strings.Cut(trimmed, "\n")
	_, found, err := ParseFormatDirective(strings.TrimSpace(first))
	return err == nil && found
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
