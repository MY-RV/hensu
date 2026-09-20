package codec

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAML codec: nested mapping ↔ flat KEY__PATH.
type YAML struct{}

func (YAML) Decode(body string) (map[string]string, error) {
	return parseYAMLBody(body)
}

func (YAML) Apply(body string, updates map[string]string) (string, error) {
	return applyYAMLUpdates(body, updates)
}

func parseYAMLBody(body string) (map[string]string, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return map[string]string{}, nil
	}
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(body), &root); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return map[string]string{}, nil
	}
	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("YAML root must be a mapping")
	}
	out := make(map[string]string)
	if err := flattenYAMLNode(nil, doc, out); err != nil {
		return nil, err
	}
	return out, nil
}

func flattenYAMLNode(prefix []string, node *yaml.Node, out map[string]string) error {
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			segments, err := yamlKeySegments(keyNode)
			if err != nil {
				return err
			}
			if err := flattenYAMLNode(append(prefix, segments...), valNode, out); err != nil {
				return err
			}
		}
	case yaml.ScalarNode:
		if node.Tag == "!!null" {
			return nil
		}
		key := envKeyFromPath(prefix)
		if key == "" {
			return fmt.Errorf("YAML scalar without key path")
		}
		out[key] = scalarToString(node)
	default:
		return fmt.Errorf("unsupported YAML node kind %v at %s", node.Kind, envKeyFromPath(prefix))
	}
	return nil
}

func yamlKeySegments(node *yaml.Node) ([]string, error) {
	if node.Kind != yaml.ScalarNode {
		return nil, fmt.Errorf("YAML mapping key must be scalar")
	}
	raw := strings.TrimSpace(node.Value)
	if raw == "" {
		return nil, fmt.Errorf("YAML mapping key is empty")
	}
	if strings.Contains(raw, "__") {
		parts := strings.Split(raw, "__")
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			part = normalizeSegment(part)
			if part == "" {
				return nil, fmt.Errorf("YAML mapping key %q has empty __ segment", raw)
			}
			out = append(out, part)
		}
		return out, nil
	}
	return []string{normalizeSegment(raw)}, nil
}

func envKeyFromPath(segments []string) string {
	if len(segments) == 0 {
		return ""
	}
	parts := make([]string, len(segments))
	for i, segment := range segments {
		parts[i] = normalizeSegment(segment)
	}
	return strings.Join(parts, "__")
}

func normalizeSegment(segment string) string {
	segment = strings.TrimSpace(segment)
	segment = strings.ReplaceAll(segment, "-", "_")
	return strings.ToUpper(segment)
}

func scalarToString(node *yaml.Node) string {
	if node.Tag == "!!bool" {
		switch strings.ToLower(strings.TrimSpace(node.Value)) {
		case "true", "yes", "on", "1":
			return "true"
		default:
			return "false"
		}
	}
	if node.Tag == "!!int" || node.Tag == "!!float" {
		return node.Value
	}
	if node.Tag == "!!null" {
		return ""
	}
	return node.Value
}

func applyYAMLUpdates(body string, updates map[string]string) (string, error) {
	body = strings.TrimRight(body, "\n") + "\n"
	if strings.TrimSpace(body) == "" {
		var root yaml.Node
		root.Kind = yaml.MappingNode
		root.Tag = "!!map"
		for _, k := range sortedKeys(updates) {
			if err := upsertYAMLPath(&root, pathSegments(k), updates[k]); err != nil {
				return "", err
			}
		}
		return encodeYAMLMapping(&root)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(body), &doc); err != nil {
		return "", fmt.Errorf("parse YAML: %w", err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return "", fmt.Errorf("YAML document is empty")
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return "", fmt.Errorf("YAML root must be a mapping")
	}
	for _, k := range sortedKeys(updates) {
		if err := upsertYAMLPath(root, pathSegments(k), updates[k]); err != nil {
			return "", err
		}
	}
	return encodeYAMLDocument(&doc)
}

func encodeYAMLDocument(doc *yaml.Node) (string, error) {
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return "", fmt.Errorf("invalid YAML document node")
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(4)
	if err := enc.Encode(doc.Content[0]); err != nil {
		_ = enc.Close()
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func encodeYAMLMapping(root *yaml.Node) (string, error) {
	doc := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}}
	return encodeYAMLDocument(doc)
}

func upsertYAMLPath(root *yaml.Node, segments []string, value string) error {
	if len(segments) == 0 {
		return fmt.Errorf("empty YAML path")
	}
	node := root
	for i, seg := range segments {
		if node.Kind != yaml.MappingNode {
			return fmt.Errorf("YAML path %q is not a mapping", strings.Join(segments[:i], "__"))
		}
		last := i == len(segments)-1
		idx := findYAMLMapKey(node, seg)
		if idx < 0 {
			keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: seg}
			var valNode *yaml.Node
			if last {
				valNode = scalarNodeForValue(value)
			} else {
				valNode = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			}
			node.Content = append(node.Content, keyNode, valNode)
			if last {
				return nil
			}
			node = valNode
			continue
		}
		valNode := node.Content[idx+1]
		if last {
			keepHead := valNode.HeadComment
			keepLine := valNode.LineComment
			keepFoot := valNode.FootComment
			*valNode = *scalarNodeForValue(value)
			valNode.HeadComment = keepHead
			valNode.LineComment = keepLine
			valNode.FootComment = keepFoot
			return nil
		}
		if valNode.Kind != yaml.MappingNode {
			return fmt.Errorf("YAML path %q is not a mapping", strings.Join(segments[:i+1], "__"))
		}
		node = valNode
	}
	return nil
}

func findYAMLMapKey(m *yaml.Node, want string) int {
	want = normalizeSegment(want)
	for i := 0; i < len(m.Content); i += 2 {
		keyNode := m.Content[i]
		segs, err := yamlKeySegments(keyNode)
		if err != nil || len(segs) != 1 {
			if err == nil && canonicalLookup(keyNode.Value) == want {
				return i
			}
			continue
		}
		if segs[0] == want {
			return i
		}
	}
	return -1
}

func scalarNodeForValue(value string) *yaml.Node {
	n := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
	if strings.Contains(value, "\n") {
		n.Style = yaml.LiteralStyle
	}
	return n
}

func pathSegments(key string) []string {
	key = canonicalLookup(key)
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

// FlatToYAML nests a flat map into YAML with FORMAT header (dump command).
func FlatToYAML(m map[string]string) (string, error) {
	root := nestFlatEnv(m)
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(4)
	if err := enc.Encode(root); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("# FORMAT: YAML\n")
	b.WriteString("# Canonical config\n\n")
	b.Write(buf.Bytes())
	return b.String(), nil
}

func nestFlatEnv(m map[string]string) map[string]interface{} {
	keys := sortedKeys(m)
	root := make(map[string]interface{})
	for _, key := range keys {
		setNested(root, strings.Split(key, "__"), m[key])
	}
	return root
}

func setNested(node map[string]interface{}, parts []string, value string) {
	if len(parts) == 0 {
		return
	}
	if len(parts) == 1 {
		node[parts[0]] = value
		return
	}
	head := parts[0]
	child, ok := node[head].(map[string]interface{})
	if !ok {
		child = make(map[string]interface{})
		node[head] = child
	}
	setNested(child, parts[1:], value)
}