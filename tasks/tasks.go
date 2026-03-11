package tasks

import (
	"github.com/pocketbase/pocketbase"
)

const (
	PostToFacebook  = "postto:facebook"
	PostToTwitter   = "postto:twitter"
	PostToInstagram = "postto:instagram"
	PostToLinkedin  = "postto:linkedin"
	PostToPinterest = "postto:pinterest"
	PostToDiscord   = "postto:discord"
)

type PostToSocialPayload struct {
	Title        string
	Content      string
	Link         string
	Images       []string
	ConnectionId string
	AccessToken  string
	SocialPostId string
}

func FacebookPagePost(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string, link string) error {

	data := PostToSocialPayload{
		Content:      postContent,
		Link:         link,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	// StartQueue("facebook", data)
	return HandleFacebookPagePostTask(app, data)
}

func LinkedinPost(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string) error {

	data := PostToSocialPayload{
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	// StartQueue("linkedin", data)
	return HandleLinkedinProfilePostTask(app, data)
}

func PostToTwitterProfile(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string) error {

	data := PostToSocialPayload{
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	// StartQueue("twitter", data)
	return HandleTwitterPostTask(app, data)
}
func InstagramPost(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string) error {

	data := PostToSocialPayload{
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	// StartQueue("instagram", data)
	return HandleInstagramPostTask(app, data)
}

func PostToPinterestBoard(app *pocketbase.PocketBase, title, postContent string, images []string, connectionId string, accessToken string, socialPostId string) error {

	data := PostToSocialPayload{
		Title:        title,
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	// StartQueue("pinterest", data)
	return HandlePinterestBoardPostTask(app, data)
}

func PostToDiscordChannel(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string) error {

	data := PostToSocialPayload{
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	// StartQueue("discord", data)
	return HandleDiscordPostTask(app, data)

}

func PostToMastodon(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string) error {

	data := PostToSocialPayload{
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	app.Logger().Info("Postinh to mastodon", "data", data)

	// StartQueue("discord", data)
	return HandlePostToMastodon(app, data)

}

func PostToThreads(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string) error {

	data := PostToSocialPayload{
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	app.Logger().Info("Postinh to threads", "data", data)

	// StartQueue("discord", data)
	return HandlePostToThreads(app, data)

}
func PostToReddit(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string) error {

	data := PostToSocialPayload{
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	app.Logger().Info("Postinh to reddit", "data", data)

	// StartQueue("discord", data)
	return HandlePostToReddit(app, data)

}

func FailedPost(app *pocketbase.PocketBase, platform string, postId string, err error) {
	record, findErr := app.FindRecordById("posts", postId)
	if findErr != nil {
		app.Logger().Error("Failed to load post for failed status update", "postId", postId, "error", findErr.Error())
		return
	}
	previousStatus := record.GetString("status")
	record.Set("status", "failed")
	record.Set("logs", err.Error())
	if saveErr := app.Save(record); saveErr != nil {
		app.Logger().Error("Failed to update post status to failed", "postId", postId, "error", saveErr.Error())
		return
	}
	if previousStatus != "failed" {
		createPostNotification(app, record, "post_failed", platform, err.Error())
	}
	app.Logger().Error("Failed to post on "+platform, "type", "posting", "platform", platform, "postId", postId, "error", err.Error())
}

func SuccessPost(app *pocketbase.PocketBase, platform string, postId string, publishedPostId string) {
	record, findErr := app.FindRecordById("posts", postId)
	if findErr != nil {
		app.Logger().Error("Failed to load post for published status update", "postId", postId, "error", findErr.Error())
		return
	}
	previousStatus := record.GetString("status")
	record.Set("status", "published")
	record.Set("published_post_id", publishedPostId)
	record.Set("logs", "")
	if saveErr := app.Save(record); saveErr != nil {
		app.Logger().Error("Failed to update post status to published", "postId", postId, "error", saveErr.Error())
		return
	}
	if previousStatus != "published" {
		createPostNotification(app, record, "post_published", platform, "")
	}
	app.Logger().Info("Successfully posted on "+platform, "type", "posting", "platform", platform, "postId", postId, "publishedPostId", publishedPostId)
}
