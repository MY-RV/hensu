package codec

import (
	"strconv"
	"strings"
	"unicode"
)

// Dotenv is KEY=value without forcing export prefixes on write.
type Dotenv struct{}

func (Dotenv) Decode(body string) (map[string]string, error) {
	return parseEnvLines(body), nil
}

func (Dotenv) Apply(body string, updates map[string]string) (string, error) {
	return applyEnvUpdates(body, updates, false)
}

// Shexport is KEY=value forcing export on write.
type Shexport struct{}

func (Shexport) Decode(body string) (map[string]string, error) {
	return parseEnvLines(body), nil
}

func (Shexport) Apply(body string, updates map[string]string) (string, error) {
	return applyEnvUpdates(body, updates, true)
}

func parseEnvLines(raw string) map[string]string {
	m := make(map[string]string)
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		line = strings.TrimSpace(line)
		k, vpart, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		m[strings.ToUpper(strings.TrimSpace(k))] = parseValue(vpart)
	}
	return m
}

func parseValue(vpart string) string {
	s := strings.TrimSpace(vpart)
	if s == "" {
		return ""
	}
	switch s[0] {
	case '"':
		val, _ := parseDoubleQuoted(s)
		return val
	case '\'':
		val, _ := parseSingleQuoted(s)
		return val
	default:
		if i := strings.Index(s, " #"); i >= 0 {
			s = strings.TrimSpace(s[:i])
		}
		return s
	}
}

func parseDoubleQuoted(s string) (value string, end int) {
	if len(s) == 0 || s[0] != '"' {
		return "", 0
	}
	var b strings.Builder
	rest := s[1:]
	for len(rest) > 0 {
		if rest[0] == '"' {
			return b.String(), len(s) - len(rest) + 1
		}
		r, _, tail, err := strconv.UnquoteChar(rest, '"')
		if err != nil {
			b.WriteString(rest)
			return b.String(), len(s)
		}
		b.WriteRune(r)
		rest = tail
	}
	return b.String(), len(s)
}

func parseSingleQuoted(s string) (value string, end int) {
	i := 1
	for i < len(s) {
		if s[i] == '\'' {
			return s[1:i], i + 1
		}
		i++
	}
	return s[1:], len(s)
}

func applyEnvUpdates(body string, updates map[string]string, forceExport bool) (string, error) {
	pending := make(map[string]string, len(updates))
	for k, v := range updates {
		pending[k] = v
	}

	lines := splitLinesPreserve(body)
	out := make([]string, 0, len(lines)+len(updates))
	for _, line := range lines {
		key, exportPref, trail, ok := parseEnvAssignmentLine(line)
		if !ok {
			out = append(out, line)
			continue
		}
		nk := canonicalLookup(key)
		if newVal, hit := pending[nk]; hit {
			useExport := forceExport || exportPref
			out = append(out, formatEnvAssignment(nk, newVal, useExport, trail))
			delete(pending, nk)
			continue
		}
		out = append(out, line)
	}

	if len(pending) > 0 {
		if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		for _, k := range sortedKeys(pending) {
			out = append(out, formatEnvAssignment(k, pending[k], forceExport, ""))
		}
	}
	return joinLinesPreserve(out), nil
}

func splitLinesPreserve(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	parts := strings.Split(s, "\n")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func joinLinesPreserve(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func parseEnvAssignmentLine(line string) (key string, exportPref bool, trailingComment string, ok bool) {
	trim := strings.TrimSpace(line)
	if trim == "" || strings.HasPrefix(trim, "#") {
		return "", false, "", false
	}
	work := trim
	if strings.HasPrefix(work, "export ") {
		exportPref = true
		work = strings.TrimSpace(strings.TrimPrefix(work, "export "))
	}
	k, vpart, cutOK := strings.Cut(work, "=")
	if !cutOK {
		return "", false, "", false
	}
	k = strings.TrimSpace(k)
	if k == "" || strings.ContainsAny(k, " \t") {
		return "", false, "", false
	}
	trailingComment = extractTrailingComment(vpart)
	return k, exportPref, trailingComment, true
}

func extractTrailingComment(vpart string) string {
	s := strings.TrimSpace(vpart)
	if s == "" {
		return ""
	}
	switch s[0] {
	case '"':
		_, end := parseDoubleQuoted(s)
		rest := strings.TrimSpace(s[end:])
		if strings.HasPrefix(rest, "#") {
			return rest
		}
		return ""
	case '\'':
		_, end := parseSingleQuoted(s)
		rest := strings.TrimSpace(s[end:])
		if strings.HasPrefix(rest, "#") {
			return rest
		}
		return ""
	default:
		if i := strings.Index(s, " #"); i >= 0 {
			return strings.TrimSpace(s[i:])
		}
		return ""
	}
}

func formatEnvAssignment(key, value string, exportPref bool, trailingComment string) string {
	rendered := renderEnvValue(value)
	var b strings.Builder
	if exportPref {
		b.WriteString("export ")
	}
	b.WriteString(key)
	b.WriteByte('=')
	b.WriteString(rendered)
	if trailingComment != "" {
		if !strings.HasPrefix(trailingComment, "#") {
			trailingComment = "# " + trailingComment
		}
		b.WriteByte(' ')
		b.WriteString(trailingComment)
	}
	return b.String()
}

func renderEnvValue(value string) string {
	if value == "" {
		return `""`
	}
	needsQuote := strings.ContainsAny(value, " \t\n\r\"'#\\")
	for _, r := range value {
		if !unicode.IsPrint(r) && r != '\n' && r != '\t' {
			needsQuote = true
			break
		}
	}
	if !needsQuote {
		return value
	}
	return strconv.Quote(value)
}

func canonicalLookup(key string) string {
	key = strings.TrimSpace(key)
	key = strings.ReplaceAll(key, ":", "__")
	for strings.Contains(key, "____") {
		key = strings.ReplaceAll(key, "____", "__")
	}
	return strings.ToUpper(strings.TrimSpace(key))
}
