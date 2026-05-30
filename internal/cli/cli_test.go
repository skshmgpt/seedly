package cli

import "testing"

func TestConfigDatabaseURLPrefersFlag(t *testing.T) {
	cfg := Config{DatabaseURL: "from-flag", EnvDatabaseURL: "from-env"}
	if got := cfg.EffectiveDatabaseURL(); got != "from-flag" {
		t.Fatalf("database url = %q", got)
	}
}

func TestConfigDatabaseURLFallsBackToEnv(t *testing.T) {
	cfg := Config{EnvDatabaseURL: "from-env"}
	if got := cfg.EffectiveDatabaseURL(); got != "from-env" {
		t.Fatalf("database url = %q", got)
	}
}
