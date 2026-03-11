package controllers

import (
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func SetupPostHooks(app *pocketbase.PocketBase) {
	app.OnRecordUpdate("posts").BindFunc(func(e *core.RecordEvent) error {
		deletedAt := strings.TrimSpace(e.Record.GetString("deleted"))
		status := strings.ToLower(strings.TrimSpace(e.Record.GetString("status")))

		// Soft deleted posts should not keep media files in storage.
		if deletedAt != "" || status == "deleted" {
			images := e.Record.GetStringSlice("images")
			if len(images) > 0 {
				e.Record.Set("images", []string{})
				app.Logger().Info("Post media cleared on soft delete", "post", e.Record.Id, "files", len(images))
			}
		}

		return e.Next()
	})
}

