package codec

import (
	"fmt"
	"strings"
	"sync"
)

// Registry maps FORMAT names to Codec implementations.
type Registry struct {
	mu     sync.RWMutex
	codecs map[string]Codec
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{codecs: make(map[string]Codec)}
}

// DefaultRegistry returns a registry with built-in formats.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	_ = r.Register("DOTENV", Dotenv{})
	_ = r.Register("SHEXPORT", Shexport{})
	_ = r.Register("YAML", YAML{})
	_ = r.Register("JSON", JSON{})
	return r
}

var defaultRegistry = DefaultRegistry()

// SharedRegistry is the process-wide default (used by Store unless overridden).
func SharedRegistry() *Registry {
	return defaultRegistry
}

// Register adds or replaces a Codec for name (uppercased).
func (r *Registry) Register(name string, c Codec) error {
	if r == nil {
		return fmt.Errorf("nil codec registry")
	}
	if c == nil {
		return fmt.Errorf("nil codec")
	}
	name = normalizeFormatName(name)
	if name == "" {
		return fmt.Errorf("empty format name")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.codecs[name] = c
	return nil
}

// Lookup returns the Codec for name.
func (r *Registry) Lookup(name string) (Codec, error) {
	if r == nil {
		return nil, fmt.Errorf("nil codec registry")
	}
	name = normalizeFormatName(name)
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.codecs[name]
	if !ok {
		return nil, fmt.Errorf("unknown FORMAT %q", name)
	}
	return c, nil
}

// Names returns registered FORMAT names (unsorted).
func (r *Registry) Names() []string {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.codecs))
	for k := range r.codecs {
		out = append(out, k)
	}
	return out
}

func normalizeFormatName(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}
