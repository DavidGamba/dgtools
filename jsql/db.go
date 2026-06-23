package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/DavidGamba/dgtools/clitable"
)

func dbConn(ctx context.Context) (*sql.Conn, error) {
	db, err := sql.Open("duckdb", DBNAME)
	if err != nil {
		return nil, fmt.Errorf("failed: %w", err)
	}
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed: %w", err)
	}
	return conn, nil
}

func runQuery(ctx context.Context, w io.Writer, conn *sql.Conn, mode outputMode, query string) error {
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed: %v\n", err)
	}
	cols, err := rows.Columns()
	if err != nil {
		_ = rows.Close()
		return fmt.Errorf("failed: %w", err)
	}

	var results []map[string]any
	keys := map[string]struct{}{}
	rowCount := 0
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			_ = rows.Close()
			return fmt.Errorf("failed: %w", err)
		}
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			val := values[i]
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			row[col] = val
			keys[col] = struct{}{}
		}
		results = append(results, row)
		rowCount++
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return fmt.Errorf("failed: %w", err)
	}
	_ = rows.Close()

	switch mode {
	case outputModePretty:
		out, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Fprintf(w, "%s\n", string(out))
	case outputModeSingleLine:
		for _, row := range results {
			out, err := json.Marshal(row)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Fprintf(w, "%s\n", string(out))
		}
	case outputModeTable:
		clitable.NewTablePrinter().Fprint(w, clitable.MapTable{MapList: results})
	default:
		return fmt.Errorf("unknown output mode: %q", mode)
	}
	fmt.Fprintf(os.Stderr, "query rows: %d\n", rowCount)
	return nil
}
