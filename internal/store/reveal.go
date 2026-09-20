package store

import (
	"strings"
	"unicode/utf8"
)

// Lookup builds reveal entries for the given keys against m.
func Lookup(m Map, keys []string, mode Reveal) map[string]Entry {
	out := make(map[string]Entry, len(keys))
	for _, raw := range keys {
		k := CanonicalKey(raw)
		if k == "" {
			continue
		}
		entry := Entry{Defined: false}
		if v, ok := m[k]; ok {
			entry.Defined = true
			entry.Value = mode.Mask(v)
		}
		out[k] = entry
	}
	return out
}

// Mask applies the reveal mode to a defined value.
func (mode Reveal) Mask(value string) string {
	switch mode {
	case RevealTrust:
		return value
	case RevealMask:
		return redactFully(value)
	default:
		return peekValue(value)
	}
}

func redactFully(value string) string {
	if value == "" {
		return ""
	}
	return "***"
}

func peekValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	n := utf8.RuneCountInString(value)
	if n <= ShortValueMax {
		return "***"
	}
	runes := []rune(value)
	return string(runes[:4]) + "***" + string(runes[len(runes)-4:])
}
