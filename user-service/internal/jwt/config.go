package jwt

import "time"

type Config struct {
	Issuer     string        `env:"JWT_ISSUER"      env-default:"jwt.token-manager" yaml:"issuer"`
	Secret     string        `env:"JWT_SECRET"                                      yaml:"secret"      env-required:"true"`
	AccessTTL  time.Duration `env:"JWT_ACCESS_TTL"  env-default:"15m"               yaml:"access_ttl"`
	RefreshTTL time.Duration `env:"JWT_REFRESH_TTL" env-default:"7d"                yaml:"refresh_ttl"`
}
