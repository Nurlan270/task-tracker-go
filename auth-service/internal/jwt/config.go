package jwt

import "time"

type Config struct {
	Secret     string        `env:"JWT_SECRET"      env-required:"true" yaml:"secret"`
	AccessTTL  time.Duration `env:"JWT_ACCESS_TTL"                      yaml:"access_ttl"  env-default:"15m"`
	RefreshTTL time.Duration `env:"JWT_REFRESH_TTL"                     yaml:"refresh_ttl" env-default:"7d"`
}
