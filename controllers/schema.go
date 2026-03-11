package controllers

import (
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase"
)

type tableExistsResult struct {
	Count int `db:"count"`
}

// TableExists checks whether a sqlite table exists in the current PocketBase database.
func TableExists(app *pocketbase.PocketBase, tableName string) (bool, error) {
	var result tableExistsResult
	query := `SELECT COUNT(*) as count FROM sqlite_master WHERE type = 'table' AND name = {:table}`
	err := app.DB().NewQuery(query).Bind(map[string]any{
		"table": tableName,
	}).One(&result)
	if err != nil {
		return false, err
	}
	return result.Count > 0, nil
}

// EnsureTables returns an error listing missing tables, if any.
func EnsureTables(app *pocketbase.PocketBase, tables ...string) error {
	missing := make([]string, 0)
	for _, table := range tables {
		exists, err := TableExists(app, table)
		if err != nil {
			return err
		}
		if !exists {
			missing = append(missing, table)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("required table(s) missing: %s", strings.Join(missing, ", "))
	}

	return nil
}
