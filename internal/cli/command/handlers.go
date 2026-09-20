package command

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/my-rv/hensu"
)

// DefaultRegistry registers production CLI verbs.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(NewFunc("get", cmdGet))
	r.Register(NewFunc("keys", cmdKeys))
	r.Register(NewFunc("read", cmdRead))
	r.Register(NewFunc("set", cmdSet))
	r.Register(NewFunc("exec", cmdExec))
	return r
}

// cmdGet prints the named keys at the invocation's reveal mode. There is no
// bulk read: naming the key is the point.
func cmdGet(ctx *Context, args []string) error {
	entries, err := ctx.Store.Get(ctx.Req(), args, ctx.Reveal)
	if err != nil {
		return err
	}
	return writeJSON(ctx, entries)
}

// cmdRead prints one value verbatim. That is a grant, so it has to be asked
// for: masking the output would defeat the command's only purpose.
func cmdRead(ctx *Context, args []string) error {
	if len(args) != 1 {
		return &hensu.InvalidArgument{Msg: "read: expected exactly one KEY"}
	}
	if ctx.Reveal != hensu.RevealTrust {
		return &hensu.InvalidArgument{Msg: "read prints the raw value; pass -r trust, " +
			"or hand the value to a process with `hensu exec -- CMD`"}
	}
	v, err := ctx.Store.Read(ctx.Req(), args[0])
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	_, err = io.WriteString(ctx.Stdout, v)
	return err
}

func cmdKeys(ctx *Context, args []string) error {
	if len(args) != 0 {
		return &hensu.InvalidArgument{Msg: fmt.Sprintf("keys: unexpected arguments %v", args)}
	}
	keys, err := ctx.Store.Keys(ctx.Req())
	if err != nil {
		return err
	}
	return writeJSON(ctx, keys)
}

// writeJSON is the one output encoding. JSON is the communication layer;
// anything that needs another shape pipes it somewhere else.
func writeJSON(ctx *Context, v any) error {
	enc := json.NewEncoder(ctx.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func cmdSet(ctx *Context, args []string) error {
	updates, err := ParseSetArgs(args, ctx.Stdin)
	if err != nil {
		return err
	}
	return ctx.Store.Set(ctx.Req(), updates)
}

// ParseSetArgs parses set KEY VALUE | set --json … .
func ParseSetArgs(args []string, stdin io.Reader) (map[string]string, error) {
	if len(args) == 0 {
		return nil, &hensu.InvalidArgument{Msg: "set: expected KEY VALUE or --json"}
	}
	if args[0] == "--json" {
		if len(args) != 2 {
			return nil, &hensu.InvalidArgument{Msg: "set --json: expected a JSON object or '-' for stdin"}
		}
		return parseSetJSON(args[1], stdin)
	}
	if len(args) < 2 {
		return nil, &hensu.InvalidArgument{Msg: "set: expected KEY VALUE"}
	}
	key := args[0]
	value := strings.Join(args[1:], " ")
	return map[string]string{key: value}, nil
}

func parseSetJSON(src string, stdin io.Reader) (map[string]string, error) {
	var raw []byte
	var err error
	if src == "-" {
		raw, err = io.ReadAll(stdin)
		if err != nil {
			return nil, err
		}
	} else {
		raw = []byte(src)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, &hensu.ParseError{Msg: "set --json", Err: err}
	}
	if len(obj) == 0 {
		return nil, &hensu.InvalidArgument{Msg: "set --json: empty object"}
	}
	out := make(map[string]string, len(obj))
	for k, v := range obj {
		s, err := jsonValueToStoredString(v)
		if err != nil {
			return nil, &hensu.InvalidArgument{Msg: fmt.Sprintf("set --json key %q: %v", k, err)}
		}
		out[k] = s
	}
	return out, nil
}

func jsonValueToStoredString(v any) (string, error) {
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
