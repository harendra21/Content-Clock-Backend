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
	var isConnection ConnectionResult
	query := fmt.Sprintf("SELECT id FROM connections WHERE connection_id = '%s'", connection.ConnectionId)
	err := app.DB().NewQuery(query).One(&isConnection)
	if err != nil && err.Error() != "sql: no rows in result set" {
		app.Logger().Error("Error checking for existing connection: %s", err.Error())
		return err
	}

	if isConnection.ID != "" {
		app.Logger().Info("Connection already exists with ID: %s", isConnection.ID)
		return nil
	}

	result, err := app.DB().Insert("connections", dbx.Params{
		"name":            connection.Name,
		"username":        connection.Username,
		"connection_name": connection.ConnectionName,
		"connection_id":   connection.ConnectionId,
		"access_token":    connection.AccessToken,
		"refresh_token":   connection.RefreshToken,
		"meta_data":       connection.MetaData,
		"timezone":        connection.Timezone,
		"user":            connection.UserId,
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
			app.Logger().Error("Error in downloading images", "Error", err.Error())
			return err
		}
	}

	imageFile, err := filesystem.NewFileFromPath(image)
	if err != nil {
		app.Logger().Error("Error in downloading images", "Error", err.Error())
		return err
	}

	getConnection := fmt.Sprintf("SELECT id FROM connections WHERE connection_id = '%s'", connection.ConnectionId)
	err = app.DB().NewQuery(getConnection).One(&isConnection)
	if err != nil {
		app.Logger().Error("Error in downloading images", "Error", err.Error())
		return err
	}

	record, err := app.FindRecordById("connections", isConnection.ID)
	if err != nil {
		app.Logger().Error("Error in downloading images", "Error", err.Error())
		return err
	}

	record.Set("profile_image", imageFile)
	err = app.Save(record)
	if err != nil {
		app.Logger().Error("Error in downloading images", "Error", err.Error())
		return err
	}

	return nil

}
