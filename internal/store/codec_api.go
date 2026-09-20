package store

// FormatCodec converts a file body (no FORMAT header) ↔ flat map.
type FormatCodec interface {
	Decode(body string) (map[string]string, error)
	Apply(body string, updates map[string]string) (string, error)
}

// CodecRegistry resolves FORMAT name → FormatCodec.
type CodecRegistry interface {
	Lookup(name string) (FormatCodec, error)
}
