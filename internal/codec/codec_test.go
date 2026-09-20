package codec_test

import (
	"testing"

	"github.com/my-rv/hensu/internal/codec"
)

func TestParseFormatDirective(t *testing.T) {
	tests := []struct {
		line    string
		want    string
		wantOK  bool
		wantErr bool
	}{
		{"# FORMAT: YAML", "YAML", true, false},
		{"# format: dotenv", "DOTENV", true, false},
		{"# FORMAT: SHEXPORT", "SHEXPORT", true, false},
		{"# FORMAT: JSON", "JSON", true, false},
		{"# Copy to .env", "", false, false},
		{"# FORMAT: XML", "", false, true},
	}
	for _, tc := range tests {
		got, ok, err := codec.ParseFormatDirective(tc.line)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("%q: expected error", tc.line)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%q: %v", tc.line, err)
		}
		if ok != tc.wantOK || got != tc.want {
			t.Fatalf("%q: got (%q, %v)", tc.line, got, ok)
		}
	}
}

func TestDotenvDecode_spacesComments(t *testing.T) {
	c := codec.Dotenv{}
	m, err := c.Decode(`
export ANON_KEY='abc'
PUBLISHABLE_KEY="pub"
FOO=plain
export ENV= "asd" # A random comment
URL=http://host#frag
U=http://host # tail
`)
	if err != nil {
		t.Fatal(err)
	}
	checks := map[string]string{
		"ANON_KEY":        "abc",
		"PUBLISHABLE_KEY": "pub",
		"FOO":             "plain",
		"ENV":             "asd",
		"URL":             "http://host#frag",
		"U":               "http://host",
	}
	for k, want := range checks {
		if m[k] != want {
			t.Fatalf("%s = %q want %q", k, m[k], want)
		}
	}
}

func TestJSONDecode_flat(t *testing.T) {
	c := codec.JSON{}
	m, err := c.Decode(`{"FOO":"bar","N":2}`)
	if err != nil {
		t.Fatal(err)
	}
	if m["FOO"] != "bar" || m["N"] != "2" {
		t.Fatalf("%#v", m)
	}
}
