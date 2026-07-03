package rnd

import (
	"crypto/rand"
	"math/big"
)

type stringConfig struct {
	charset string
}

func newStringConfig() *stringConfig {
	return &stringConfig{
		charset: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
	}
}

type StringOption func(*stringConfig)

func WithCustomCharset(charset string) StringOption {
	return func(sc *stringConfig) {
		if len(charset) == 0 {
			return
		}

		sc.charset = charset
	}
}

func String(n uint, opts ...StringOption) (string, error) {
	cfg := newStringConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	b := make([]byte, n)
	for i := range b {
		// math/big is used to ensure the generated number is within the charset bounds.
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(cfg.charset))))
		if err != nil {
			return "", err
		}
		b[i] = cfg.charset[num.Int64()]
	}

	return string(b), nil
}
