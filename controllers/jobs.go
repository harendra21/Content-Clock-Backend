package controllers

import (
	"content-clock/tasks"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
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
				app.Logger().Error("Failed to parse images.", "postId", postId, "error", err.Error())
				continue
			}

			connectionId := post.Connection

			var connections []Connections

			selectConnectionsQuery := `SELECT connection_id, connection_name, access_token FROM connections WHERE id = '` + connectionId + `' AND deleted = "" ORDER BY id DESC;`

			// get social media connection details (access_token)

			// err = models.DB.Raw(selectConnectionsQuery).Scan(&connections).Error
			err = app.DB().NewQuery(selectConnectionsQuery).All(&connections)

			if err != nil {
				app.Logger().Error("Failed to fetch connections: ", "query", selectConnectionsQuery, "error", err.Error())
				continue
			}

			if len(connections) == 0 {
				app.Logger().Error("No connections found", "ConnectionId", connectionId, "postId", postId)
				continue
			}

			for _, connection := range connections {
				connectionName := connection.ConnectionName
				accessToken := connection.AccessToken
				connectionId := connection.ConnectionId

				app.Logger().Info("Posting to social media"+connectionName, "connectionName", connectionName, "postId", postId, "content", content, "images", images)

				switch connectionName {
				case "facebook":
					// schedule facebook posts
					tasks.FacebookPagePost(app, content, images, connectionId, accessToken, postId, link)
				case "instagram":
					// Post On Instagram
					tasks.InstagramPost(app, content, images, connectionId, accessToken, postId)
				case "twitter":
					// Post on twitter
					tasks.PostToTwitterProfile(app, content, images, connectionId, accessToken, postId)
				case "linkedin":
					// Post on linkedin
					tasks.LinkedinPost(app, content, images, connectionId, accessToken, postId)
				case "pinterest":
					// Post on linkedin
					tasks.PostToPinterestBoard(app, title, content, images, connectionId, accessToken, postId)
				case "discord":
					// Post on linkedin
					tasks.PostToDiscordChannel(content, images, connectionId, accessToken, postId)
				case "mastodon":
					// Post to mastadon
					tasks.PostToMastodon(app, content, images, connectionId, accessToken, postId)
				case "threads":
					// Post to mastadon
					tasks.PostToThreads(app, content, images, connectionId, accessToken, postId)
				case "reddit":
					// Post to mastadon
					tasks.PostToReddit(app, content, images, connectionId, accessToken, postId)
				}

			}

		}
	}
}

func parsePostgresStringArray(input string) ([]string, error) {
	input = strings.Trim(input, "{}")
	rawItems := strings.Split(input, ",")
	re := regexp.MustCompile(`^\\?"?(.+?)"?\\?$`)
	var results []string
	for _, item := range rawItems {
		item = strings.TrimSpace(item)
		matches := re.FindStringSubmatch(item)
		if len(matches) == 2 {
			results = append(results, matches[1])
		} else {
			return nil, fmt.Errorf("could not parse item: %s", item)
		}
	}
	return results, nil
}
