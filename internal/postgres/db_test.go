package postgres

import "testing"

func TestNormalizeDatabaseURLConvertsPrismaSchema(t *testing.T) {
	got := normalizeDatabaseURL("postgresql://seedly:seedly@localhost:55432/seedly?schema=public&sslmode=disable")
	want := "postgresql://seedly:seedly@localhost:55432/seedly?search_path=public&sslmode=disable"
	if got != want {
		t.Fatalf("url = %q, want %q", got, want)
	}
}

func TestNormalizeDatabaseURLKeepsExistingSearchPath(t *testing.T) {
	got := normalizeDatabaseURL("postgresql://localhost/db?schema=public&search_path=custom")
	want := "postgresql://localhost/db?search_path=custom"
	if got != want {
		t.Fatalf("url = %q, want %q", got, want)
	}
}
