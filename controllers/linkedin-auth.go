package controllers

import (
	"content-clock/helpers"
	"content-clock/models"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
)

func SetupLinkedinRoutes(se *core.ServeEvent, app *pocketbase.PocketBase) {
	se.Router.GET("/api/v1/auth/linkedin/start", func(e *core.RequestEvent) error {
		BeginLinkedinAuth(e)
		return nil
	})
	se.Router.GET("/api/v1/auth/linkedin/callback", func(e *core.RequestEvent) error {
		LinkedinOAuthCallback(e, app)
		return nil
	})
	se.Router.GET("/api/v1/add-linkedin-pages", func(e *core.RequestEvent) error {
		AddLinkedinPages(e, app)
		return nil
	})
}

func BeginLinkedinAuth(e *core.RequestEvent) {
	linkedinAppId := os.Getenv("LINKEDIN_APP_ID")
	linkedinSecret := os.Getenv("LINKEDIN_SECRET")
	if linkedinAppId == "" || linkedinSecret == "" {
		helpers.Error(e, "LinkedIn App ID or Secret is not set")
		return
	}

	linkedinOauthConfig := &oauth2.Config{
		ClientID:     linkedinAppId,
		ClientSecret: linkedinSecret,
		RedirectURL:  os.Getenv("API_HOST") + "/api/v1/auth/linkedin/callback",
		Scopes:       []string{"w_member_social", "profile", "openid", "email"},
		Endpoint:     linkedin.Endpoint,
	}

	url := linkedinOauthConfig.AuthCodeURL("state")
	e.Redirect(http.StatusTemporaryRedirect, url)
}

func LinkedinOAuthCallback(e *core.RequestEvent, app *pocketbase.PocketBase) {
	linkedinOauthConfig := &oauth2.Config{
		ClientID:     os.Getenv("LINKEDIN_APP_ID"),
		ClientSecret: os.Getenv("LINKEDIN_SECRET"),
		RedirectURL:  os.Getenv("API_HOST") + "/api/v1/auth/linkedin/callback",
		Scopes:       []string{"w_member_social", "profile", "openid", "email"},
		Endpoint:     linkedin.Endpoint,
	}

	code := e.Request.URL.Query().Get("code")
	ctx := context.Background()
	token, err := linkedinOauthConfig.Exchange(ctx, code)
	if err != nil {
		app.Logger().Error("Failed to exchange LinkedIn token: " + err.Error())
		e.String(http.StatusInternalServerError, "Failed to exchange token")
		return
	}
	accessToken := token.AccessToken
	var redirectHost string = os.Getenv("REDIRECT_HOST")
	e.Redirect(http.StatusTemporaryRedirect, redirectHost+"/connect/linkedin?accessToken="+accessToken)
	return
}

// Struct to match the JSON response
type User struct {
	Sub           string `json:"sub"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Locale        struct {
		Country  string `json:"country"`
		Language string `json:"language"`
	} `json:"locale"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Email      string `json:"email"`
	Picture    string `json:"picture"`
}

func AddLinkedinPages(e *core.RequestEvent, app *pocketbase.PocketBase) {

	accessToken := e.Request.URL.Query().Get("accessToken")
	authUserId := e.Request.URL.Query().Get("userId")

	if accessToken == "" || authUserId == "" {
		helpers.Error(e, "Missing required parameters")
		return
	}

	url := "https://api.linkedin.com/v2/userinfo"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		app.Logger().Error("Linkedin: Failed to create HTTP request: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		app.Logger().Error("Linkedin: Failed to send HTTP request: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		app.Logger().Error("Linkedin: API request failed with status: " + resp.Status)
		helpers.Error(e, err.Error())
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		app.Logger().Error("Linkedin: Failed to read response body: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		app.Logger().Error("Linkedin: Failed to parse JSON response: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}

	userData := models.Connections{
		UserId:         authUserId,
		Name:           user.Name,
		ConnectionName: "linkedin",
		ConnectionId:   user.Sub,
		AccessToken:    accessToken,
		MetaData:       string(body),
		ProfileImage:   user.Picture,
		Username:       user.Sub,
		RefreshToken:   "",
	}

	err = AddNewConnection(app, &userData)

	if err != nil {
		app.Logger().Error("Linkedin: Failed to add connection: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}

	helpers.Success(e, "Pages connected", map[string]interface{}{})

	return

}
