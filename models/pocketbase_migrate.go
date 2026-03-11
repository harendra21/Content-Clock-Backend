package models

import (
	"fmt"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func MigrateCollectionsIfEnabled(app *pocketbase.PocketBase) error {
	if !isMigrationEnabled() {
		return nil
	}

	if err := ensureCollection(app, "connections", ApplyConnectionsCollectionSchema); err != nil {
		return err
	}
	if err := ensureCollection(app, "posts", ApplyPostsCollectionSchema); err != nil {
		return err
	}
	if err := ensureCollection(app, "notifications", ApplyNotificationsCollectionSchema); err != nil {
		return err
	}
	if err := ensureCollection(app, "analytics", ApplyAnalyticsCollectionSchema); err != nil {
		return err
	}
	return nil
}

func isMigrationEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("DB_MIGRATE")))
	return v == "1" || v == "true" || v == "yes"
}

func ensureCollection(app *pocketbase.PocketBase, name string, configure func(c *core.Collection)) error {
	collection, err := app.FindCollectionByNameOrId(name)
	if err != nil {
		collection = core.NewBaseCollection(name)
		configure(collection)
		if err := app.Save(collection); err != nil {
			return fmt.Errorf("failed to create %s collection: %w", name, err)
		}
		app.Logger().Info("Collection initialized", "name", name)
		return nil
	}

	configure(collection)
	if err := app.Save(collection); err != nil {
		return fmt.Errorf("failed to update %s collection: %w", name, err)
	}
	app.Logger().Info("Collection ensured", "name", name)
	return nil
}
