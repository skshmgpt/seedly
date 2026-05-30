package generator

import (
	"reflect"
	"testing"

	"github.com/skshmgpt/seedly/internal/prisma"
)

func TestGenerateRowsIsDeterministicWithSeed(t *testing.T) {
	model := prisma.Model{Name: "User", Fields: []prisma.Field{
		{Name: "id", Type: "Int", IsID: true, HasDefault: true},
		{Name: "email", Type: "String"},
		{Name: "name", Type: "String"},
	}}

	first := New(42).Rows(model, 2, nil)
	second := New(42).Rows(model, 2, nil)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("rows differ with same seed:\n%#v\n%#v", first, second)
	}
	if _, ok := first[0]["id"]; ok {
		t.Fatal("defaulted id should be omitted")
	}
	if _, ok := first[0]["email"].(string); !ok {
		t.Fatal("expected generated email string")
	}
}

func TestAssignsForeignKeysRoundRobin(t *testing.T) {
	model := prisma.Model{Name: "Order", Fields: []prisma.Field{
		{Name: "userId", Type: "Int"},
		{Name: "user", Type: "User", Relation: &prisma.Relation{TargetModel: "User", Fields: []string{"userId"}, References: []string{"id"}}},
	}}
	parents := map[string][]map[string]any{"User": {{"id": 10}, {"id": 20}}}

	rows := New(1).Rows(model, 3, parents)
	got := []any{rows[0]["userId"], rows[1]["userId"], rows[2]["userId"]}
	want := []any{10, 20, 10}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fk values = %#v, want %#v", got, want)
	}
}
