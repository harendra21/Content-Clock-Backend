package tasks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase"
)

func HandleInstagramPostTask(app *pocketbase.PocketBase, p PostToSocialPayload) error {
	content := p.Content
	images := p.Images
	connectionId := p.ConnectionId
	accessToken := p.AccessToken
	socialPostId := p.SocialPostId

	app.Logger().Info("Posting to Instagram", "connectionId", connectionId, "content", content, "images", images)

	backendHost := os.Getenv("API_HOST")
	imageCount := len(images)

	if imageCount == 0 {
		err := errors.New("No images for Instagram")
		FailedPost(app, "instagram", socialPostId, err)
		return err
	}

	if imageCount == 1 {

		imageURL := fmt.Sprintf("%s/api/files/posts/%s/%s", backendHost, socialPostId, images[0])
		postBody := map[string]string{
			"image_url": imageURL,
			"caption":   content,
		}

		url := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/media?access_token=%s", connectionId, accessToken)
		jsonData, err := json.Marshal(postBody)
		if err != nil {
			FailedPost(app, "instagram", socialPostId, err)
			return err
		}

		res, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			FailedPost(app, "instagram", socialPostId, err)
			return err
		}
		defer res.Body.Close()

		body, _ := ioutil.ReadAll(res.Body)
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			FailedPost(app, "instagram", socialPostId, err)
			return err
		}

		if result["id"] == nil {
			errMsg := "Failed to create single image post"
			if e, ok := result["error"]; ok {
				errMsg = e.(map[string]interface{})["message"].(string)
			}
			FailedPost(app, "instagram", socialPostId, errors.New(errMsg))
			return errors.New(errMsg)
		}

		postID := result["id"].(string)
		err = PublishMedia(connectionId, postID, accessToken)
		if err != nil {
			FailedPost(app, "instagram", socialPostId, err)
			return err
		}

		SuccessPost(app, "instagram", socialPostId, postID)
		return nil
	}

	if imageCount < 2 || imageCount > 10 {
		err := fmt.Errorf("carousel posts require 2 to 10 images. Got: %d", imageCount)
		FailedPost(app, "instagram", socialPostId, err)
		return err
	}

	// === CAROUSEL POST ===

	var creationIDs []string

	for _, img := range images {
		imageURL := fmt.Sprintf("%s/api/files/posts/%s/%s", backendHost, socialPostId, img)

		postBody := map[string]string{
			"image_url":        imageURL,
			"is_carousel_item": "true",
		}

		url := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/media?access_token=%s", connectionId, accessToken)
		jsonData, err := json.Marshal(postBody)
		if err != nil {
			FailedPost(app, "instagram", socialPostId, err)
			return err
		}

		res, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			FailedPost(app, "instagram", socialPostId, err)
			return err
		}
		defer res.Body.Close()

		body, _ := ioutil.ReadAll(res.Body)
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			FailedPost(app, "instagram", socialPostId, err)
			return err
		}

		if result["id"] == nil {
			errMsg := "Failed to get media container ID"
			if e, ok := result["error"]; ok {
				errMsg = e.(map[string]interface{})["message"].(string)
			}
			FailedPost(app, "instagram", socialPostId, errors.New(errMsg))
			return errors.New(errMsg)
		}

		creationIDs = append(creationIDs, result["id"].(string))
	}

	carouselPostBody := map[string]interface{}{
		"caption":    content,
		"children":   creationIDs,
		"media_type": "CAROUSEL",
	}

	carouselJson, err := json.Marshal(carouselPostBody)
	if err != nil {
		FailedPost(app, "instagram", socialPostId, err)
		return err
	}

	carouselURL := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/media?access_token=%s", connectionId, accessToken)
	res, err := http.Post(carouselURL, "application/json", bytes.NewBuffer(carouselJson))
	if err != nil {
		FailedPost(app, "instagram", socialPostId, err)
		return err
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)
	var carouselResp map[string]interface{}
	if err := json.Unmarshal(body, &carouselResp); err != nil {
		FailedPost(app, "instagram", socialPostId, err)
		return err
	}

	if carouselResp["id"] == nil {
		errMsg := "Failed to create carousel container"
		if e, ok := carouselResp["error"]; ok {
			errMsg = e.(map[string]interface{})["message"].(string)
		}
		FailedPost(app, "instagram", socialPostId, errors.New(errMsg))
		return errors.New(errMsg)
	}

	postID := carouselResp["id"].(string)
	err = PublishMedia(connectionId, postID, accessToken)
	if err != nil {
		FailedPost(app, "instagram", socialPostId, err)
		return err
	}

	SuccessPost(app, "instagram", socialPostId, postID)
	return nil
}

func PublishMedia(connectionId, postId, accessToken string) error {

	url := "https://graph.facebook.com/v19.0/" + connectionId + "/media_publish?creation_id=" + postId + "&access_token=" + accessToken

	response, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	post := make(map[string]interface{})

	err = json.Unmarshal(body, &post)
	if err != nil {
		return err
	}
	if _, ok := post["error"]; ok {
		errMsg := post["error"].(map[string]interface{})["message"].(string)
		return errors.New(errMsg)
	} else {
		return nil
	}

}
