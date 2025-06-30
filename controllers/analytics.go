package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

type Post struct {
	Id              string `db:"id" json:"id"`
	Connection      string `db:"connection" json:"connection"`
	PublishedPostId string `db:"published_post_id" json:"published_post_id"`
}

type Connection struct {
	ConnectionName string `db:"connection_name" json:"connection_name"`
	AccessToken    string `db:"access_token" json:"access_token"`
}

type Analytics struct {
	Id string `db:"id" json:"id"`
}

func FetchPostsAnalytics(app *pocketbase.PocketBase) {
	time.Sleep(5)
	posts := []Post{}
	err := app.DB().Select("id", "connection", "published_post_id").From("posts").Where(dbx.NewExp("status = {:status}", dbx.Params{"status": "published"})).All(&posts)
	if err != nil {
		app.Logger().Error("Error in getting the posts", "error", err)
	}

	for _, post := range posts {
		connection := Connection{}
		err := app.DB().Select("connection_name", "access_token").From("connections").Where(dbx.NewExp("id = {:id}", dbx.Params{"id": post.Connection})).One(&connection)
		if err != nil {
			app.Logger().Error("Error in getting the connection", "error", err)
		}

		switch connection.ConnectionName {
		case "facebook":
			FetchFacebookAnalytics(app, post, connection)
		case "linkedin":
			FetchLinkedInPostAnalytics(app, post, connection)
		case "instagram":
			FetchInstagramPostAnalytics(app, post, connection)
		case "mastodon":
			FetchMastodonPostAnalytics(app, post, connection)
		case "pinterest":
			FetchPinterestPostAnalytics(app, post, connection)
		default:
			app.Logger().Warn("Unsupported connection type", "type", connection.ConnectionName)
		}
	}

}

func SaveUpdateAnalyticsData(app *pocketbase.PocketBase, postId string, data string) {

	analytics := Analytics{}
	err := app.DB().Select("id").From("analytics").Where(dbx.NewExp("post = {:post}", dbx.Params{"post": postId})).One(&analytics)
	if err != nil && err.Error() != "sql: no rows in result set" {
		app.Logger().Error("Error in fetching facebook analytics", "error", err.Error())
		return
	}
	if err != nil && err.Error() == "sql: no rows in result set" {
		_, err := app.DB().Insert("analytics", dbx.Params{
			"post":    postId,
			"data":    data,
			"created": time.Now(),
			"updated": time.Now(),
		}).Execute()

		if err != nil {
			app.Logger().Error("Error inserting new analytics", "error", err.Error())
			return
		}

	} else {
		record, _ := app.FindRecordById("analytics", analytics.Id)
		record.Set("data", data)
		app.Save(record)
		app.Logger().Info("Successfully fetched the analytics")
	}

}

func FetchFacebookAnalytics(app *pocketbase.PocketBase, post Post, connection Connection) {
	url := fmt.Sprintf("https://graph.facebook.com/%s/insights?metric=post_impressions,post_reactions_like_total,post_reactions_love_total,post_reactions_wow_total&access_token=%s", post.PublishedPostId, connection.AccessToken)
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		app.Logger().Error("Error in fetching facebook analytics", "error", err.Error())
		return
	}
	res, err := client.Do(req)
	if err != nil {
		app.Logger().Error("Error in fetching facebook analytics", "error", err.Error())
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		app.Logger().Error("Error in fetching facebook analytics", "error", err.Error())
		return
	}
	data := string(body)
	SaveUpdateAnalyticsData(app, post.Id, data)

}

func FetchLinkedInPostAnalytics(app *pocketbase.PocketBase, post Post, connection Connection) {
	url := fmt.Sprintf("https://api.linkedin.com/v2/socialActions/%s", post.PublishedPostId)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+connection.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		app.Logger().Error("Error in fetching linkedin analytics", "error", err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	data := string(body)
	SaveUpdateAnalyticsData(app, post.Id, data)
}

func FetchInstagramPostAnalytics(app *pocketbase.PocketBase, post Post, connection Connection) {
	url := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/insights?metric=impressions,reach,engagement,saved&access_token=%s", post.PublishedPostId, connection.AccessToken)

	req, _ := http.NewRequest("GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		app.Logger().Error("Error in fetching instagram analytics", "error", err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	data := string(body)
	SaveUpdateAnalyticsData(app, post.Id, data)
}

func FetchPinterestPostAnalytics(app *pocketbase.PocketBase, post Post, connection Connection) {
	url := fmt.Sprintf("https://api.pinterest.com/v5/pins/%s/analytics", post.PublishedPostId)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+connection.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		app.Logger().Error("Error in fetching pinterest analytics", "error", err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	data := string(body)
	SaveUpdateAnalyticsData(app, post.Id, data)
}

func FetchMastodonPostAnalytics(app *pocketbase.PocketBase, post Post, connection Connection) {
	instanceBaseURL := os.Getenv("MASTODON_BASE_URL")

	var result struct {
		ID string `json:"id"`
	}

	err := json.Unmarshal([]byte(post.PublishedPostId), &result)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	url := fmt.Sprintf("%s/api/v1/statuses/%s", instanceBaseURL, result.ID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+connection.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		app.Logger().Error("Error in fetching mastodon analytics", "error", err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	data := string(body)
	SaveUpdateAnalyticsData(app, post.Id, data)
}
