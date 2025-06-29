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

func SetupInstagramRoutes(se *core.ServeEvent, app *pocketbase.PocketBase) {
	se.Router.GET("/api/v1/auth/instagram/start", func(e *core.RequestEvent) error {
		BeginInstagramAuth(e)
		return nil
	})
	se.Router.GET("/api/v1/auth/instagram/callback", func(e *core.RequestEvent) error {
		InstagramOAuthCallback(e, app)
		return nil
	})
	se.Router.GET("/api/v1/add-instagram-pages", func(e *core.RequestEvent) error {
		AddInstagramPages(e, app)
		return nil
	})
}

func BeginInstagramAuth(e *core.RequestEvent) {
	var apiHost string = os.Getenv("API_HOST")
	var fbAppId string = os.Getenv("FACEBOOK_APP_ID")
	var fbAppSecret string = os.Getenv("FACEBOOK_SECRET")

	if fbAppId == "" || fbAppSecret == "" {
		helpers.Error(e, "Facebook App ID or Secret is not set")
		return
	}

	goth.UseProviders(facebook.New(
		fbAppId,
		fbAppSecret,
		apiHost+"/api/v1/auth/instagram/callback",
		"instagram_basic",
		"instagram_content_publish",
		"instagram_manage_comments",
		"instagram_manage_insights",
		// "pages_show_list",       // Added missing scope for accessing Facebook pages
		"pages_read_engagement", // Added for page insights
	))

	q := e.Request.URL.Query()
	q.Add("provider", "facebook")
	e.Request.URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(e.Response, e.Request)
}

func InstagramOAuthCallback(e *core.RequestEvent, app *pocketbase.PocketBase) {
	// Initialize providers again in callback to ensure consistency
	var apiHost string = os.Getenv("API_HOST")
	var fbAppId string = os.Getenv("FACEBOOK_APP_ID")
	var fbAppSecret string = os.Getenv("FACEBOOK_SECRET")

	goth.UseProviders(facebook.New(
		fbAppId,
		fbAppSecret,
		apiHost+"/api/v1/auth/instagram/callback",
		"instagram_basic",
		"instagram_content_publish",
		"instagram_manage_comments",
		"instagram_manage_insights",
		// "pages_show_list",
		"pages_read_engagement",
	))

	q := e.Request.URL.Query()
	q.Add("provider", "facebook")
	e.Request.URL.RawQuery = q.Encode()

	user, err := gothic.CompleteUserAuth(e.Response, e.Request)
	if err != nil {
		e.Response.WriteHeader(http.StatusInternalServerError)
		app.Logger().Error("Instagram OAuth callback failed: " + err.Error())
		return
	}

	accessToken := user.AccessToken
	userId := user.UserID
	var redirectHost string = os.Getenv("REDIRECT_HOST")

	// Pass the Facebook user ID as fbUserId to match the expected parameter in AddInstagramPages
	e.Redirect(http.StatusTemporaryRedirect, redirectHost+"/connect/instagram?accessToken="+accessToken+"&fbUserId="+userId+"&userId="+userId)
	return
}

// InstagramBusinessAccount represents the Instagram business account details.
type InstagramBusinessAccount struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Username          string `json:"username"`
	ProfilePictureURL string `json:"profile_picture_url"`
}

// DataItem represents an item in the "data" array.
type DataItem struct {
	InstagramBusinessAccount *InstagramBusinessAccount `json:"instagram_business_account,omitempty"`
	ID                       string                    `json:"id"`
	Name                     string                    `json:"name"`         // Added Facebook page name
	AccessToken              string                    `json:"access_token"` // Added page access token
}

// Fixed Paging struct - was referencing undefined "Paging" type
type InstaPaging struct {
	Cursors struct {
		Before string `json:"before"`
		After  string `json:"after"`
	} `json:"cursors"`
}

// Response represents the entire HTTP response structure.
type InstaResponse struct {
	Data   []DataItem  `json:"data"`
	Paging InstaPaging `json:"paging"` // Fixed: was using undefined "Paging" type
}

func AddInstagramPages(e *core.RequestEvent, app *pocketbase.PocketBase) {
	fbUserId := e.Request.URL.Query().Get("fbUserId") // Fixed: was "userId"
	accessToken := e.Request.URL.Query().Get("accessToken")
	authUserId := e.Request.URL.Query().Get("userId")

	if fbUserId == "" || accessToken == "" || authUserId == "" {
		app.Logger().Error("Missing required parameters: fbUserId, accessToken, or userId")
		helpers.Error(e, "Missing required parameters")
		return
	}

	// Updated API URL to include access_token for pages and their Instagram accounts
	graphApi := "https://graph.facebook.com/" + fbUserId + "/accounts?fields=name,access_token,instagram_business_account{id,name,username,profile_picture_url}&access_token=" + accessToken

	app.Logger().Info("Connecting to Instagram API", "url", graphApi)
	resp, err := http.Get(graphApi)
	if err != nil {
		app.Logger().Error("Failed to connect to Instagram API: " + err.Error())
		helpers.Error(e, "Unable to connect to Instagram API")
		return
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		app.Logger().Error("Instagram API returned error status: " + resp.Status)
		helpers.Error(e, "Instagram API returned an error")
		return
	}

	// Decode JSON response
	var response InstaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		app.Logger().Error("Failed to parse Instagram API response: " + err.Error())
		helpers.Error(e, "Unable to parse Instagram API response")
		return // Added missing return statement
	}

	if len(response.Data) == 0 {
		app.Logger().Info("No Instagram business accounts found for the user", "fbUserId", fbUserId)
		return
	}

	// Track successful connections
	connectedCount := 0

	for _, data := range response.Data {
		// Only process pages that have Instagram business accounts
		if data.InstagramBusinessAccount == nil {
			app.Logger().Info("Skipping Facebook page without Instagram account", "pageId", data.ID, "pageName", data.Name)
			continue
		}

		app.Logger().Info("Adding Instagram business account",
			"instagramId", data.InstagramBusinessAccount.ID,
			"instagramName", data.InstagramBusinessAccount.Name,
			"instagramUsername", data.InstagramBusinessAccount.Username)

		jsonData, err := json.Marshal(data)
		if err != nil {
			app.Logger().Error("Failed to marshal Instagram business account data: " + err.Error())
			continue // Continue with other accounts instead of returning
		}

		userData := models.Connections{
			UserId:         authUserId,
			Name:           data.InstagramBusinessAccount.Name,
			ConnectionName: "instagram",
			ConnectionId:   data.InstagramBusinessAccount.ID,
			AccessToken:    data.AccessToken, // Use page access token instead of user access token
			MetaData:       string(jsonData),
			ProfileImage:   data.InstagramBusinessAccount.ProfilePictureURL,
			Username:       data.InstagramBusinessAccount.Username,
			RefreshToken:   "", // Instagram doesn't use refresh tokens in this flow
		}

		err = AddNewConnection(app, &userData)
		if err != nil {
			app.Logger().Error("Failed to add Instagram business account connection: " + err.Error())
			continue
		}
		connectedCount++
	}

	if connectedCount == 0 {
		app.Logger().Info("No Instagram business accounts were successfully connected")
		return
	}

	helpers.Success(e, "Instagram Pages Connected", map[string]interface{}{
		"connected_count": connectedCount,
	})
	return
}
