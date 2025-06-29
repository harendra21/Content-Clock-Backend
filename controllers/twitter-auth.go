package controllers

import (
	"content-clock/helpers"
	"content-clock/models"
	"fmt"
	"net/http"
	"os"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/twitter"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func SetupTwitterRoutes(se *core.ServeEvent, app *pocketbase.PocketBase) {
	se.Router.GET("/api/v1/auth/twitter/start", func(e *core.RequestEvent) error {
		BeginTwitterAuth(e)
		return nil
	})
	se.Router.GET("/api/v1/auth/twitter/callback", func(e *core.RequestEvent) error {
		TwitterOAuthCallback(e)
		return nil
	})
	se.Router.GET("/api/v1/add-twitter-pages", func(e *core.RequestEvent) error {
		AddTwitterPages(e, app)
		return nil
	})
}

func BeginTwitterAuth(e *core.RequestEvent) {
	var apiHost string = os.Getenv("API_HOST")
	var twApiKey string = os.Getenv("TWITTER_KEY")
	var twSecret string = os.Getenv("TWITTER_SECRET")

	if twApiKey == "" || twSecret == "" {
		helpers.Error(e, "Twitter API Key or Secret is not set")
		return
	}

	goth.UseProviders(twitter.New(twApiKey, twSecret, apiHost+"/api/v1/auth/twitter/callback"))
	q := e.Request.URL.Query()
	q.Add("provider", "twitter")
	e.Request.URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(e.Response, e.Request)
}

func TwitterOAuthCallback(e *core.RequestEvent) {
	user, err := gothic.CompleteUserAuth(e.Response, e.Request)
	if err != nil {
		helpers.Error(e, err.Error())
		return
	}

	var redirectHost string = os.Getenv("REDIRECT_HOST")
	redirectURL := fmt.Sprintf(
		"%s/connect/twitter?token=%s&secret=%s&user=%s&name=%s&username=%s&avatar=%s",
		redirectHost,
		user.AccessToken,
		user.AccessTokenSecret,
		user.UserID,
		user.Name,
		user.NickName,
		user.AvatarURL,
	)

	e.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

type TwitterSuccessResponse struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

func AddTwitterPages(e *core.RequestEvent, app *pocketbase.PocketBase) {
	accessToken := e.Request.URL.Query().Get("token")
	accessSecret := e.Request.URL.Query().Get("secret")
	authUserId := e.Request.URL.Query().Get("userId")

	// These come from the redirect query
	name := e.Request.URL.Query().Get("name")
	userID := e.Request.URL.Query().Get("user")
	username := e.Request.URL.Query().Get("username")
	image := e.Request.URL.Query().Get("avatar")
	if accessToken == "" || accessSecret == "" || authUserId == "" || userID == "" || name == "" || username == "" || image == "" {
		helpers.Error(e, "Missing required parameters")
		return
	}

	uAccessToken := accessToken + " " + accessSecret

	userData := models.Connections{
		UserId:         authUserId,
		Name:           name,
		ConnectionName: "twitter",
		ConnectionId:   userID,
		AccessToken:    uAccessToken,
		MetaData:       fmt.Sprintf(`{"user_id": "%s", "username": "%s"}`, userID, username),
		ProfileImage:   image,
		Username:       username,
		RefreshToken:   "", // Twitter OAuth1.0a has no refresh tokens
	}

	if err := AddNewConnection(app, &userData); err != nil {
		app.Logger().Error("Twitter: Failed to add connection: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}

	helpers.Success(e, "Twitter connected successfully", map[string]interface{}{})

}
