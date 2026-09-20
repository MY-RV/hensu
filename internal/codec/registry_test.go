package codec_test

import (
	"testing"

	"github.com/my-rv/hensu/internal/codec"
)

func TestRegistry_lookupBuiltin(t *testing.T) {
	r := codec.DefaultRegistry()
	for _, name := range []string{"DOTENV", "SHEXPORT", "YAML", "JSON"} {
		if _, err := r.Lookup(name); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	if _, err := r.Lookup("XML"); err == nil {
		t.Fatal("expected unknown FORMAT")
	}
}
