package store

import (
	"fmt"

	"github.com/my-rv/hensu/internal/codec"
)

type codecRegistryAdapter struct {
	reg *codec.Registry
}

func (a codecRegistryAdapter) Lookup(name string) (FormatCodec, error) {
	c, err := a.reg.Lookup(name)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func defaultCodecRegistry() CodecRegistry {
	return codecRegistryAdapter{reg: codec.SharedRegistry()}
}

func requireRegistry(reg CodecRegistry) CodecRegistry {
	if reg == nil {
		return defaultCodecRegistry()
	}
	return reg
}

var _ CodecRegistry = codecRegistryAdapter{}

func lookupCodec(reg CodecRegistry, f Format) (FormatCodec, error) {
	if reg == nil {
		return nil, fmt.Errorf("nil codec registry")
	}
	return reg.Lookup(string(f))
}
