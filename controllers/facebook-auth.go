package controllers

import (
	"content-clock/helpers"
	"content-clock/models"
	"encoding/json"
	"net/http"
	"os"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Response struct {
	Data   []Data `json:"data"`
	Paging Paging `json:"paging"`
}

type Data struct {
	AccessToken  string     `json:"access_token"`
	Category     string     `json:"category"`
	CategoryList []Category `json:"category_list"`
	Name         string     `json:"name"`
	ID           string     `json:"id"`
	Tasks        []string   `json:"tasks"`
	Picture      Picture    `json:"picture"`
}

type Picture struct {
	Data PictureData `json:"data"`
}

type PictureData struct {
	Height       int    `json:"height"`
	IsSilhouette bool   `json:"is_silhouette"`
	URL          string `json:"url"`
	Width        int    `json:"width"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Paging struct {
	Cursors Cursors `json:"cursors"`
}

type Cursors struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

func SetupFacebookRoutes(se *core.ServeEvent, app *pocketbase.PocketBase) {
	se.Router.GET("/api/v1/auth/facebook/start", func(e *core.RequestEvent) error {
		BeginFacebookAuth(e)
		return nil
	})
	se.Router.GET("/api/v1/auth/facebook/callback", func(e *core.RequestEvent) error {
		FacebookOAuthCallback(e)
		return nil
	})
	se.Router.GET("/api/v1/add-facebook-pages", func(e *core.RequestEvent) error {
		AddFacebookPages(e, app)
		return nil
	})
}

func BeginFacebookAuth(e *core.RequestEvent) {
	var apiHost string = os.Getenv("API_HOST")
	var fbAppId string = os.Getenv("FACEBOOK_APP_ID")
	var fbAppSecret string = os.Getenv("FACEBOOK_SECRET")
	if fbAppId == "" || fbAppSecret == "" {
		helpers.Error(e, "Facebook App ID or Secret is not set")
		return
	}
	goth.UseProviders(facebook.New(fbAppId, fbAppSecret, apiHost+"/api/v1/auth/facebook/callback", "pages_manage_posts", "pages_show_list", "read_insights", "pages_read_engagement", "publish_video"))
	q := e.Request.URL.Query()
	q.Add("provider", "facebook")
	e.Request.URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(e.Response, e.Request)
}

func FacebookOAuthCallback(e *core.RequestEvent) {
	// Complete the user authentication
	user, err := gothic.CompleteUserAuth(e.Response, e.Request)
	if err != nil {
		e.Response.WriteHeader(http.StatusInternalServerError)
		return
	}

	accessToken := user.AccessToken
	userId := user.UserID
	var redirectHost string = os.Getenv("REDIRECT_HOST")
	e.Redirect(http.StatusTemporaryRedirect, redirectHost+"/connect/facebook?accessToken="+accessToken+"&userId="+userId)
	return
}

func AddFacebookPages(e *core.RequestEvent, app *pocketbase.PocketBase) {

	userId := e.Request.URL.Query().Get("fbUserId")
	accessToken := e.Request.URL.Query().Get("accessToken")
	authUserId := e.Request.URL.Query().Get("userId")

	if userId == "" || accessToken == "" || authUserId == "" {
		helpers.Error(e, "Missing required parameters")
		return
	}

	resp, err := http.Get("https://graph.facebook.com/" + userId + "/accounts?fields=picture,name,access_token&access_token=" + accessToken)
	if err != nil {
		app.Logger().Error("Failed to fetch Facebook pages: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}
	defer resp.Body.Close()

	// Decode JSON response
	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		app.Logger().Error("Failed to decode Facebook response: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}

	if len(response.Data) == 0 {
		helpers.Error(e, "No pages found for the user")
		return
	}

	for _, data := range response.Data {

		jsonData, err := json.Marshal(data)
		if err != nil {
			app.Logger().Error("Failed to marshal Facebook page data: " + err.Error())
			helpers.Error(e, err.Error())
			return
		}

		userData := models.Connections{
			UserId:         authUserId,
			Name:           data.Name,
			ConnectionName: "facebook",
			ConnectionId:   data.ID,
			AccessToken:    data.AccessToken,
			MetaData:       string(jsonData),
			ProfileImage:   data.Picture.Data.URL,
			Username:       data.Name,
			RefreshToken:   "",
		}

		err = AddNewConnection(app, &userData)

		if err != nil {
			app.Logger().Error("Failed to add Facebook page connection: " + err.Error())
			helpers.Error(e, err.Error())
			return
		}
	}

	helpers.Success(e, "Pages connected", map[string]interface{}{})
	return
}
