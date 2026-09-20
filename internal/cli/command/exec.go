package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/my-rv/hensu"
)

// ExitError carries a child process exit status so hensu can exit with it.
type ExitError struct {
	Cmd  string
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("%s exited with code %d", e.Cmd, e.Code)
}

// Runner starts a child process. Tests replace it; production uses OSRunner.
type Runner func(ctx context.Context, name string, args, env []string, stdin io.Reader, stdout, stderr io.Writer) error

// OSRunner runs the command with os/exec, passing stdio through untouched.
func OSRunner(ctx context.Context, name string, args, env []string, stdin io.Reader, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = env
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		// A negative code means a signal killed the child; there is no status
		// to forward, so let the error surface as an ordinary failure.
		if errors.As(err, &ee) && ee.ExitCode() >= 0 {
			return &ExitError{Cmd: name, Code: ee.ExitCode()}
		}
		return err
	}
	return nil
}

// cmdExec runs a child process with the config in its environment.
//
// This is the path that makes the masked modes usable: the process that needs
// the value gets it, and the caller that spawned it never sees one. Values are
// not printed, not logged, and not echoed on failure.
func cmdExec(ctx *Context, args []string) error {
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}
	if len(args) == 0 {
		return &hensu.InvalidArgument{Msg: "exec: expected a command, as: hensu exec -- CMD [ARGS...]"}
	}
	m, _, err := ctx.Store.Load(ctx.Req())
	if err != nil {
		return err
	}
	env, skipped := childEnv(ctx.Environ(), m)
	for _, k := range skipped {
		fmt.Fprintf(ctx.Stderr, "hensu: exec: skipping %q (not a valid environment variable name)\n", k)
	}
	run := ctx.Runner
	if run == nil {
		run = OSRunner
	}
	return run(ctx.Req(), args[0], args[1:], env, ctx.Stdin, ctx.Stdout, ctx.Stderr)
}

// childEnv layers the config over the inherited environment. The file wins:
// it is the machine's answer for those keys, and a stale inherited value is
// the bug hensu exists to remove.
//
// Keys that are not valid environment variable names are skipped and returned,
// so a YAML mapping with a spaced key cannot smuggle anything into the child.
func childEnv(parent []string, m hensu.Map) (env []string, skipped []string) {
	merged := make(map[string]string, len(parent)+len(m))
	order := make([]string, 0, len(parent)+len(m))
	put := func(k, v string) {
		if _, seen := merged[k]; !seen {
			order = append(order, k)
		}
		merged[k] = v
	}
	for _, kv := range parent {
		i := strings.IndexByte(kv, '=')
		if i <= 0 {
			continue
		}
		put(kv[:i], kv[i+1:])
	}
	for _, kv := range sortedPairs(m) {
		if !validEnvName(kv.key) {
			skipped = append(skipped, kv.key)
			continue
		}
		put(kv.key, kv.value)
	}
	sort.Strings(skipped)
	// Nothing about the reveal mode is passed down: the default is already
	// safe, and a grant this call was given is not a grant its children get.
	env = make([]string, 0, len(order))
	for _, k := range order {
		env = append(env, k+"="+merged[k])
	}
	return env, skipped
}

type pair struct{ key, value string }

func sortedPairs(m hensu.Map) []pair {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]pair, 0, len(keys))
	for _, k := range keys {
		out = append(out, pair{key: k, value: m[k]})
	}
	return out
}

// validEnvName is the POSIX name rule: letters, digits, underscore; no leading
// digit. Anything else (spaces, '=', dots) cannot be exported safely.
func validEnvName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		switch {
		case r == '_':
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9' && i > 0:
		default:
			return false
		}
	}
	return true
}

// Environ returns the inherited environment (overridable in tests).
func (c *Context) Environ() []string {
	if c == nil || c.Env == nil {
		return os.Environ()
	}
	return c.Env()
}
