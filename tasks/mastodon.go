package tasks

import (
	"bytes"
	"content-clock/helpers"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase"
)

func HandlePostToMastodon(app *pocketbase.PocketBase, p PostToSocialPayload) error {

	content := p.Content
	images := p.Images
	// connectionId := p.ConnectionId
	accessToken := p.AccessToken
	socialPostId := p.SocialPostId

	backendHost := os.Getenv("API_HOST")
	var mediaIDs []string

	for _, image := range images {
		imageUrl := fmt.Sprintf("%s/api/files/posts/%s/%s", backendHost, socialPostId, image)
		mediaID, err := UploadMedia(accessToken, imageUrl)
		if err != nil {
			FailedPost(app, "mastodon", socialPostId, err)
			return err
		}
		mediaIDs = append(mediaIDs, mediaID)
	}

	body, err := PostStatus(accessToken, content, mediaIDs)
	if err != nil {
		FailedPost(app, "mastodon", socialPostId, err)
		return err
	}

	SuccessPost(app, "mastodon", socialPostId, body)

	return nil
}

func PostStatus(accessToken string, content string, mediaIDs []string) (string, error) {
	data := map[string]interface{}{
		"status":     content,
		"media_ids":  mediaIDs,
		"visibility": "public", // Optional: public, unlisted, private, direct
	}

	body, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	mastodonInstance := os.Getenv("MASTODON_BASE_URL")
	req, err := http.NewRequest("POST", mastodonInstance+"/api/v1/statuses", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	bodyString := string(respBody)
	if resp.StatusCode >= 300 {

		return "", fmt.Errorf("post failed: %s", string(respBody))
	}
	return bodyString, nil
}

func UploadMedia(accessToken string, imageUrl string) (string, error) {
	imagePath, err := helpers.DownloadImage(imageUrl, true)

	if err != nil {
		return "", err
	}

	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(imagePath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}

	// Optional: alt text
	_ = writer.WriteField("description", "Uploaded via Go")

	if err := writer.Close(); err != nil {
		return "", err
	}

	mastodonInstance := os.Getenv("MASTODON_BASE_URL")
	req, err := http.NewRequest("POST", mastodonInstance+"/api/v1/media", &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("media upload failed: %s", string(body))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.ID, nil
}
