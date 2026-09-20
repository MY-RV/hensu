package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/my-rv/hensu"
)

// RevealEnv names the environment variable that sets the reveal mode for a
// whole session, for callers that should not have to pass a flag every time.
const RevealEnv = "HENSU_REVEAL"

// ParseReveal maps a mode name to a Reveal.
func ParseReveal(name string) (hensu.Reveal, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "peek":
		return hensu.RevealPeek, nil
	case "mask":
		return hensu.RevealMask, nil
	case "trust":
		return hensu.RevealTrust, nil
	default:
		return hensu.RevealPeek, &hensu.InvalidArgument{
			Msg: fmt.Sprintf("unknown reveal mode %q (want peek, mask, or trust)", name),
		}
	}
}

// RevealName is the inverse of ParseReveal.
func RevealName(mode hensu.Reveal) string {
	switch mode {
	case hensu.RevealMask:
		return "mask"
	case hensu.RevealTrust:
		return "trust"
	default:
		return "peek"
	}
}

// AmbientReveal reads the mode from the environment. Unset means peek — the
// default is already safe, so there is no floor to carry around. An invalid
// value is an error rather than a silent fallback.
func AmbientReveal(lookup func(string) (string, bool)) (hensu.Reveal, error) {
	if lookup == nil {
		lookup = os.LookupEnv
	}
	v, ok := lookup(RevealEnv)
	if !ok || strings.TrimSpace(v) == "" {
		return hensu.RevealPeek, nil
	}
	mode, err := ParseReveal(v)
	if err != nil {
		return hensu.RevealPeek, &hensu.InvalidArgument{
			Msg: fmt.Sprintf("%s=%q is not a reveal mode (want peek, mask, or trust)", RevealEnv, v),
		}
	}
	return mode, nil
}
