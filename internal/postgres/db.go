package postgres

import (
	"context"
	"database/sql"
	"net/url"
)

func Open(ctx context.Context, databaseURL string) (*sql.DB, error) {
	databaseURL = normalizeDatabaseURL(databaseURL)
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func normalizeDatabaseURL(databaseURL string) string {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return databaseURL
	}
	query := parsed.Query()
	schema := query.Get("schema")
	if schema == "" {
		return databaseURL
	}

	query.Del("schema")
	if query.Get("search_path") == "" {
		query.Set("search_path", schema)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
