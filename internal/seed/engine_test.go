package seed

import (
	"context"
	"testing"

	"github.com/skshmgpt/seedly/internal/prisma"
)

func TestEngineSeedsInPlanOrder(t *testing.T) {
	schema := prisma.Schema{Models: []prisma.Model{
		{Name: "User", Fields: []prisma.Field{{Name: "id", Type: "Int", IsID: true, HasDefault: true}, {Name: "email", Type: "String"}}},
		{Name: "Order", Fields: []prisma.Field{{Name: "id", Type: "Int", IsID: true, HasDefault: true}, {Name: "userId", Type: "Int"}, {Name: "user", Type: "User", Relation: &prisma.Relation{TargetModel: "User", Fields: []string{"userId"}, References: []string{"id"}}}}},
	}}
	inserter := &fakeInserter{}
	engine := Engine{Inserter: inserter}

	result, err := engine.Run(context.Background(), Request{Schema: schema, SelectedModels: []string{"Order"}, Count: 2, Seed: 7})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(inserter.models) != 2 || inserter.models[0] != "User" || inserter.models[1] != "Order" {
		t.Fatalf("insert order = %#v", inserter.models)
	}
	if result.Inserted["User"] != 2 || result.Inserted["Order"] != 2 {
		t.Fatalf("inserted = %#v", result.Inserted)
	}
}

type fakeInserter struct {
	models []string
}

func (f *fakeInserter) Insert(ctx context.Context, model prisma.Model, rows []map[string]any) ([]map[string]any, error) {
	f.models = append(f.models, model.Name)
	inserted := make([]map[string]any, len(rows))
	for i, row := range rows {
		copy := map[string]any{"id": i + 1}
		for key, value := range row {
			copy[key] = value
		}
		inserted[i] = copy
	}
	return inserted, nil
}
