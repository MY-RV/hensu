package codec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// JSON codec: object (flat or nested) ↔ flat KEY__PATH.
// Apply rewrites the body with indent 2 (no comment preservation).
type JSON struct{}

func (JSON) Decode(body string) (map[string]string, error) {
	return parseJSONBody(body)
}

func (JSON) Apply(body string, updates map[string]string) (string, error) {
	m, err := parseJSONBody(body)
	if err != nil {
		return "", err
	}
	if m == nil {
		m = map[string]string{}
	}
	for k, v := range updates {
		m[canonicalLookup(k)] = v
	}
	nested := nestFlatEnv(m)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(nested); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func parseJSONBody(body string) (map[string]string, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return map[string]string{}, nil
	}
	var root any
	dec := json.NewDecoder(strings.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&root); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}
	obj, ok := root.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("JSON root must be an object")
	}
	out := make(map[string]string)
	if err := flattenJSON(nil, obj, out); err != nil {
		return nil, err
	}
	return out, nil
}

func flattenJSON(prefix []string, v any, out map[string]string) error {
	switch t := v.(type) {
	case map[string]any:
		for k, child := range t {
			seg := normalizeSegment(k)
			if seg == "" {
				return fmt.Errorf("JSON key is empty")
			}
			// Allow embedded __ in a single key name.
			var segs []string
			if strings.Contains(seg, "__") {
				for _, p := range strings.Split(seg, "__") {
					if p == "" {
						return fmt.Errorf("JSON key %q has empty __ segment", k)
					}
					segs = append(segs, p)
				}
			} else {
				segs = []string{seg}
			}
			if err := flattenJSON(append(prefix, segs...), child, out); err != nil {
				return err
			}
		}
	case nil:
		return nil
	default:
		key := envKeyFromPath(prefix)
		if key == "" {
			return fmt.Errorf("JSON scalar without key path")
		}
		s, err := jsonValueToString(t)
		if err != nil {
			return err
		}
		out[key] = s
	}
	return nil
}

func jsonValueToString(v any) (string, error) {
	switch t := v.(type) {
	case nil:
		return "", nil
	case string:
		return t, nil
	case bool:
		return strconv.FormatBool(t), nil
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10), nil
		}
		return strconv.FormatFloat(t, 'f', -1, 64), nil
	case json.Number:
		return t.String(), nil
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}
