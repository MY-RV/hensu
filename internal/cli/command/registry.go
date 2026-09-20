package command

import "fmt"

// Registry holds named Commands.
type Registry struct {
	byName map[string]Command
}

// NewRegistry returns an empty command registry.
func NewRegistry() *Registry {
	return &Registry{byName: make(map[string]Command)}
}

// Register adds a command. Panics on duplicate names (programmer error).
func (r *Registry) Register(cmd Command) {
	if cmd == nil {
		panic("nil command")
	}
	name := cmd.Name()
	if name == "" {
		panic("empty command name")
	}
	if _, exists := r.byName[name]; exists {
		panic(fmt.Sprintf("duplicate command %q", name))
	}
	r.byName[name] = cmd
}

// Lookup finds a command by name.
func (r *Registry) Lookup(name string) (Command, bool) {
	cmd, ok := r.byName[name]
	return cmd, ok
}
