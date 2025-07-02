package tasks

import (
	"content-clock/helpers"
	"fmt"
	"net/url"
	"os"

	"github.com/pocketbase/pocketbase"
)

var threadsUrl string = "https://graph.threads.net/v1.0"

type ThreadsResponse struct {
	ID string `json:"id"`
}

func HandlePostToThreads(app *pocketbase.PocketBase, p PostToSocialPayload) error {
	content := p.Content
	images := p.Images
	connectionId := p.ConnectionId
	accessToken := p.AccessToken
	socialPostId := p.SocialPostId
	backendHost := os.Getenv("API_HOST")
	// backendHost = "https://content-clock.loca.lt"

	app.Logger().Info("Posting to threads", "connectionId", connectionId, "content", content, "images", images)

	if len(images) == 0 {
		params := url.Values{}
		params.Add("media_type", "TEXT")
		params.Add("text", content)
		params.Add("access_token", accessToken)
		reqUrl := fmt.Sprintf("%s/%s/threads", threadsUrl, connectionId)
		resp, err := helpers.MakeHTTPRequest[ThreadsResponse](app, "POST", reqUrl, nil, params, nil)
		if err != nil {
			FailedPost(app, "threads", socialPostId, err)
			return err
		}

		publishParams := url.Values{}
		publishParams.Add("creation_id", resp.ID)
		publishParams.Add("access_token", accessToken)

		reqUrl = fmt.Sprintf("%s/%s/threads_publish", threadsUrl, connectionId)
		resp, err = helpers.MakeHTTPRequest[ThreadsResponse](app, "POST", reqUrl, nil, publishParams, nil)
		if err != nil {
			FailedPost(app, "threads", socialPostId, err)
			return err
		}
		SuccessPost(app, "threads", socialPostId, resp.ID)
	} else if len(images) == 1 {
		imageUrl := fmt.Sprintf("%s/api/files/posts/%s/%s", backendHost, socialPostId, images[0])

		params := url.Values{}
		params.Add("media_type", "IMAGE")
		params.Add("is_carousel_item", "false")
		params.Add("text", content)
		params.Add("image_url", imageUrl)
		params.Add("access_token", accessToken)

		reqUrl := fmt.Sprintf("%s/%s/threads", threadsUrl, connectionId)
		resp, err := helpers.MakeHTTPRequest[ThreadsResponse](app, "POST", reqUrl, nil, params, nil)
		if err != nil {
			FailedPost(app, "threads", socialPostId, err)
			return err
		}

		publishParams := url.Values{}
		publishParams.Add("creation_id", resp.ID)
		publishParams.Add("access_token", accessToken)
		reqUrl = fmt.Sprintf("%s/%s/threads_publish", threadsUrl, connectionId)
		resp, err = helpers.MakeHTTPRequest[ThreadsResponse](app, "POST", reqUrl, nil, publishParams, nil)
		if err != nil {
			FailedPost(app, "threads", socialPostId, err)
			return err
		}
		SuccessPost(app, "threads", socialPostId, resp.ID)
	} else if len(images) > 1 {
		var children string
		for _, image := range images {
			imageUrl := fmt.Sprintf("%s/api/files/posts/%s/%s", backendHost, socialPostId, image)

			params := url.Values{}
			params.Add("media_type", "IMAGE")
			params.Add("is_carousel_item", "true")
			params.Add("image_url", imageUrl)
			params.Add("text", content)
			params.Add("access_token", accessToken)

			url := fmt.Sprintf("%s/%s/threads", threadsUrl, connectionId)
			resp, err := helpers.MakeHTTPRequest[ThreadsResponse](app, "POST", url, nil, params, nil)
			if err != nil {
				FailedPost(app, "threads", socialPostId, err)
				return err
			}
			children = children + resp.ID + ","

		}

		params := url.Values{}
		params.Add("media_type", "CAROUSEL")
		params.Add("children", children)
		params.Add("access_token", accessToken)

		reqUrl := fmt.Sprintf("%s/%s/threads", threadsUrl, connectionId)
		resp, err := helpers.MakeHTTPRequest[ThreadsResponse](app, "POST", reqUrl, nil, params, nil)
		if err != nil {
			FailedPost(app, "threads", socialPostId, err)
			return err
		}

		paramsPublish := url.Values{}
		paramsPublish.Add("creation_id", resp.ID)
		paramsPublish.Add("access_token", accessToken)

		reqUrl = fmt.Sprintf("%s/%s/threads_publish", threadsUrl, connectionId)
		resp, err = helpers.MakeHTTPRequest[ThreadsResponse](app, "POST", reqUrl, nil, paramsPublish, nil)
		if err != nil {
			FailedPost(app, "threads", socialPostId, err)
			return err
		}
		SuccessPost(app, "threads", socialPostId, resp.ID)

	}

	return nil

}
