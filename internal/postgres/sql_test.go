package postgres

import (
	"reflect"
	"testing"
)

func TestBuildInsertSQL(t *testing.T) {
	query, args := BuildInsertSQL("User", []map[string]any{
		{"email": "a@example.com", "name": "A"},
		{"email": "b@example.com", "name": "B"},
	}, []string{"id"})

	wantQuery := `INSERT INTO "User" ("email", "name") VALUES ($1, $2), ($3, $4) RETURNING "id"`
	if query != wantQuery {
		t.Fatalf("query = %q, want %q", query, wantQuery)
	}
	wantArgs := []any{"a@example.com", "A", "b@example.com", "B"}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args = %#v, want %#v", args, wantArgs)
	}
}

func TestBuildInsertSQLDefaultValues(t *testing.T) {
	query, args := BuildInsertSQL("Token", []map[string]any{{}}, []string{"id"})
	if query != `INSERT INTO "Token" DEFAULT VALUES RETURNING "id"` {
		t.Fatalf("query = %q", query)
	}
	if len(args) != 0 {
		t.Fatalf("args = %#v, want none", args)
	}
}
