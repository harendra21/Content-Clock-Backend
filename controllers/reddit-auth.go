package controllers

import (
	"content-clock/helpers"
	"content-clock/models"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type RedditUser struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Icon  string `json:"icon_img"`
	Token string
}

func SetupRedditRoutes(se *core.ServeEvent, app *pocketbase.PocketBase) {
	se.Router.GET("/api/v1/auth/reddit/start", func(e *core.RequestEvent) error {
		BeginRedditAuth(e)
		return nil
	})
	se.Router.GET("/api/v1/auth/reddit/callback", func(e *core.RequestEvent) error {
		RedditOAuthCallback(e, app)
		return nil
	})
	se.Router.GET("/api/v1/add-reddit-pages", func(e *core.RequestEvent) error {
		AddRedditConnection(e, app)
		return nil
	})
	se.Router.GET("/api/v1/reddit/post", func(e *core.RequestEvent) error {
		PostToReddit(e)
		return nil
	})
	se.Router.GET("/api/v1/reddit/analytics", func(e *core.RequestEvent) error {
		GetRedditAnalytics(e)
		return nil
	})
}

func BeginRedditAuth(e *core.RequestEvent) {
	var clientID string = os.Getenv("REDDIT_CLIENT_ID")
	var redirectURL string = os.Getenv("API_HOST") + "/api/v1/auth/reddit/callback"
	var scopes string = "identity submit read"
	//
	state := os.Getenv("JWT_KEY")
	url := fmt.Sprintf("https://www.reddit.com/api/v1/authorize?client_id=%s&response_type=code&state=%s&redirect_uri=%s&duration=permanent&scope=%s", clientID, state, redirectURL, scopes)

	http.Redirect(e.Response, e.Request, url, http.StatusTemporaryRedirect)
}

type RedditTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func RedditOAuthCallback(e *core.RequestEvent, app *pocketbase.PocketBase) {
	var clientID string = os.Getenv("REDDIT_CLIENT_ID")
	var clientSecret string = os.Getenv("REDDIT_SECRET")
	var redirectURL string = os.Getenv("API_HOST") + "/api/v1/auth/reddit/callback"

	state := e.Request.URL.Query().Get("state")
	if state != os.Getenv("JWT_KEY") {
		helpers.Error(e, "Invalid state")
		return
	}

	code := e.Request.URL.Query().Get("code")
	if code == "" {
		helpers.Error(e, "Code not found")
		return
	}

	exchangeUrl := "https://www.reddit.com/api/v1/access_token"
	method := "POST"
	basicAuthToken := basicAuth(clientID, clientSecret)

	header := map[string]string{
		"Authorization": fmt.Sprintf("Basic %s", basicAuthToken),
		"Content-Type":  "application/x-www-form-urlencoded",
		"User-agent":    "Content Clock Local 0.1",
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURL)
	resp, err := helpers.MakeHTTPRequest[RedditTokenResponse](app, method, exchangeUrl, header, nil, data)
	if err != nil {
		helpers.Error(e, "Error in reddit token exchange")
		app.Logger().Error("Error in reddit token exchange", "error", err.Error())
		return
	}

	var redirectHost string = os.Getenv("REDIRECT_HOST")
	e.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/connect/reddit?accessToken=%s&refreshToken=%s&expiresIn=%d", redirectHost, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn))

}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

type RedditUserDetails struct {
	Name         string `json:"name"`
	SnoovatarImg string `json:"snoovatar_img"`
	Subreddit    struct {
		Title string `json:"title"`
	} `json:"subreddit"`
}

func AddRedditConnection(e *core.RequestEvent, app *pocketbase.PocketBase) {

	accessToken := e.Request.URL.Query().Get("accessToken")
	refreshToken := e.Request.URL.Query().Get("refreshToken")
	authUserId := e.Request.URL.Query().Get("userId")
	expiresIn := e.Request.URL.Query().Get("expiresIn")

	reqUrl := "https://oauth.reddit.com/api/v1/me"
	header := map[string]string{
		"Authorization": fmt.Sprintf("bearer %s", accessToken),
		"User-agent":    "Content Clock Local 0.1",
	}

	resp, err := helpers.MakeHTTPRequest[RedditUserDetails](app, "GET", reqUrl, header, nil, nil)
	if err != nil {
		helpers.Error(e, "Error in reddit profile fetch")
		app.Logger().Error("Error in reddit profile fetch", "error", err.Error())
		return
	}

	userData := models.Connections{
		UserId:         authUserId,
		Name:           resp.Subreddit.Title,
		ConnectionName: "reddit",
		ConnectionId:   resp.Name,
		AccessToken:    accessToken,
		MetaData:       "expiresIn=" + expiresIn,
		ProfileImage:   resp.SnoovatarImg,
		Username:       resp.Name,
		RefreshToken:   refreshToken,
	}

	err = AddNewConnection(app, &userData)

	if err != nil {
		app.Logger().Error("Failed to add reddit page connection", "error", err.Error())
		helpers.Error(e, err.Error())
		return
	}

	helpers.Success(e, "Pages connected", map[string]interface{}{})
	return
}

func PostToReddit(e *core.RequestEvent) error {
	subreddit := e.Request.URL.Query().Get("subreddit")
	title := e.Request.URL.Query().Get("title")
	text := e.Request.URL.Query().Get("text")
	token := e.Request.URL.Query().Get("accessToken")

	if subreddit == "" || title == "" || token == "" {
		helpers.Error(e, "Missing required fields")
		return nil
	}

	req, _ := http.NewRequest("POST", "https://oauth.reddit.com/api/submit", strings.NewReader("sr="+subreddit+"&title="+title+"&text="+text+"&kind=self"))
	req.Header.Set("Authorization", "bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		helpers.Error(e, "Failed to post: "+err.Error())
		return nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	helpers.Success(e, "Post created", result)
	return nil
}

func GetRedditAnalytics(e *core.RequestEvent) error {
	postID := e.Request.URL.Query().Get("postId")
	token := e.Request.URL.Query().Get("accessToken")
	if postID == "" || token == "" {
		helpers.Error(e, "Missing postId or accessToken")
		return nil
	}

	req, _ := http.NewRequest("GET", "https://oauth.reddit.com/api/info?id=t3_"+postID, nil)
	req.Header.Set("Authorization", "bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		helpers.Error(e, "Failed to get analytics: "+err.Error())
		return nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	helpers.Success(e, "Fetched analytics", result)
	return nil
}
