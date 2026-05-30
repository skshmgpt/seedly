package seed

import (
	"context"
	"fmt"

	"github.com/skshmgpt/seedly/internal/generator"
	"github.com/skshmgpt/seedly/internal/graph"
	"github.com/skshmgpt/seedly/internal/prisma"
)

type Inserter interface {
	Insert(ctx context.Context, model prisma.Model, rows []map[string]any) ([]map[string]any, error)
}

type Engine struct {
	Inserter Inserter
}

type Request struct {
	Schema         prisma.Schema
	SelectedModels []string
	Count          int
	Seed           int64
}

type Result struct {
	Inserted map[string]int
	Order    []string
	Skipped  []graph.Skip
}

func (e Engine) Run(ctx context.Context, request Request) (Result, error) {
	if e.Inserter == nil {
		return Result{}, fmt.Errorf("inserter is required")
	}
	if request.Count <= 0 {
		return Result{}, fmt.Errorf("count must be greater than zero")
	}

	plan := graph.BuildPlan(request.Schema, request.SelectedModels)
	if err := plan.Validate(); err != nil {
		return Result{Skipped: plan.Skipped}, err
	}

	gen := generator.New(request.Seed)
	parents := map[string][]map[string]any{}
	result := Result{Inserted: map[string]int{}, Order: plan.Order, Skipped: plan.Skipped}

	for _, modelName := range plan.Order {
		model, ok := request.Schema.Model(modelName)
		if !ok {
			continue
		}
		rows := gen.Rows(model, request.Count, parents)
		inserted, err := e.Inserter.Insert(ctx, model, rows)
		if err != nil {
			return result, fmt.Errorf("insert %s: %w", model.Name, err)
		}
		parents[model.Name] = inserted
		result.Inserted[model.Name] = len(inserted)
	}

	return result, nil
}
