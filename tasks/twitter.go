package tasks

import (
	"bytes"
	"content-clock/helpers"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"strconv"
	"strings"

	"github.com/dghubble/oauth1"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/tweet/managetweet"
	"github.com/michimani/gotwi/tweet/managetweet/types"
	"github.com/pocketbase/pocketbase"
)

type TwitterResponse struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

type MediaUpload struct {
	MediaId int `json:"media_id"`
}

func HandleTwitterPostTask(app *pocketbase.PocketBase, p PostToSocialPayload) error {

	content := p.Content
	images := p.Images
	accessToken := p.AccessToken
	socialPostId := p.SocialPostId

	tokens := strings.Split(accessToken, " ")

	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           tokens[0],
		OAuthTokenSecret:     tokens[1],
	}

	client, err := gotwi.NewClient(in)
	if err != nil {
		FailedPost(app, "twitter", socialPostId, err)
		return err
	}

	var post types.CreateInput // Remove the pointer
	post.Text = gotwi.String(content)

	if len(images) > 0 {
		mediaIds := []string{}
		for _, image := range images {
			backendHost := os.Getenv("API_HOST")
			imageUrl := fmt.Sprintf("%s/api/files/posts/%s/%s", backendHost, socialPostId, image)
			mediaId, err := uploadTwitterMedia(imageUrl, tokens[0], tokens[1])
			if err != nil {
				FailedPost(app, "twitter", socialPostId, err)
				return err
			}
			mediaIds = append(mediaIds, mediaId)
		}

		mediaInput := &types.CreateInputMedia{
			MediaIDs: mediaIds,
		}
		post.Media = mediaInput
	}

	res, err := managetweet.Create(context.Background(), client, &post)
	if err != nil {
		FailedPost(app, "twitter", socialPostId, err)
		return err
	}

	tweetId := gotwi.StringValue(res.Data.ID)
	SuccessPost(app, "twitter", socialPostId, tweetId)
	return nil

}

func uploadTwitterMedia(path string, oauthToken string, oauthTokenSecret string) (string, error) {

	config := oauth1.NewConfig(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"))

	var token oauth1.Token

	token.Token = oauthToken
	token.TokenSecret = oauthTokenSecret

	// httpClient will automatically authorize http.Request's
	httpClient := config.Client(oauth1.NoContext, &token)
	b := &bytes.Buffer{}
	form := multipart.NewWriter(b)

	// resp, err := http.Get(path)
	// if err != nil {
	// 	helpers.Logging("error", err.Error())
	// 	return "", err
	// }
	// defer resp.Body.Close()

	fileLocation, err := helpers.DownloadImage(path, true)
	if err != nil {
		helpers.Logging("error", err.Error())
		return "", err
	}

	// Open the file to be uploaded
	file, err := os.Open(fileLocation)
	if err != nil {
		helpers.Logging("error", err.Error())
		return "", err
	}
	defer file.Close()

	fw, err := form.CreateFormFile("media", fileLocation)
	if err != nil {
		helpers.Logging("error", err.Error())
		return "", err
	}

	// copy to form
	_, err = io.Copy(fw, file)
	if err != nil {
		helpers.Logging("error", err.Error())
		return "", err
	}

	// close form
	form.Close()

	// example Twitter API request
	uploadResp, err := httpClient.Post("https://upload.twitter.com/1.1/media/upload.json?media_category=tweet_image", form.FormDataContentType(), bytes.NewReader(b.Bytes()))
	if err != nil {
		helpers.Logging("error", err.Error())
		return "", err
	}

	body, err := ioutil.ReadAll(uploadResp.Body)
	if err != nil {
		helpers.Logging("error", err.Error())
		return "", err
	}

	m := &MediaUpload{}
	if err := json.Unmarshal(body, &m); err != nil {
		helpers.Logging("error", err.Error())
		return "", err
	}

	mid := strconv.Itoa(m.MediaId)

	return mid, nil

}
