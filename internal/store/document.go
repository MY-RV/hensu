package store

import (
	"fmt"

	"github.com/my-rv/hensu/internal/codec"
)

// Document is a config file split into optional FORMAT header + body.
type Document struct {
	Format    Format
	Body      string
	HadHeader bool
}

// ParseDocument splits raw file bytes/string into FORMAT + body.
// Empty input → DOTENV with empty body. No header → DOTENV and full raw as body.
func ParseDocument(raw string) (Document, error) {
	if raw == "" {
		return Document{Format: FormatDotenv, Body: "", HadHeader: false}, nil
	}
	format, body, err := codec.SplitHeader(raw)
	if err != nil {
		return Document{}, &ParseError{Msg: "FORMAT header", Err: err}
	}
	f := Format(format)
	if !f.Valid() {
		return Document{}, &ParseError{Msg: fmt.Sprintf("unknown FORMAT %q", format)}
	}
	return Document{
		Format:    f,
		Body:      body,
		HadHeader: codec.FirstLineIsFormatDirective(raw),
	}, nil
}

// SerializeDocument rebuilds file contents from a Document and body text.
// Writes a FORMAT header when HadHeader is set, or when the format is not
// plain DOTENV (YAML / SHEXPORT / JSON always carry an explicit header).
func SerializeDocument(doc Document, body string) string {
	needHeader := doc.HadHeader ||
		doc.Format == FormatYAML ||
		doc.Format == FormatShexport ||
		doc.Format == FormatJSON
	var out string
	if needHeader {
		out = fmt.Sprintf("# FORMAT: %s\n", doc.Format)
	}
	out += body
	if body != "" && body[len(body)-1] != '\n' {
		out += "\n"
	}
	return out
}
