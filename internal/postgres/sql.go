package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/skshmgpt/seedly/internal/prisma"
)

type Inserter struct {
	Tx *sql.Tx
}

func (i Inserter) Insert(ctx context.Context, model prisma.Model, rows []map[string]any) ([]map[string]any, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	returning := primaryKeyNames(model)
	query, args := BuildInsertSQL(model.Name, rows, returning)

	if len(returning) == 0 {
		_, err := i.Tx.ExecContext(ctx, query, args...)
		return rows, err
	}

	resultRows, err := i.Tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer resultRows.Close()

	inserted := make([]map[string]any, 0, len(rows))
	for resultRows.Next() {
		values := make([]any, len(returning))
		scan := make([]any, len(returning))
		for idx := range values {
			scan[idx] = &values[idx]
		}
		if err := resultRows.Scan(scan...); err != nil {
			return nil, err
		}
		row := map[string]any{}
		for key, value := range rows[len(inserted)] {
			row[key] = value
		}
		for idx, column := range returning {
			row[column] = values[idx]
		}
		inserted = append(inserted, row)
	}
	if err := resultRows.Err(); err != nil {
		return nil, err
	}
	return inserted, nil
}

func BuildInsertSQL(table string, rows []map[string]any, returning []string) (string, []any) {
	columns := sortedColumns(rows)
	if len(columns) == 0 {
		query := fmt.Sprintf("INSERT INTO %s DEFAULT VALUES", quoteIdent(table))
		if len(returning) > 0 {
			query += " RETURNING " + quoteList(returning)
		}
		return query, nil
	}

	args := make([]any, 0, len(rows)*len(columns))
	values := make([]string, 0, len(rows))
	placeholder := 1
	for _, row := range rows {
		parts := make([]string, len(columns))
		for idx, column := range columns {
			parts[idx] = fmt.Sprintf("$%d", placeholder)
			placeholder++
			args = append(args, row[column])
		}
		values = append(values, "("+strings.Join(parts, ", ")+")")
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", quoteIdent(table), quoteList(columns), strings.Join(values, ", "))
	if len(returning) > 0 {
		query += " RETURNING " + quoteList(returning)
	}
	return query, args
}

func sortedColumns(rows []map[string]any) []string {
	seen := map[string]bool{}
	for _, row := range rows {
		for column := range row {
			seen[column] = true
		}
	}
	columns := make([]string, 0, len(seen))
	for column := range seen {
		columns = append(columns, column)
	}
	sort.Strings(columns)
	return columns
}

func primaryKeyNames(model prisma.Model) []string {
	keys := model.PrimaryKeyFields()
	names := make([]string, 0, len(keys))
	for _, key := range keys {
		names = append(names, key.Name)
	}
	return names
}

func quoteList(values []string) string {
	quoted := make([]string, len(values))
	for idx, value := range values {
		quoted[idx] = quoteIdent(value)
	}
	return strings.Join(quoted, ", ")
}

func quoteIdent(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}
