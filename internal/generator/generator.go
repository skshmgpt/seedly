package generator

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/skshmgpt/seedly/internal/prisma"
)

type Generator struct {
	faker    *gofakeit.Faker
	rng      *rand.Rand
	baseTime time.Time
}

func New(seed int64) *Generator {
	baseTime := time.Unix(1700000000, 0).UTC()
	if seed == 0 {
		seed = time.Now().UnixNano()
		baseTime = time.Now().UTC()
	}
	return &Generator{
		faker:    gofakeit.New(uint64(seed)),
		rng:      rand.New(rand.NewSource(seed)),
		baseTime: baseTime,
	}
}

func (g *Generator) Rows(model prisma.Model, count int, parents map[string][]map[string]any) []map[string]any {
	rows := make([]map[string]any, count)
	for i := range rows {
		row := map[string]any{}
		for _, field := range model.Fields {
			if field.Relation != nil || field.IsList || !field.IsScalar() || shouldOmit(field) {
				continue
			}
			row[field.Name] = g.value(field, i)
		}
		rows[i] = row
	}
	g.assignForeignKeys(model, rows, parents)
	return rows
}

func shouldOmit(field prisma.Field) bool {
	return field.HasDefault
}

func (g *Generator) assignForeignKeys(model prisma.Model, rows []map[string]any, parents map[string][]map[string]any) {
	for _, field := range model.Fields {
		if field.Relation == nil {
			continue
		}
		parentRows := parents[field.Relation.TargetModel]
		if len(parentRows) == 0 {
			continue
		}
		for i, row := range rows {
			parent := parentRows[i%len(parentRows)]
			for idx, localField := range field.Relation.Fields {
				if idx >= len(field.Relation.References) {
					continue
				}
				row[localField] = parent[field.Relation.References[idx]]
			}
		}
	}
}

func (g *Generator) value(field prisma.Field, index int) any {
	name := strings.ToLower(field.Name)
	switch {
	case strings.Contains(name, "email"):
		return g.faker.Email()
	case strings.Contains(name, "name"):
		return g.faker.Name()
	case strings.Contains(name, "title"):
		return g.faker.Sentence(4)
	case strings.HasPrefix(name, "is") || strings.HasPrefix(name, "has"):
		return g.faker.Bool()
	case strings.HasSuffix(name, "count"):
		return g.faker.IntRange(0, 100)
	}

	switch field.Type {
	case "String":
		return g.faker.Word()
	case "Int":
		return g.faker.IntRange(0, 100000)
	case "BigInt":
		return int64(g.faker.IntRange(0, 1000000))
	case "Float", "Decimal":
		return g.faker.Float64Range(0, 1000)
	case "Boolean":
		return g.faker.Bool()
	case "DateTime":
		return g.faker.DateRange(g.baseTime.Add(-30*24*time.Hour), g.baseTime)
	case "Json":
		return fmt.Sprintf(`{"value":%d}`, g.rng.Intn(100000))
	case "Bytes":
		buf := make([]byte, 8)
		_, _ = g.rng.Read(buf)
		return base64.StdEncoding.EncodeToString(buf)
	default:
		return nil
	}
}
