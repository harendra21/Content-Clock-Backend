package tasks

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase"
)

var reddit string = "https://graph.threads.net/v1.0"

type RedditResponse struct {
	ID string `json:"id"`
}

func HandlePostToReddit(app *pocketbase.PocketBase, p PostToSocialPayload) error {
	content := p.Content
	images := p.Images
	connectionId := p.ConnectionId
	accessToken := p.AccessToken
	socialPostId := p.SocialPostId
	// backendHost := os.Getenv("API_HOST")

	app.Logger().Info("Posting to reddit", "connectionId", connectionId, "content", content, "images", images)

	subreddit := ""
	title := ""
	text := content
	token := accessToken

	if subreddit == "" || title == "" || token == "" {
		FailedPost(app, "reddit", socialPostId, errors.New("All params required"))
		return nil
	}

	req, _ := http.NewRequest("POST", "https://oauth.reddit.com/api/submit", strings.NewReader("sr="+subreddit+"&title="+title+"&text="+text+"&kind=self"))
	req.Header.Set("Authorization", "bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		FailedPost(app, "reddit", socialPostId, err)
		return nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	SuccessPost(app, "reddit", socialPostId, "")
	return nil

	return nil

}
