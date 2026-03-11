package controllers

import "strings"

// Centralized OAuth scopes used while generating platform tokens.
// Keep these minimal and aligned with actual API calls in tasks/* and analytics.go.
const (
	facebookScopePagesManagePosts    = "pages_manage_posts"
	facebookScopePagesShowList       = "pages_show_list"
	facebookScopePagesReadEngagement = "pages_read_engagement"
	facebookScopeReadInsights        = "read_insights"
	facebookScopePublishVideo        = "publish_video"

	instagramScopeBasic          = "instagram_basic"
	instagramScopeContentPublish = "instagram_content_publish"
	instagramScopeManageInsights = "instagram_manage_insights"

	threadsScopeBasic          = "threads_basic"
	threadsScopeContentPublish = "threads_content_publish"
	threadsScopeManageInsights = "threads_manage_insights"
)

func FacebookOAuthScopes() []string {
	return []string{
		facebookScopePagesManagePosts,
		facebookScopePagesShowList,
		facebookScopePagesReadEngagement,
		facebookScopeReadInsights,
		facebookScopePublishVideo,
	}
}

func InstagramOAuthScopes() []string {
	// pages_show_list is required to fetch pages and linked IG business accounts.
	return []string{
		instagramScopeBasic,
		instagramScopeContentPublish,
		instagramScopeManageInsights,
		facebookScopePagesShowList,
		facebookScopePagesReadEngagement,
	}
}

func LinkedinOAuthScopes() []string {
	// w_member_social is required for posting; openid/profile/email used for userinfo identity.
	return []string{"w_member_social", "openid", "profile", "email"}
}

func PinterestOAuthScopes() []string {
	return []string{"boards:read", "pins:read", "user_accounts:read", "pins:write"}
}

func RedditOAuthScopes() string {
	return "identity submit read"
}

func ThreadsOAuthScopes() string {
	return strings.Join([]string{
		threadsScopeBasic,
		threadsScopeContentPublish,
		threadsScopeManageInsights,
	}, ",")
}

func MastodonOAuthScopes() string {
	// follow removed; app only needs read/write in current implementation.
	return "read write"
}
