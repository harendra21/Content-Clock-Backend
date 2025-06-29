package controllers

import (
	"bytes"
	"content-clock/helpers"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const (
	chatGptEndpoint = "https://openrouter.ai/api/v1"
	modelName       = "mistralai/mistral-small-3.2-24b-instruct-2506:free" // https://openrouter.ai/models
)

// Request body structure
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream,omitempty"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int `json:"index"`
		Message      ChatMessage
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage map[string]int `json:"usage"`
}

func SetupAiRoutes(se *core.ServeEvent, app *pocketbase.PocketBase) {
	se.Router.GET("/api/v1/post-with-ai", func(e *core.RequestEvent) error {
		GenratePostWithAi(e, app)
		return nil
	})
}

func GenratePostWithAi(e *core.RequestEvent, app *pocketbase.PocketBase) {
	topic := e.Request.URL.Query().Get("topic")
	apiKey := os.Getenv("CHAT_GPT_KEY")
	if topic == "" {
		helpers.Error(e, "Topic is required")
		return
	}
	postType := e.Request.URL.Query().Get("type")
	if postType == "" {
		postType = "write"
	}
	fieldName := e.Request.URL.Query().Get("field")
	if fieldName == "" {
		fieldName = "content"
	}

	topic = fmt.Sprintf("%s post %s for - %s. Do not include single or double quotes, just write a post that is short (2-3 lines), engaging and informative, add hashtags.", postType, fieldName, topic)

	reqBody := ChatRequest{
		Model: modelName,
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a helpful assistant that generates or edit social media posts for given data."},
			{Role: "user", Content: topic},
		},
		Stream: false,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		app.Logger().Error("Failed to marshal request body", "error", err)
		helpers.Error(e, "Failed to prepare request for AI service")
		return
	}

	req, err := http.NewRequest("POST", chatGptEndpoint+"/chat/completions", bytes.NewBuffer(payload))
	if err != nil {
		app.Logger().Error("Failed to create request", "error", err)
		helpers.Error(e, "Failed to create request to AI service")
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		app.Logger().Error("Failed to send request", "error", err)
		helpers.Error(e, "Failed to connect to AI service")
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		app.Logger().Error("Failed to read response body", "error", err)
		helpers.Error(e, "Failed to read response from AI")
		return
	}

	var chatResp ChatResponse
	err = json.Unmarshal(bodyBytes, &chatResp)
	if err != nil {
		app.Logger().Error("Failed to unmarshal response", "error", err, "body", string(bodyBytes))
		helpers.Error(e, "Failed to parse response from AI")
		return
	}

	// Print the assistant's reply
	if len(chatResp.Choices) > 0 {
		app.Logger().Info("AI Response", "response", chatResp.Choices[0].Message.Content)
		helpers.Success(e, "", chatResp.Choices[0].Message.Content)
		return
	} else {
		app.Logger().Error("No choices found in response", "response", string(bodyBytes))
		helpers.Error(e, "No response from AI")
		return
	}

}
