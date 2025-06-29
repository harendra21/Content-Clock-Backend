package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
)

func HandlePinterestBoardPostTask(app *pocketbase.PocketBase, p PostToSocialPayload) error {
	content := p.Content
	images := p.Images
	connectionId := p.ConnectionId
	accessToken := p.AccessToken
	socialPostId := p.SocialPostId
	title := p.Title

	url := "https://api.pinterest.com/v5/pins"
	method := "POST"

	backendHost := os.Getenv("API_HOST")
	imageUrl := fmt.Sprintf("%s/api/files/posts/%s/%s", backendHost, socialPostId, images[0])

	payload := strings.NewReader(`{
      "title": "` + title + `",
      "description": "` + content + `",
      "board_id": "` + connectionId + `",
      "media_source": {
        "source_type": "image_url",
        "url": "` + imageUrl + `"
      }
  }`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		FailedPost(app, "pinterest", socialPostId, err)
		return err
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		FailedPost(app, "pinterest", socialPostId, err)
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		FailedPost(app, "pinterest", socialPostId, err)
		return err
	}

	post := make(map[string]interface{})

	err = json.Unmarshal(body, &post)
	if err != nil {
		FailedPost(app, "pinterest", socialPostId, err)
		return err

	}

	if post["id"] != nil {
		SuccessPost(app, "pinterest", socialPostId, post["id"].(string))
		return nil
	} else {
		FailedPost(app, "pinterest", socialPostId, errors.New(string(body)))
		return err
	}

}
