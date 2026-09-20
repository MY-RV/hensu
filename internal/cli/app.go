package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/my-rv/hensu"
	"github.com/my-rv/hensu/internal/cli/command"
	"github.com/my-rv/hensu/internal/update"
)

// App is the CLI application.
type App struct {
	Stdout   io.Writer
	Stderr   io.Writer
	Stdin    io.Reader
	Getwd    func() (string, error)
	Commands *command.Registry
	NewStore func(path string) *hensu.Store
	Ctx      context.Context
	// UpdateClient overrides release checks (tests).
	UpdateClient *update.Client
	Executable   func() (string, error)
	// LookupEnv reads the ambient reveal mode (tests).
	LookupEnv func(string) (string, bool)
	// Runner spawns exec children (tests).
	Runner command.Runner
	// Environ is the environment exec hands to children (tests).
	Environ func() []string
}

// ExitError is a child process exit status propagated by exec.
type ExitError = command.ExitError

// New returns an App wired to process stdio and the default command set.
func New() *App {
	return &App{
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		Stdin:    os.Stdin,
		Getwd:    os.Getwd,
		Commands: command.DefaultRegistry(),
	}
}

// Run executes hensu with argv (without the program name).
func (a *App) Run(args []string) error {
	opts, rest, err := ParseGlobalFlags(args)
	if err != nil {
		return err
	}
	if len(rest) > 0 && (rest[0] == "help" || rest[0] == "-h" || rest[0] == "--help") {
		Usage(a.Stderr)
		return nil
	}
	if len(rest) > 0 && (rest[0] == "version" || rest[0] == "--version") {
		fmt.Fprintln(a.Stdout, hensu.Version)
		return nil
	}
	if len(rest) > 0 && rest[0] == "--update" {
		return a.runUpdate(false)
	}
	if len(rest) > 0 && rest[0] == "--update-check" {
		return a.runUpdate(true)
	}

	reveal, err := a.reveal(opts)
	if err != nil {
		return err
	}
	path, err := a.resolvePath(opts.File)
	if err != nil {
		return err
	}
	newStore := a.NewStore
	if newStore == nil {
		newStore = func(path string) *hensu.Store { return hensu.NewStore(path) }
	}
	ctx := &command.Context{
		Ctx:    a.Ctx,
		Store:  newStore(path),
		Stdout: a.Stdout,
		Stderr: a.Stderr,
		Stdin:  a.Stdin,
		Reveal: reveal,
		Runner: a.Runner,
		Env:    a.Environ,
	}

	if len(rest) == 0 {
		// No bulk read: hensu with no command prints usage, not the config.
		Usage(a.Stderr)
		return &hensu.InvalidArgument{Msg: "expected a command"}
	}
	cmd, ok := a.Commands.Lookup(rest[0])
	if !ok {
		Usage(a.Stderr)
		return &hensu.InvalidArgument{Msg: fmt.Sprintf("unknown command %q", rest[0])}
	}
	return cmd.Run(ctx, rest[1:])
}

func (a *App) runUpdate(checkOnly bool) error {
	client := a.UpdateClient
	if client == nil {
		client = &update.Client{}
	}
	cand, err := client.Latest()
	if err != nil {
		return err
	}
	newer, err := update.Newer(hensu.Version, cand.Tag)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.Stdout, "current: %s\nlatest:  %s\n", hensu.Version, cand.Tag)
	if !newer {
		fmt.Fprintln(a.Stdout, "already up to date")
		return nil
	}
	if checkOnly {
		fmt.Fprintln(a.Stdout, "update available")
		return nil
	}
	if cand.DownloadURL == "" {
		return fmt.Errorf("update available (%s) but no download URL for this platform", cand.Tag)
	}
	exeFn := a.Executable
	if exeFn == nil {
		exeFn = os.Executable
	}
	dest, err := exeFn()
	if err != nil {
		return err
	}
	dest, err = filepath.EvalSymlinks(dest)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.Stdout, "downloading %s → %s\n", cand.DownloadURL, dest)
	if err := update.Download(client.HTTP, cand, dest); err != nil {
		return err
	}
	fmt.Fprintf(a.Stdout, "updated (%s verified against %s)\n", cand.AssetName, update.ChecksumsName)
	return nil
}

func (a *App) resolvePath(fileFlag string) (string, error) {
	path := fileFlag
	if path == "" {
		path = hensu.DefaultPath()
	}
	path = filepath.Clean(path)
	if filepath.IsAbs(path) {
		return path, nil
	}
	getwd := a.Getwd
	if getwd == nil {
		getwd = os.Getwd
	}
	cwd, err := getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, path), nil
}

// Options are the context flags parsed before the command keyword.
type Options struct {
	File string
	// Reveal is the mode named by -r; RevealFlag says whether it was given,
	// so the flag can beat HENSU_REVEAL in either direction.
	Reveal     hensu.Reveal
	RevealFlag bool
}

// ParseGlobalFlags consumes context flags before the command keyword.
func ParseGlobalFlags(args []string) (opts Options, rest []string, err error) {
	setReveal := func(name string) error {
		mode, err := command.ParseReveal(name)
		if err != nil {
			return err
		}
		opts.Reveal, opts.RevealFlag = mode, true
		return nil
	}
	for len(args) > 0 {
		a := args[0]
		switch {
		case a == "-h" || a == "--help":
			return opts, []string{"help"}, nil
		case a == "--version":
			return opts, []string{"version"}, nil
		case a == "--update":
			return opts, []string{"--update"}, nil
		case a == "--update-check":
			return opts, []string{"--update-check"}, nil
		case a == "-r" || a == "--reveal":
			if len(args) < 2 {
				return opts, nil, &hensu.InvalidArgument{Msg: a + " needs a mode (peek, mask, or trust)"}
			}
			if err := setReveal(args[1]); err != nil {
				return opts, nil, err
			}
			args = args[2:]
		case strings.HasPrefix(a, "--reveal="):
			if err := setReveal(strings.TrimPrefix(a, "--reveal=")); err != nil {
				return opts, nil, err
			}
			args = args[1:]
		case strings.HasPrefix(a, "-r=") && len(a) > 3:
			if err := setReveal(a[3:]); err != nil {
				return opts, nil, err
			}
			args = args[1:]
		case a == "-f" || a == "--file":
			if len(args) < 2 {
				return opts, nil, &hensu.InvalidArgument{Msg: a + " needs a path"}
			}
			opts.File = strings.TrimSpace(args[1])
			if opts.File == "" {
				return opts, nil, &hensu.InvalidArgument{Msg: a + " needs a path"}
			}
			args = args[2:]
		case strings.HasPrefix(a, "--file="):
			opts.File = strings.TrimSpace(strings.TrimPrefix(a, "--file="))
			if opts.File == "" {
				return opts, nil, &hensu.InvalidArgument{Msg: "--file= needs a path"}
			}
			args = args[1:]
		case strings.HasPrefix(a, "-f=") && len(a) > 3:
			opts.File = strings.TrimSpace(a[3:])
			if opts.File == "" {
				return opts, nil, &hensu.InvalidArgument{Msg: "-f= needs a path"}
			}
			args = args[1:]
		default:
			return opts, args, nil
		}
	}
	return opts, args, nil
}

// reveal resolves the mode for this invocation: an explicit -r wins outright,
// otherwise HENSU_REVEAL decides, otherwise peek.
func (a *App) reveal(opts Options) (hensu.Reveal, error) {
	if opts.RevealFlag {
		return opts.Reveal, nil
	}
	return command.AmbientReveal(a.LookupEnv)
}

// Usage writes CLI help to w.
func Usage(w io.Writer) {
	fmt.Fprintln(w, `usage: hensu [-f path] [-r mode] <command>

context (before the keyword):
  -f, --file <path>      config file (default: ./.env)
  -r, --reveal <mode>    peek | mask | trust   (default peek, env: HENSU_REVEAL)
  --version              print the version
  --update               download the latest release binary (checksum verified)
  --update-check         report whether an update is available

commands:
  get KEY [KEY...]       those keys, as JSON
  keys                   key names, as JSON
  read KEY               exact bytes of one value — requires -r trust
  set KEY VALUE          set one key
  set --json '{...}'     set many keys from a JSON object
  set --json -           set many keys from JSON on stdin
  exec [--] CMD [ARGS]   run CMD with the config in its environment

reveal modes:
  peek    long values show a tip (abcd***wxyz), short ones "***"
  mask    every defined value is "***"
  trust   raw — a grant, asked for explicitly

There is no bulk read: to see a value, name its key. To give a value to a
program without seeing it, use exec.

keys: ':' ≡ '__'. formats: # FORMAT: DOTENV|SHEXPORT|YAML|JSON`)
}
