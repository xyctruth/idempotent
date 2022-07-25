package idempotent

import "time"

func DefaultConfig() *Config {
	return &Config{
		TTL:                 time.Hour * 60 * 24,
		ClearExpiryDuration: time.Hour * 1,
	}
}

type Config struct {
	// key过期时间
	TTL time.Duration
	//清除过期间隔
	ClearExpiryDuration time.Duration
}

type ConfigOpt func(op *Config)

// WithTTL
func WithTTL(ttl time.Duration) ConfigOpt {
	return ConfigOpt(func(op *Config) {
		op.TTL = ttl
	})
}

// WithClearExpiry
func WithClearExpiry(duration time.Duration) ConfigOpt {
	return ConfigOpt(func(op *Config) {
		op.ClearExpiryDuration = duration
	})
}
