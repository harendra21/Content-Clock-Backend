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

func FacebookPagePost(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string, link string) {

	data := PostToSocialPayload{
		Content:      postContent,
		Link:         link,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	// StartQueue("facebook", data)
	HandleFacebookPagePostTask(app, data)

}

func LinkedinPost(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string) {

	data := PostToSocialPayload{
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	// StartQueue("linkedin", data)
	HandleLinkedinProfilePostTask(app, data)

}

func PostToTwitterProfile(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string) {

	data := PostToSocialPayload{
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	// StartQueue("twitter", data)
	HandleTwitterPostTask(app, data)

}
func InstagramPost(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string) {

	data := PostToSocialPayload{
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	// StartQueue("instagram", data)
	HandleInstagramPostTask(app, data)
}

func PostToPinterestBoard(app *pocketbase.PocketBase, title, postContent string, images []string, connectionId string, accessToken string, socialPostId string) {

	data := PostToSocialPayload{
		Title:        title,
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	// StartQueue("pinterest", data)
	HandlePinterestBoardPostTask(app, data)

}

func PostToDiscordChannel(postContent string, images []string, connectionId string, accessToken string, socialPostId string) {

	data := PostToSocialPayload{
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	// StartQueue("discord", data)
	HandleDiscordPostTask(data)

}

func PostToMastodon(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string) {

	data := PostToSocialPayload{
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	app.Logger().Info("Postinh to mastodon", "data", data)

	// StartQueue("discord", data)
	HandlePostToMastodon(app, data)

}

func PostToThreads(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string) {

	data := PostToSocialPayload{
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	app.Logger().Info("Postinh to threads", "data", data)

	// StartQueue("discord", data)
	HandlePostToThreads(app, data)

}
func PostToReddit(app *pocketbase.PocketBase, postContent string, images []string, connectionId string, accessToken string, socialPostId string) {

	data := PostToSocialPayload{
		Content:      postContent,
		Images:       images,
		ConnectionId: connectionId,
		AccessToken:  accessToken,
		SocialPostId: socialPostId,
	}

	app.Logger().Info("Postinh to reddit", "data", data)

	// StartQueue("discord", data)
	HandlePostToReddit(app, data)

}

func FailedPost(app *pocketbase.PocketBase, platform string, postId string, err error) {
	record, _ := app.FindRecordById("posts", postId)
	record.Set("status", "failed")
	record.Set("logs", err.Error())
	app.Save(record)
	app.Logger().Error("Failed to post on "+platform, "type", "posting", "platform", platform, "postId", postId, "error", err.Error())
}

func SuccessPost(app *pocketbase.PocketBase, platform string, postId string, publishedPostId string) {
	record, _ := app.FindRecordById("posts", postId)
	record.Set("status", "published")
	record.Set("published_post_id", publishedPostId)
	app.Save(record)
	app.Logger().Info("Successfully posted on "+platform, "type", "posting", "platform", platform, "postId", postId, "publishedPostId", publishedPostId)
}
