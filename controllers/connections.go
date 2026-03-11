package controllers

import (
	"content-clock/helpers"
	"content-clock/models"
	"fmt"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

type ConnectionResult struct {
	ID string `db:"id"`
}

func AddNewConnection(app *pocketbase.PocketBase, connection *models.Connections) error {
	if err := EnsureTables(app, "connections"); err != nil {
		app.Logger().Error("Schema check failed", "error", err.Error())
		return fmt.Errorf("connections collection is not initialized. Please create the 'connections' collection in PocketBase first")
	}

	var isConnection ConnectionResult
	checkExistingExp := dbx.NewExp(
		"connection_id = {:connectionId} AND connection_name = {:connectionName} AND user = {:userId} AND coalesce(deleted, '') = ''",
		dbx.Params{
			"connectionId":   connection.ConnectionId,
			"connectionName": connection.ConnectionName,
			"userId":         connection.UserId,
		},
	)
	err := app.DB().Select("id").From("connections").Where(checkExistingExp).One(&isConnection)
	if err != nil && err.Error() != "sql: no rows in result set" {
		app.Logger().Error("Error checking for existing connection", "error", err.Error())
		return err
	}

	if isConnection.ID != "" {
		app.Logger().Info("Connection already exists", "id", isConnection.ID)
		return nil
	}

	result, err := app.DB().Insert("connections", dbx.Params{
		"name":              connection.Name,
		"username":          connection.Username,
		"connection_name":   connection.ConnectionName,
		"connection_id":     connection.ConnectionId,
		"access_token":      connection.AccessToken,
		"refresh_token":     connection.RefreshToken,
		"meta_data":         connection.MetaData,
		"timezone":          connection.Timezone,
		"user":              connection.UserId,
		"profile_image_url": connection.ProfileImage,
	}).Execute()

	if err != nil {
		app.Logger().Error("Error inserting new connection", "error", err.Error())
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		app.Logger().Error("Error getting rows affected", "error", err.Error())
		return err
	}
	if rowsAffected == 0 {
		app.Logger().Error("No rows affected, connection not added")
		return nil
	}
	id, _ := result.LastInsertId()
	app.Logger().Info("Connection added successfully with ID", "Id", id)

	if connection.ProfileImage == "" {
		return nil
	}

	image := ""

	if connection.ProfileImage != "" {
		image, err = helpers.DownloadImage(connection.ProfileImage, true)
		if err != nil {
			app.Logger().Warn("Failed to download profile image; keeping profile_image_url only", "error", err.Error(), "connectionId", connection.ConnectionId)
			return nil
		}
	}

	imageFile, err := filesystem.NewFileFromPath(image)
	if err != nil {
		app.Logger().Warn("Failed to read downloaded profile image; keeping profile_image_url only", "error", err.Error(), "connectionId", connection.ConnectionId)
		return nil
	}

	getCreatedExp := dbx.NewExp(
		"connection_id = {:connectionId} AND connection_name = {:connectionName} AND user = {:userId} AND coalesce(deleted, '') = ''",
		dbx.Params{
			"connectionId":   connection.ConnectionId,
			"connectionName": connection.ConnectionName,
			"userId":         connection.UserId,
		},
	)
	err = app.DB().Select("id").From("connections").Where(getCreatedExp).One(&isConnection)
	if err != nil {
		app.Logger().Warn("Failed to fetch created connection for profile image save", "error", err.Error(), "connectionId", connection.ConnectionId)
		return nil
	}

	record, err := app.FindRecordById("connections", isConnection.ID)
	if err != nil {
		app.Logger().Warn("Failed to find created connection for profile image save", "error", err.Error(), "connectionId", connection.ConnectionId)
		return nil
	}

	record.Set("profile_image", imageFile)
	err = app.Save(record)
	if err != nil {
		app.Logger().Warn("Failed to save profile image file; keeping profile_image_url only", "error", err.Error(), "connectionId", connection.ConnectionId)
		return nil
	}

	return nil

}
