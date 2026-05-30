package graph

import (
	"fmt"
	"sort"

	"github.com/skshmgpt/seedly/internal/prisma"
)

type Plan struct {
	Order    []string
	Included map[string]bool
	Skipped  []Skip
}

type Skip struct {
	Model  string
	Reason string
}

func BuildPlan(schema prisma.Schema, selected []string) Plan {
	deps := dependencies(schema)
	included := map[string]bool{}
	var skipped []Skip

	var include func(string)
	include = func(name string) {
		if included[name] {
			return
		}
		if _, ok := schema.Model(name); !ok {
			skipped = append(skipped, Skip{Model: name, Reason: "model not found"})
			return
		}
		included[name] = true
		for _, dep := range deps[name] {
			include(dep)
		}
	}

	for _, name := range selected {
		include(name)
	}

	order, cycleModels := topo(included, deps)
	for _, name := range cycleModels {
		delete(included, name)
		skipped = append(skipped, Skip{Model: name, Reason: "dependency cycle"})
	}

	return Plan{Order: order, Included: included, Skipped: skipped}
}

func dependencies(schema prisma.Schema) map[string][]string {
	deps := map[string][]string{}
	for _, model := range schema.Models {
		seen := map[string]bool{}
		for _, field := range model.Fields {
			if field.Relation == nil || field.Relation.TargetModel == "" {
				continue
			}
			if field.Relation.TargetModel == model.Name {
				continue
			}
			if !seen[field.Relation.TargetModel] {
				deps[model.Name] = append(deps[model.Name], field.Relation.TargetModel)
				seen[field.Relation.TargetModel] = true
			}
		}
		sort.Strings(deps[model.Name])
	}
	return deps
}

func topo(included map[string]bool, deps map[string][]string) ([]string, []string) {
	state := map[string]int{}
	var order []string
	cycle := map[string]bool{}

	var visit func(string, []string)
	visit = func(name string, stack []string) {
		if cycle[name] || !included[name] {
			return
		}
		switch state[name] {
		case 2:
			return
		case 1:
			markCycle(cycle, stack, name)
			return
		}

		state[name] = 1
		for _, dep := range deps[name] {
			visit(dep, append(stack, name))
		}
		state[name] = 2
		if !cycle[name] {
			order = append(order, name)
		}
	}

	names := make([]string, 0, len(included))
	for name := range included {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		visit(name, nil)
	}

	for _, name := range order {
		if dependsOnCycle(name, deps, cycle, map[string]bool{}) {
			cycle[name] = true
		}
	}

	filtered := order[:0]
	for _, name := range order {
		if !cycle[name] {
			filtered = append(filtered, name)
		}
	}

	cycleModels := make([]string, 0, len(cycle))
	for name := range cycle {
		cycleModels = append(cycleModels, name)
	}
	sort.Strings(cycleModels)
	return filtered, cycleModels
}

func markCycle(cycle map[string]bool, stack []string, name string) {
	cycle[name] = true
	for i := len(stack) - 1; i >= 0; i-- {
		cycle[stack[i]] = true
		if stack[i] == name {
			return
		}
	}
}

func dependsOnCycle(name string, deps map[string][]string, cycle map[string]bool, seen map[string]bool) bool {
	if seen[name] {
		return false
	}
	seen[name] = true
	for _, dep := range deps[name] {
		if cycle[dep] || dependsOnCycle(dep, deps, cycle, seen) {
			return true
		}
	}
	return false
}

func (p Plan) Validate() error {
	if len(p.Order) == 0 {
		return fmt.Errorf("no resolvable models remain to seed")
	}
	return nil
}
