package controllers

import (
	"content-clock/helpers"
	"content-clock/models"
	"fmt"
	"net/http"
	"net/url"
	"os"

	// Hypothetical Threads provider, replace with actual if available
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func SetupThreadsRoutes(se *core.ServeEvent, app *pocketbase.PocketBase) {
	se.Router.GET("/api/v1/auth/threads/start", func(e *core.RequestEvent) error {
		BeginThreadsAuth(e, app)
		return nil
	})
	se.Router.GET("/api/v1/auth/threads/callback", func(e *core.RequestEvent) error {
		ThreadsOAuthCallback(e, app)
		return nil
	})
	se.Router.GET("/api/v1/add-threads-pages", func(e *core.RequestEvent) error {
		AddThreadsAccounts(e, app)
		return nil
	})
}

func BeginThreadsAuth(e *core.RequestEvent, app *pocketbase.PocketBase) {
	var apiHost string = os.Getenv("API_HOST")
	var threadsAppId string = os.Getenv("THREADS_APP_ID")
	var threadsAppSecret string = os.Getenv("THREADS_SECRET_KEY")
	if threadsAppId == "" || threadsAppSecret == "" {
		helpers.Error(e, "Threads App ID or Secret is not set")
		return
	}

	callbackUrl := apiHost + "/api/v1/auth/threads/callback"
	scopes := "threads_basic,threads_content_publish,threads_manage_insights"

	redirectUrl := fmt.Sprintf("https://threads.net/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s", threadsAppId, callbackUrl, scopes)
	e.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	UserId      any    `json:"user_id"`
}

type LongLivedTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func ThreadsOAuthCallback(e *core.RequestEvent, app *pocketbase.PocketBase) {
	var apiHost string = os.Getenv("API_HOST")
	code := e.Request.URL.Query().Get("code")
	var url string = "https://graph.threads.net/oauth/access_token"
	var method string = "POST"
	var threadsAppId string = os.Getenv("THREADS_APP_ID")
	var threadsAppSecret string = os.Getenv("THREADS_SECRET_KEY")
	var callbackUrl string = apiHost + "/api/v1/auth/threads/callback"
	var body map[string]interface{} = map[string]interface{}{
		"client_id":     threadsAppId,
		"client_secret": threadsAppSecret,
		"grant_type":    "authorization_code",
		"redirect_uri":  callbackUrl,
		"code":          code,
	}

	resp, err := helpers.MakeHTTPRequest[TokenResponse](app, method, url, nil, nil, body)
	if err != nil {
		app.Logger().Error("Error in fetching token", "error", err.Error())
		e.Response.WriteHeader(http.StatusInternalServerError)
		return
	}

	var accessToken string = resp.AccessToken
	var userId any = resp.UserId
	var redirectHost string = os.Getenv("REDIRECT_HOST")
	e.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/connect/threads?accessToken=%s&userId=%v", redirectHost, accessToken, userId))
	return
}

type ThreadsProfileResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Image    string `json:"threads_profile_picture_url"`
}

func AddThreadsAccounts(e *core.RequestEvent, app *pocketbase.PocketBase) {
	userId := e.Request.URL.Query().Get("threadsUserId")
	accessToken := e.Request.URL.Query().Get("accessToken")
	authUserId := e.Request.URL.Query().Get("userId")

	if userId == "" || accessToken == "" || authUserId == "" {
		helpers.Error(e, "Missing required parameters")
		return
	}
	var threadsAppSecret string = os.Getenv("THREADS_SECRET_KEY")

	params := url.Values{}
	params.Add("grant_type", "th_exchange_token")
	params.Add("client_secret", threadsAppSecret)
	params.Add("access_token", accessToken)

	longToken := "https://graph.threads.net/access_token"

	lResp, err := helpers.MakeHTTPRequest[LongLivedTokenResponse](app, "GET", longToken, nil, params, nil)
	if err != nil {
		app.Logger().Error("Error in getting threads long lived token", "error", err.Error())
		helpers.Error(e, "Error in getting threads long lived token")
		return
	}

	accessToken = lResp.AccessToken

	profileParams := url.Values{}
	profileParams.Add("fields", "id,username,threads_profile_picture_url,name")
	profileParams.Add("access_token", accessToken)
	url := "https://graph.threads.net/v1.0/me"
	method := "GET"

	resp, err := helpers.MakeHTTPRequest[ThreadsProfileResponse](app, method, url, nil, profileParams, nil)
	if err != nil {
		app.Logger().Error("Error in getting threads profile", "error", err.Error())
		helpers.Error(e, "Error in getting threads profile")
		return
	}

	userData := models.Connections{
		UserId:         authUserId,
		Name:           resp.Name,
		ConnectionName: "threads",
		ConnectionId:   resp.ID,
		AccessToken:    accessToken,
		MetaData:       "",
		ProfileImage:   resp.Image,
		Username:       resp.Username,
		RefreshToken:   "",
	}

	err = AddNewConnection(app, &userData)
	if err != nil {
		app.Logger().Error("Failed to add Threads account connection: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}
	helpers.Success(e, "Pages connected", map[string]interface{}{})
	return
}
