package controllers

import (
	"content-clock/helpers"
	"content-clock/models"
	"fmt"
	"net/http"
	"os"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/mastodon"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func SetupMastodonRoutes(se *core.ServeEvent, app *pocketbase.PocketBase) {
	se.Router.GET("/api/v1/auth/mastodon/start", func(e *core.RequestEvent) error {
		BeginMastodonAuth(e)
		return nil
	})
	se.Router.GET("/api/v1/auth/mastodon/callback", func(e *core.RequestEvent) error {
		MastodonCallback(e)
		return nil
	})
	se.Router.GET("/api/v1/add-mastodon-pages", func(e *core.RequestEvent) error {
		AddMastodonPages(e, app)
		return nil
	})
}

// GET /api/v1/auth/mastodon
func BeginMastodonAuth(e *core.RequestEvent) {

	var apiHost string = os.Getenv("API_HOST")
	var mdApiKey string = os.Getenv("MASTODON_CLIENT_KEY")
	var mdSecret string = os.Getenv("MASTODON_CLIENT_SECRET")
	if mdApiKey == "" || mdSecret == "" {
		helpers.Error(e, "Mastodon Client Key or Secret is not set")
		return
	}
	// var mdBaseUrl string = os.Getenv("MASTODON_BASE_URL")
	goth.UseProviders(mastodon.New(mdApiKey, mdSecret, apiHost+"/api/v1/auth/mastodon/callback", "read write follow"))
	q := e.Request.URL.Query()
	q.Add("provider", "mastodon")
	e.Request.URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(e.Response, e.Request)
}

// GET /api/v1/auth/mastodon/callback
func MastodonCallback(e *core.RequestEvent) {
	user, err := gothic.CompleteUserAuth(e.Response, e.Request)
	if err != nil {
		helpers.Error(e, "Authentication failed: "+err.Error())
		return
	}

	// You can save token or user info to DB here
	// For now, just show user data
	// helpers.Success(c, "Mastodon authenticated", gin.H{
	// 	"name":        user.Name,
	// 	"email":       user.Email,
	// 	"nickname":    user.NickName,
	// 	"accessToken": user.AccessToken,
	// 	"avatar":      user.AvatarURL,
	// })

	var redirectHost string = os.Getenv("REDIRECT_HOST")
	redirectURL := fmt.Sprintf(
		"%s/connect/mastodon?token=%s&user=%s&name=%s&username=%s&avatar=%s",
		redirectHost,
		user.AccessToken,
		user.UserID,
		user.Name,
		user.NickName,
		user.AvatarURL,
	)

	e.Redirect(http.StatusTemporaryRedirect, redirectURL)

}
func AddMastodonPages(e *core.RequestEvent, app *pocketbase.PocketBase) {
	accessToken := e.Request.URL.Query().Get("token")
	authUserId := e.Request.URL.Query().Get("userId")

	// These come from the redirect query
	name := e.Request.URL.Query().Get("name")
	userID := e.Request.URL.Query().Get("user")
	image := e.Request.URL.Query().Get("avatar")
	if accessToken == "" || authUserId == "" || userID == "" || name == "" || image == "" {
		helpers.Error(e, "Missing required parameters")
		return
	}

	userData := models.Connections{
		UserId:         authUserId,
		Name:           name,
		ConnectionName: "mastodon",
		ConnectionId:   userID,
		AccessToken:    accessToken,
		MetaData:       "",
		ProfileImage:   image,
		Username:       name,
		RefreshToken:   "",
	}

	if err := AddNewConnection(app, &userData); err != nil {
		app.Logger().Error("Mastodon: Failed to add connection: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}

	helpers.Success(e, "Mastodon connected successfully", map[string]interface{}{})

}
