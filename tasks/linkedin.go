package tasks

import (
	"bytes"
	"content-clock/helpers"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pocketbase/pocketbase"
)

type LinkedinResponse struct {
	ID string `json:"id"`
}

func HandleLinkedinProfilePostTask(app *pocketbase.PocketBase, p PostToSocialPayload) error {

	content := p.Content
	images := p.Images
	connectionId := p.ConnectionId
	accessToken := p.AccessToken
	socialPostId := p.SocialPostId

	url := "https://api.linkedin.com/v2/ugcPosts"

	escapedContent := strconv.QuoteToASCII(content)

	body := ""

	if len(images) > 0 {

		uploadRegisResp, err := LinkedinRegisterUpload(connectionId, accessToken)
		if err != nil {
			FailedPost(app, "linkedin", socialPostId, err)
			return err
		}

		uploadUrl := uploadRegisResp.Value.UploadMechanism.MediaUploadHttpRequest.UploadUrl

		if uploadUrl == "" {

			FailedPost(app, "linkedin", socialPostId, errors.New("upload url is empty"))
			return err
		}
		backendHost := os.Getenv("API_HOST")
		imageUrl := fmt.Sprintf("%s/api/files/posts/%s/%s", backendHost, socialPostId, images[0])
		err = linkedinUploadMedia(imageUrl, uploadUrl, accessToken)
		if err != nil {
			FailedPost(app, "linkedin", socialPostId, err)
			return err
		}
		time.Sleep(time.Second * 5)

		assetId := uploadRegisResp.Value.Asset

		body = `{
			"author": "urn:li:person:` + connectionId + `",
			"lifecycleState": "PUBLISHED",
			"specificContent": {
				"com.linkedin.ugc.ShareContent": {
					"shareCommentary": {
						"text": ` + escapedContent + `
					},
					"shareMediaCategory": "IMAGE",
					"media": [
						{
							"status": "READY",
							"description": {
								"text": "LinkedIn Upload"
							},
							"media": "` + assetId + `",
							"title": {
								"text": "LinkedIn Upload"
							}
						}
					]
				}
			},
			"visibility": {
				"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC"
			}
		}`

	} else {

		body = `{
			"author": "urn:li:person:` + connectionId + `",
			"lifecycleState": "PUBLISHED",
			"specificContent": {
				"com.linkedin.ugc.ShareContent": {
					"shareCommentary": {
						"text": ` + escapedContent + `
					},
					"shareMediaCategory": "NONE"
				}
			},
			"visibility": {
				"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC"
			}
		}`

	}

	requestBody := []byte(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		FailedPost(app, "linkedin", socialPostId, err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")

	// Create an HTTP client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		FailedPost(app, "linkedin", socialPostId, err)
		return err
	}
	defer resp.Body.Close()

	// Read and print the response body
	responseBody := new(bytes.Buffer)
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		FailedPost(app, "linkedin", socialPostId, err)
		return err
	}

	var lresp LinkedinResponse
	err = json.Unmarshal([]byte(responseBody.String()), &lresp)
	if err != nil {
		FailedPost(app, "linkedin", socialPostId, err)
		return err
	}

	if lresp.ID == "" {
		FailedPost(app, "linkedin", socialPostId, errors.New(responseBody.String()))
		return err
	}

	SuccessPost(app, "linkedin", socialPostId, lresp.ID)

	app.Logger().Info("Linkedin post created successfully", "postId", lresp.ID)
	return nil
}

func linkedinUploadMedia(image, uploadUrl, accessToken string) error {

	fileLocation, err := helpers.DownloadImage(image, true)
	if err != nil {
		return err
	}

	// Open the file to be uploaded
	file, err := os.Open(fileLocation)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new HTTP request with the file as the request body
	req, err := http.NewRequest("POST", uploadUrl, file)
	if err != nil {
		return err
	}

	// Set the required headers, including the authorization header
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Send the request
	uploadResp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer uploadResp.Body.Close()

	// Print the response body, if needed
	responseBody := new(bytes.Buffer)
	_, err = responseBody.ReadFrom(uploadResp.Body)
	if err != nil {
		return err
	}
	return nil
}

type UploadRegisterResponse struct {
	Value struct {
		UploadMechanism struct {
			MediaUploadHttpRequest struct {
				Headers   map[string]interface{} `json:"headers"`
				UploadUrl string                 `json:"uploadUrl"`
			} `json:"com.linkedin.digitalmedia.uploading.MediaUploadHttpRequest"`
		} `json:"uploadMechanism"`
		Asset string `json:"asset"`
	} `json:"value"`
}

func LinkedinRegisterUpload(connectionId string, accessToken string) (UploadRegisterResponse, error) {
	url := "https://api.linkedin.com/v2/assets?action=registerUpload"

	body := `{
		"registerUploadRequest": {
			"recipes": [
				"urn:li:digitalmediaRecipe:feedshare-image"
			],
			"owner": "urn:li:person:` + connectionId + `",
			"serviceRelationships": [
				{
					"relationshipType": "OWNER",
					"identifier": "urn:li:userGeneratedContent"
				}
			]
		}
	}`

	requestBody := []byte(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return UploadRegisterResponse{}, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	// req.Header.Set("X-Restli-Protocol-Version", "2.0.0")

	// Create an HTTP client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return UploadRegisterResponse{}, err
	}
	defer resp.Body.Close()

	// Read and print the response body
	responseBody := new(bytes.Buffer)
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		return UploadRegisterResponse{}, err
	}

	var lresp UploadRegisterResponse
	err = json.Unmarshal([]byte(responseBody.String()), &lresp)
	if err != nil {
		return UploadRegisterResponse{}, err
	}
	return lresp, nil

}
