package controllers

import (
	"content-clock/tasks"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/pocketbase/pocketbase"
)

type ScheduledPost struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Link       string `json:"link"`
	Images     string `json:"images"`
	Status     string `json:"status"`
	Connection string `json:"connection"`
}

type Connections struct {
	ConnectionName string `json:"connection_name"`
	AccessToken    string `json:"access_token"`
	ConnectionId   string `json:"connection_id"`
}

func GetScheduledPosts(app *pocketbase.PocketBase) {
	time.Sleep(5 * time.Second) // wait for 5 seconds before executing the task

	if err := EnsureTables(app, "posts", "connections"); err != nil {
		app.Logger().Warn("Skipping scheduled post worker", "error", err.Error())
		return
	}

	var posts []ScheduledPost
	selectPostsQuery := `SELECT * FROM posts
			WHERE status = 'scheduled'
			AND datetime(publish_at) <= datetime('now')
			AND deleted = '';`
	// get all scheduled posts
	err := app.DB().NewQuery(selectPostsQuery).All(&posts)
	if err != nil {
		app.Logger().Error("Failed to fetch scheduled posts ", "query", selectPostsQuery, "error", err.Error())
		return
	}

	if len(posts) > 0 {

		for _, post := range posts {
			postId := post.ID

			record, err := app.FindRecordById("posts", postId)
			if err != nil {
				app.Logger().Error("Failed to find post record ", "postId", postId, "error", err.Error())
				continue
			}
			record.Set("status", "sending")
			if err := app.Save(record); err != nil {
				app.Logger().Error("Failed to update post status to sending.", "postId", postId, "error", err.Error())
				continue
			}

			content := post.Content
			title := post.Title
			link := post.Link

			var images []string
			err = json.Unmarshal([]byte(post.Images), &images)
			if err != nil {
				markPostFailed(app, postId, "scheduler: failed to parse images payload", err)
				continue
			}

			connectionId := post.Connection

			var connections []Connections

			selectConnectionsQuery := `SELECT connection_id, connection_name, access_token FROM connections WHERE id = '` + connectionId + `' AND deleted = "" ORDER BY id DESC;`

			// get social media connection details (access_token)
			err = app.DB().NewQuery(selectConnectionsQuery).All(&connections)

			if err != nil {
				markPostFailed(app, postId, "scheduler: failed to fetch connection details", err)
				continue
			}

			if len(connections) == 0 {
				markPostFailed(app, postId, "scheduler: no active connection found for post", errors.New("connection not found or deleted"))
				continue
			}

			for _, connection := range connections {
				connectionName := connection.ConnectionName
				accessToken := connection.AccessToken
				connectionId := connection.ConnectionId

				app.Logger().Info("Posting to social media"+connectionName, "connectionName", connectionName, "postId", postId, "content", content, "images", images)

				var postErr error
				switch connectionName {
				case "facebook":
					// schedule facebook posts
					postErr = tasks.FacebookPagePost(app, content, images, connectionId, accessToken, postId, link)
				case "instagram":
					// Post On Instagram
					postErr = tasks.InstagramPost(app, content, images, connectionId, accessToken, postId)
				case "twitter":
					// Post on twitter
					postErr = tasks.PostToTwitterProfile(app, content, images, connectionId, accessToken, postId)
				case "linkedin":
					// Post on linkedin
					postErr = tasks.LinkedinPost(app, content, images, connectionId, accessToken, postId)
				case "pinterest":
					// Post on linkedin
					postErr = tasks.PostToPinterestBoard(app, title, content, images, connectionId, accessToken, postId)
				case "discord":
					// Post on linkedin
					postErr = tasks.PostToDiscordChannel(app, content, images, connectionId, accessToken, postId)
				case "mastodon":
					// Post to mastadon
					postErr = tasks.PostToMastodon(app, content, images, connectionId, accessToken, postId)
				case "threads":
					// Post to mastadon
					postErr = tasks.PostToThreads(app, content, images, connectionId, accessToken, postId)
				case "reddit":
					// Post to mastadon
					postErr = tasks.PostToReddit(app, content, images, connectionId, accessToken, postId)
				default:
					postErr = fmt.Errorf("unsupported connection type: %s", connectionName)
				}

				if postErr != nil {
					markPostFailed(app, postId, "scheduler: failed to dispatch post to platform "+connectionName, postErr)
				}

			}

		}
	}
}

func markPostFailed(app *pocketbase.PocketBase, postId string, context string, err error) {
	logMessage := fmt.Sprintf("%s: %s", context, err.Error())
	tasks.FailedPost(app, "scheduler", postId, errors.New(logMessage))
	app.Logger().Error("Scheduled post failed", "postId", postId, "context", context, "error", err.Error())
}
