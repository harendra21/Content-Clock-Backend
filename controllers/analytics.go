package controllers

import (
	"fmt"
	"io"
	"net/http"
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

		if connection.ConnectionName == "facebook" {
			FetchFacebookAnalytics(app, post, connection)
		}
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
	analytics := Analytics{}
	err = app.DB().Select("id").From("analytics").Where(dbx.NewExp("post = {:post}", dbx.Params{"post": post.Id})).One(&analytics)
	if err != nil && err.Error() != "sql: no rows in result set" {
		app.Logger().Error("Error in fetching facebook analytics", "error", err.Error())
		return
	}
	if err != nil && err.Error() == "sql: no rows in result set" {

		_, err := app.DB().Insert("analytics", dbx.Params{
			"post":    post.Id,
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
		app.Logger().Info("Successfully fetched the facebook analytics")
	}

}
