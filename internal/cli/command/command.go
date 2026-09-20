package command

import (
	"context"
	"io"

	"github.com/my-rv/hensu"
)

// Context carries per-invocation dependencies for a Command.
type Context struct {
	Ctx    context.Context
	Store  *hensu.Store
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
	// Reveal is the ambient reveal mode: the floor every value-printing
	// command starts from. See reveal.go.
	Reveal hensu.Reveal
	// Runner spawns the child process for exec (nil → OSRunner).
	Runner Runner
	// Env returns the inherited environment for exec (nil → os.Environ).
	Env func() []string
}

// Req returns ctx.Ctx or background.
func (c *Context) Req() context.Context {
	if c == nil || c.Ctx == nil {
		return context.Background()
	}
	return c.Ctx
}

// Command is one CLI verb.
type Command interface {
	Name() string
	Run(ctx *Context, args []string) error
}

// Func adapts a function to Command.
type Func struct {
	name string
	fn   func(ctx *Context, args []string) error
}

// NewFunc builds a Command from name + function.
func NewFunc(name string, fn func(ctx *Context, args []string) error) *Func {
	return &Func{name: name, fn: fn}
}

// Name implements Command.
func (f *Func) Name() string { return f.name }

// Run implements Command.
func (f *Func) Run(ctx *Context, args []string) error {
	return f.fn(ctx, args)
}
