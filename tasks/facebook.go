package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
)

func HandleFacebookPagePostTask(app *pocketbase.PocketBase, p PostToSocialPayload) error {
	content := p.Content
	images := p.Images
	connectionId := p.ConnectionId
	accessToken := p.AccessToken
	socialPostId := p.SocialPostId
	link := p.Link

	app.Logger().Info("Posting to Facebook", "connectionId", connectionId, "content", content, "images", images)

	postType := "feed"
	postBody := map[string]string{
		"message":      content,
		"access_token": accessToken,
		"link":         link,
		"published":    "true",
	}
	backendHost := os.Getenv("API_HOST")
	// backendHost := "https://content-clock.loca.lt"

	var mediaIds []string
	if len(images) > 0 {
		for _, image := range images {
			imageUrl := fmt.Sprintf("%s/api/files/posts/%s/%s", backendHost, socialPostId, image)
			id, err := UploadImge(accessToken, connectionId, imageUrl)
			if err != nil {
				FailedPost(app, "facebook", socialPostId, err)
				return err
			}
			mediaIds = append(mediaIds, id)
		}
	}
	if len(mediaIds) > 0 {
		for i, media := range mediaIds {
			key := fmt.Sprintf("attached_media[%d]", i)
			value := fmt.Sprintf(`{"media_fbid":"%s"}`, media)
			postBody[key] = value
		}
	}

	url := fmt.Sprintf("https://graph.facebook.com/%s/feed?access_token=%s", connectionId, postType)
	method := "POST"

	payload := EncodePostBody(postBody)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		FailedPost(app, "facebook", socialPostId, err)
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		FailedPost(app, "facebook", socialPostId, err)
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		FailedPost(app, "facebook", socialPostId, err)
		return err
	}

	post := make(map[string]interface{})
	err = json.Unmarshal(body, &post)
	if err != nil {
		FailedPost(app, "facebook", socialPostId, err)
		return err
	}

	app.Logger().Info("Facebook post body", "body", string(body))

	if errObj, ok := post["error"]; ok {
		errMsg := "Unknown error"
		if errMap, ok := errObj.(map[string]interface{}); ok {
			if msg, ok := errMap["message"].(string); ok {
				errMsg = msg
			}
		}
		FailedPost(app, "facebook", socialPostId, errors.New(errMsg))
		return errors.New(errMsg)
	} else {
		postId, _ := post["id"].(string)
		SuccessPost(app, "facebook", socialPostId, postId)
	}

	return nil
}

func EncodePostBody(postBody map[string]string) *strings.Reader {
	form := url.Values{}
	for key, value := range postBody {
		form.Set(key, value)
	}
	return strings.NewReader(form.Encode())
}

func UploadImge(accessToken, connectionId, imagePath string) (string, error) {
	url := fmt.Sprintf("https://graph.facebook.com/%s/photos", connectionId)
	payload := strings.NewReader(fmt.Sprintf("url=%s&published=false&access_token=%s", imagePath, accessToken))

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	// Check for Facebook Graph API error
	var errResp struct {
		Error struct {
			Message        string `json:"message"`
			Type           string `json:"type"`
			Code           int    `json:"code"`
			ErrorSubcode   int    `json:"error_subcode"`
			ErrorUserTitle string `json:"error_user_title"`
			ErrorUserMsg   string `json:"error_user_msg"`
			IsTransient    bool   `json:"is_transient"`
			FbtraceID      string `json:"fbtrace_id"`
		} `json:"error"`
	}

	if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
		return "", fmt.Errorf("facebook error: %s - %s (code %d, subcode %d)", errResp.Error.ErrorUserTitle, errResp.Error.ErrorUserMsg, errResp.Error.Code, errResp.Error.ErrorSubcode)
	}

	// Parse normal success response
	var respData struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &respData); err != nil {
		return "", fmt.Errorf("failed to parse success response: %w", err)
	}

	if respData.ID == "" {
		return "", fmt.Errorf("upload failed: no id returned, raw body: %s", string(body))
	}

	return respData.ID, nil
}
