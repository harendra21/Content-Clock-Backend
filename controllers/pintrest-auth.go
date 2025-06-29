package controllers

import (
	"content-clock/helpers"
	"content-clock/models"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type PinMedia struct {
	ImageCoverURL    string   `json:"image_cover_url"`
	PinThumbnailURLs []string `json:"pin_thumbnail_urls"`
}

type PinOwner struct {
	Username string `json:"username"`
}

type PinItem struct {
	ID                  string   `json:"id"`
	CreatedAt           string   `json:"created_at"`
	BoardPinsModifiedAt string   `json:"board_pins_modified_at"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	CollaboratorCount   int      `json:"collaborator_count"`
	PinCount            int      `json:"pin_count"`
	FollowerCount       int      `json:"follower_count"`
	Media               PinMedia `json:"media"`
	Owner               PinOwner `json:"owner"`
	Privacy             string   `json:"privacy"`
}

type PinResponse struct {
	Items    []PinItem `json:"items"`
	Bookmark string    `json:"bookmark"`
}

// https://www.pinterest.com/oauth/?client_id=1498608&redirect_uri=http://localhost:8080/v1/api/auth/pinterest/callback&response_type=code&scope=boards:read,pins:read,user_accounts:read&state=harendraverma21

func SetupPinterestRoutes(se *core.ServeEvent, app *pocketbase.PocketBase) {
	se.Router.GET("/api/v1/auth/pinterest/start", func(e *core.RequestEvent) error {
		BeginPinterestAuth(e)
		return nil
	})
	se.Router.GET("/api/v1/auth/pinterest/callback", func(e *core.RequestEvent) error {
		PinterestOAuthCallback(e)
		return nil
	})
	se.Router.GET("/api/v1/add-pinterest-pages", func(e *core.RequestEvent) error {
		AddPinterestPages(e, app)
		return nil
	})
}

func BeginPinterestAuth(e *core.RequestEvent) {
	var pinterestAppId string = os.Getenv("PINTEREST_APP_ID")
	var apiHost string = os.Getenv("API_HOST")

	if pinterestAppId == "" {
		helpers.Error(e, "Pinterest App ID is not set")
		return
	}

	state := os.Getenv("JWT_KEY")

	scopes := "boards:read,pins:read,user_accounts:read,boards:write,pins:write"

	redirect_uri := apiHost + "/api/v1/auth/pinterest/callback"

	url := "https://www.pinterest.com/oauth/?client_id=" + pinterestAppId + "&redirect_uri=" + redirect_uri + "&response_type=code&scope=" + scopes + "&state=" + state
	e.Redirect(http.StatusTemporaryRedirect, url)
}

func PinterestOAuthCallback(e *core.RequestEvent) {

	code := e.Request.URL.Query().Get("code")
	var redirectHost string = os.Getenv("REDIRECT_HOST")
	e.Redirect(http.StatusTemporaryRedirect, redirectHost+"/connect/pinterest?code="+code)
	return
}

type PinterestResponse struct {
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	ResponseType          string `json:"response_type"`
	TokenType             string `json:"token_type"`
	ExpiresIn             int    `json:"expires_in"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
	Scope                 string `json:"scope"`
}

func AddPinterestPages(e *core.RequestEvent, app *pocketbase.PocketBase) {

	code := e.Request.URL.Query().Get("code")
	authUserId := e.Request.URL.Query().Get("userId")
	if code == "" || authUserId == "" {
		helpers.Error(e, "Missing required parameters")
		return
	}

	pinUrl := "https://api.pinterest.com/v5/oauth/token"
	method := "POST"

	var apiHost string = os.Getenv("API_HOST")
	redirect_uri := apiHost + "/api/v1/auth/pinterest/callback"
	urlEncode := url.QueryEscape(redirect_uri)

	payload := strings.NewReader("grant_type=authorization_code&code=" + code + "&redirect_uri=" + urlEncode)

	client := &http.Client{}
	req, err := http.NewRequest(method, pinUrl, payload)

	if err != nil {
		app.Logger().Error("Pinterest: Failed to create HTTP request: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}
	var pinterestAppSecret string = os.Getenv("PINTEREST_SECRET")
	req.Header.Add("Authorization", "Basic "+pinterestAppSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		app.Logger().Error("Pinterest: Failed to send HTTP request: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		app.Logger().Error("Pinterest: Failed to read response body: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}

	var response PinterestResponse

	err = json.Unmarshal(body, &response)
	if err != nil {
		app.Logger().Error("Pinterest: Failed to parse JSON response: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}

	boards, err := GetUserBoards(response.AccessToken, app)

	if err != nil {
		app.Logger().Error("Pinterest: Failed to get user boards: " + err.Error())
		helpers.Error(e, err.Error())
		return
	}

	for _, board := range boards.Items {
		userData := models.Connections{
			UserId:         authUserId,
			Name:           board.Name,
			ConnectionName: "pinterest",
			ConnectionId:   board.ID,
			AccessToken:    response.AccessToken,
			MetaData:       string(body),
			ProfileImage:   board.Media.ImageCoverURL,
			Username:       board.Owner.Username,
			RefreshToken:   response.RefreshToken,
		}

		err = AddNewConnection(app, &userData)

		if err != nil {
			app.Logger().Error("Pinterest: Failed to add Pinterest page connection: " + err.Error())
			helpers.Error(e, err.Error())
			return
		}
	}

	helpers.Success(e, "Pages connected", map[string]interface{}{})
	return
}

func GetUserBoards(accessToken string, app *pocketbase.PocketBase) (PinResponse, error) {
	url := "https://api.pinterest.com/v5/boards"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {

		return PinResponse{}, nil
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)

	res, err := client.Do(req)
	if err != nil {
		app.Logger().Error("Pinterest: Failed to send HTTP request: " + err.Error())
		return PinResponse{}, nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		app.Logger().Error("Pinterest: Failed to read response body: " + err.Error())
		return PinResponse{}, nil
	}

	var profile PinResponse

	// Unmarshal the JSON response into the struct
	err = json.Unmarshal(body, &profile)
	if err != nil {
		app.Logger().Error("Pinterest: Failed to parse JSON response: " + err.Error())
		return PinResponse{}, nil
	}

	return profile, nil
}
