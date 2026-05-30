package cli

type Config struct {
	SchemaPath     string
	Count          int
	Seed           int64
	DatabaseURL    string
	EnvDatabaseURL string
}

func (c Config) EffectiveDatabaseURL() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	return c.EnvDatabaseURL
}
