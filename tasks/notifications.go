package tasks

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func createPostNotification(app *pocketbase.PocketBase, postRecord *core.Record, notificationType string, platform string, details string) {
	if app == nil || postRecord == nil {
		return
	}

	userID := postRecord.GetString("user")
	if userID == "" {
		return
	}

	collection, err := app.FindCollectionByNameOrId("notifications")
	if err != nil {
		app.Logger().Warn("notifications collection missing; skipping notification", "error", err.Error())
		return
	}

	postID := postRecord.Id
	connectionID := postRecord.GetString("connection")
	title := postRecord.GetString("title")
	content := postRecord.GetString("content")

	if title == "" {
		title = trimNotificationText(content, 60)
	}

	var notificationTitle string
	var message string
	switch notificationType {
	case "post_published":
		notificationTitle = "Post published"
		message = fmt.Sprintf("%s post published successfully on %s.", titleOrFallback(title), platformLabel(platform))
	case "post_failed":
		notificationTitle = "Post failed"
		message = fmt.Sprintf("%s failed to publish on %s.", titleOrFallback(title), platformLabel(platform))
	default:
		notificationTitle = "Post update"
		message = fmt.Sprintf("%s status updated on %s.", titleOrFallback(title), platformLabel(platform))
	}

	if strings.TrimSpace(details) != "" {
		message += " " + trimNotificationText(details, 180)
	}

	record := core.NewRecord(collection)
	record.Set("user", userID)
	record.Set("post", postID)
	record.Set("connection", connectionID)
	record.Set("type", notificationType)
	record.Set("title", notificationTitle)
	record.Set("message", message)
	record.Set("created_at", time.Now())

	if err := app.Save(record); err != nil {
		app.Logger().Error("Failed to save notification", "postId", postID, "type", notificationType, "error", err.Error())
	}
}

func platformLabel(platform string) string {
	platform = strings.TrimSpace(platform)
	if platform == "" {
		return "platform"
	}
	runes := []rune(platform)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func titleOrFallback(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return "Your post"
	}
	return `"` + trimNotificationText(title, 70) + `"`
}

func trimNotificationText(text string, maxLen int) string {
	clean := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(text, "\n", " "), "\r", " "))
	if maxLen <= 0 || len(clean) <= maxLen {
		return clean
	}
	return clean[:maxLen-3] + "..."
}
