package cli

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"os"

	_ "github.com/lib/pq"

	"github.com/skshmgpt/seedly/internal/postgres"
	"github.com/skshmgpt/seedly/internal/prisma"
	"github.com/skshmgpt/seedly/internal/seed"
	"github.com/skshmgpt/seedly/internal/ui"
)

func Run(ctx context.Context, args []string, in io.Reader, out io.Writer, errOut io.Writer) error {
	if len(args) == 0 || args[0] != "seed" {
		return fmt.Errorf("usage: seedly seed --schema <path> [--count 10] [--seed n] [--database-url url]")
	}

	flags := flag.NewFlagSet("seed", flag.ContinueOnError)
	flags.SetOutput(errOut)
	cfg := Config{Count: 10, EnvDatabaseURL: os.Getenv("DATABASE_URL")}
	flags.StringVar(&cfg.SchemaPath, "schema", "", "path to schema.prisma")
	flags.IntVar(&cfg.Count, "count", 10, "rows to generate per model")
	flags.Int64Var(&cfg.Seed, "seed", 0, "deterministic generation seed")
	flags.StringVar(&cfg.DatabaseURL, "database-url", "", "Postgres database URL override")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	return runSeed(ctx, cfg, in, out)
}

func runSeed(ctx context.Context, cfg Config, in io.Reader, out io.Writer) error {
	if cfg.SchemaPath == "" {
		return fmt.Errorf("--schema is required")
	}
	if cfg.Count <= 0 {
		return fmt.Errorf("--count must be greater than zero")
	}
	databaseURL := cfg.EffectiveDatabaseURL()
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL or --database-url is required")
	}

	contents, err := os.ReadFile(cfg.SchemaPath)
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	schema, err := prisma.ParseString(string(contents))
	if err != nil {
		return fmt.Errorf("parse schema: %w", err)
	}
	selected, err := ui.SelectModels(in, out, schema.Models)
	if err != nil {
		return err
	}

	db, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()

	return runInTransaction(ctx, db, schema, selected, cfg, out)
}

func runInTransaction(ctx context.Context, db *sql.DB, schema prisma.Schema, selected []string, cfg Config, out io.Writer) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	engine := seed.Engine{Inserter: postgres.Inserter{Tx: tx}}
	result, err := engine.Run(ctx, seed.Request{Schema: schema, SelectedModels: selected, Count: cfg.Count, Seed: cfg.Seed})
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true

	_, _ = fmt.Fprintf(out, "Seeded models in order: %v\n", result.Order)
	for _, skip := range result.Skipped {
		_, _ = fmt.Fprintf(out, "Skipped %s: %s\n", skip.Model, skip.Reason)
	}
	for model, count := range result.Inserted {
		_, _ = fmt.Fprintf(out, "Inserted %d %s rows\n", count, model)
	}
	return nil
}
