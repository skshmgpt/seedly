package graph

import (
	"reflect"
	"testing"

	"github.com/skshmgpt/seedly/internal/prisma"
)

func TestPlanIncludesDependenciesInOrder(t *testing.T) {
	schema := prisma.Schema{Models: []prisma.Model{
		{Name: "User", Fields: []prisma.Field{{Name: "id", Type: "Int", IsID: true}}},
		{Name: "Order", Fields: []prisma.Field{
			{Name: "userId", Type: "Int"},
			{Name: "user", Type: "User", Relation: &prisma.Relation{TargetModel: "User", Fields: []string{"userId"}, References: []string{"id"}}},
		}},
	}}

	plan := BuildPlan(schema, []string{"Order"})
	if len(plan.Skipped) != 0 {
		t.Fatalf("unexpected skips: %#v", plan.Skipped)
	}
	if !reflect.DeepEqual(plan.Order, []string{"User", "Order"}) {
		t.Fatalf("order = %#v, want User then Order", plan.Order)
	}
}

func TestPlanSkipsCyclesBestEffort(t *testing.T) {
	schema := prisma.Schema{Models: []prisma.Model{
		{Name: "A", Fields: []prisma.Field{{Name: "b", Type: "B", Relation: &prisma.Relation{TargetModel: "B", Fields: []string{"bID"}, References: []string{"id"}}}}},
		{Name: "B", Fields: []prisma.Field{{Name: "a", Type: "A", Relation: &prisma.Relation{TargetModel: "A", Fields: []string{"aID"}, References: []string{"id"}}}}},
		{Name: "C", Fields: []prisma.Field{{Name: "id", Type: "Int"}}},
	}}

	plan := BuildPlan(schema, []string{"A", "C"})
	if !reflect.DeepEqual(plan.Order, []string{"C"}) {
		t.Fatalf("order = %#v, want only C", plan.Order)
	}
	if len(plan.Skipped) == 0 {
		t.Fatal("expected skipped cycle models")
	}
}
