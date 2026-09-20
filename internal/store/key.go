package store

import "strings"

// CanonicalKey normalizes lookup sugar: ':' ≡ '__', uppercases.
func CanonicalKey(key string) string {
	key = strings.TrimSpace(key)
	key = strings.ReplaceAll(key, ":", "__")
	for strings.Contains(key, "____") {
		key = strings.ReplaceAll(key, "____", "__")
	}
	return strings.ToUpper(strings.TrimSpace(key))
}

// PathSegments splits a canonical key on '__' (empty segments dropped).
func PathSegments(key string) []string {
	key = CanonicalKey(key)
	if key == "" {
		return nil
	}
	parts := strings.Split(key, "__")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

// KeyFromPath joins segments with '__' after normalizing each segment.
func KeyFromPath(segments []string) string {
	if len(segments) == 0 {
		return ""
	}
	parts := make([]string, len(segments))
	for i, s := range segments {
		parts[i] = normalizeSegment(s)
	}
	return strings.Join(parts, "__")
}

func normalizeSegment(segment string) string {
	segment = strings.TrimSpace(segment)
	segment = strings.ReplaceAll(segment, "-", "_")
	return strings.ToUpper(segment)
}

// NormalizeMap canonicalizes keys (':' ≡ '__', uppercase); values unchanged.
func NormalizeMap(m map[string]string) Map {
	if len(m) == 0 {
		return Map{}
	}
	out := make(Map, len(m))
	for k, v := range m {
		nk := CanonicalKey(k)
		if nk == "" {
			continue
		}
		out[nk] = v
	}
	return out
}
