package tasks

import (
	"bytes"
	"content-clock/helpers"
	"content-clock/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HandleDiscordPostTask handles the actual posting to Discord using OAuth
func HandleDiscordPostTask(p PostToSocialPayload) error {

	content := p.Content
	images := p.Images
	channelID := p.ConnectionId // Discord channel ID
	oauthToken := p.AccessToken // Discord OAuth access token
	socialPostId := p.SocialPostId

	// Verify bot token first
	verifyURL := "https://discord.com/api/v10/users/@me"
	verifyReq, err := http.NewRequest("GET", verifyURL, nil)
	if err != nil {
		helpers.Logging("error", fmt.Sprintf("Failed to create verification request: %v", err))
		return err
	}
	verifyReq.Header.Set("Authorization", fmt.Sprintf("Bot %s", oauthToken))

	client := &http.Client{}
	verifyResp, err := client.Do(verifyReq)
	if err != nil {
		helpers.Logging("error", fmt.Sprintf("Failed to verify bot token: %v", err))
		return err
	}
	defer verifyResp.Body.Close()

	verifyBody, _ := io.ReadAll(verifyResp.Body)
	helpers.Logging("info", fmt.Sprintf("Bot verification response: %s", string(verifyBody)))

	if verifyResp.StatusCode != http.StatusOK {
		helpers.Logging("error", fmt.Sprintf("Bot token verification failed: %s", string(verifyBody)))
		models.DB.Model(&models.SocialPosts{}).Where("id = ?", socialPostId).Update("status", "failed").Update("logs", "Bot token verification failed")
		return fmt.Errorf("bot token verification failed")
	}

	helpers.Logging("info", fmt.Sprintf("Posting to Discord (OAuth): content='%s', images=%v, channelID=%s, socialPostId=%v", content, images, channelID, socialPostId))
	helpers.Logging("info", fmt.Sprintf("Using Discord token: %s...", oauthToken[:10])) // Log first 10 chars of token

	// Build Discord API payload
	payload := map[string]interface{}{
		"content": content,
	}
	if len(images) > 0 {
		imageUrl := fmt.Sprintf("https://fvqhirsubytaubzgrecj.supabase.co/storage/v1/object/public/post-images/%s", images[0])
		payload["embeds"] = []map[string]interface{}{
			{"image": map[string]string{"url": imageUrl}},
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		helpers.Logging("error", fmt.Sprintf("Failed to marshal Discord payload: %v", err))
		models.DB.Model(&models.SocialPosts{}).Where("id = ?", socialPostId).Update("status", "failed").Update("logs", err.Error())
		return err
	}

	apiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		helpers.Logging("error", fmt.Sprintf("Failed to create Discord API request: %v", err))
		models.DB.Model(&models.SocialPosts{}).Where("id = ?", socialPostId).Update("status", "failed").Update("logs", err.Error())
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", oauthToken))
	req.Header.Set("Content-Type", "application/json")

	helpers.Logging("info", fmt.Sprintf("Discord API request URL: %s", apiURL))
	helpers.Logging("info", fmt.Sprintf("Discord API request headers: %v", req.Header))
	helpers.Logging("info", fmt.Sprintf("Discord API request body: %s", string(jsonData)))

	client = &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		helpers.Logging("error", fmt.Sprintf("HTTP POST to Discord API failed: %v", err))
		models.DB.Model(&models.SocialPosts{}).Where("id = ?", socialPostId).Update("status", "failed").Update("logs", err.Error())
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		helpers.Logging("error", fmt.Sprintf("Failed to read Discord API response body: %v", err))
		models.DB.Model(&models.SocialPosts{}).Where("id = ?", socialPostId).Update("status", "failed").Update("logs", err.Error())
		return err
	}
	helpers.Logging("info", fmt.Sprintf("Discord API response: %s", string(body)))

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		helpers.Logging("info", "Discord post published successfully (OAuth)")
		// Discord returns a message object with an ID
		var respData map[string]interface{}
		_ = json.Unmarshal(body, &respData)
		msgID := ""
		if id, ok := respData["id"].(string); ok {
			msgID = id
		}
		models.DB.Model(&models.SocialPosts{}).Where("id = ?", socialPostId).Update("status", "published").Update("published_post_id", msgID)
	} else {
		helpers.Logging("error", fmt.Sprintf("Discord API error: %s", string(body)))
		models.DB.Model(&models.SocialPosts{}).Where("id = ?", socialPostId).Update("status", "failed").Update("logs", string(body))
	}

	return nil
}
