package controllers

import (
	"bytes"
	"content-clock/helpers"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const (
	chatGptEndpoint = "https://openrouter.ai/api/v1"
	modelName       = "inception/mercury-2" // https://openrouter.ai/models
	imageModelName  = "sourceful/riverflow-v2-fast"
	defaultAppName  = "Content Clock"
)

// Request body structure
type ChatMessage struct {
	Role    string      `json:"role"`
	Content string      `json:"content"`
	Images  []ChatImage `json:"images,omitempty"`
}

type ChatImage struct {
	Type     string `json:"type"`
	ImageURL struct {
		URL string `json:"url"`
	} `json:"image_url"`
}

type ChatRequest struct {
	Model          string        `json:"model"`
	Messages       []ChatMessage `json:"messages"`
	Stream         bool          `json:"stream,omitempty"`
	ResponseFormat interface{}   `json:"response_format,omitempty"`
	Modalities     []string      `json:"modalities,omitempty"`
	ImageConfig    *ImageConfig  `json:"image_config,omitempty"`
}

type ImageConfig struct {
	AspectRatio string      `json:"aspect_ratio,omitempty"`
	ImageSize   string      `json:"image_size,omitempty"`
	FontInputs  []FontInput `json:"font_inputs,omitempty"`
}

type FontInput struct {
	FontURL string `json:"font_url"`
	Text    string `json:"text"`
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
	Usage map[string]interface{} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type StructuredOutput struct {
	Output  string   `json:"output,omitempty"`
	Outputs []string `json:"outputs,omitempty"`
}

type ImagePromptOutput struct {
	Prompt string `json:"prompt"`
}

func SetupAiRoutes(se *core.ServeEvent, app *pocketbase.PocketBase) {
	se.Router.GET("/api/v1/post-with-ai", func(e *core.RequestEvent) error {
		GenratePostWithAi(e, app)
		return nil
	})
	se.Router.GET("/api/v1/image-prompt-with-ai", func(e *core.RequestEvent) error {
		GenerateImagePromptWithAi(e, app)
		return nil
	})
	se.Router.GET("/api/v1/image-with-ai", func(e *core.RequestEvent) error {
		GenerateImageWithAi(e, app)
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
	count := parseCountParam(e.Request.URL.Query().Get("count"))

	topic = fmt.Sprintf(
		"%s post %s for - %s. Generate exactly %d distinct options. Each option should be short (2-3 lines), engaging, informative, and include relevant hashtags. Do not include single or double quotes.",
		postType,
		fieldName,
		topic,
		count,
	)

	reqBody := ChatRequest{
		Model: modelName,
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a helpful assistant that generates or edit social media posts for given data."},
			{Role: "user", Content: topic},
		},
		Stream: false,
		ResponseFormat: map[string]interface{}{
			"type": "json_schema",
			"json_schema": map[string]interface{}{
				"name":   "post_response",
				"strict": true,
				"schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"output": map[string]interface{}{
							"type": "string",
						},
						"outputs": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "string",
							},
							"minItems": count,
							"maxItems": count,
						},
					},
					"required":             []string{"outputs"},
					"additionalProperties": false,
				},
			},
		},
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
	setOpenRouterHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		app.Logger().Error("Failed to send request", "error", err)
		helpers.Error(e, "Failed to connect to AI service")
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
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

	if chatResp.Error != nil && chatResp.Error.Message != "" {
		app.Logger().Error("AI provider returned error", "error", chatResp.Error.Message)
		helpers.Error(e, chatResp.Error.Message)
		return
	}

	if len(chatResp.Choices) > 0 {
		rawContent := chatResp.Choices[0].Message.Content
		parsedContent, parseErr := parseStructuredOutput(rawContent)
		if parseErr != nil {
			app.Logger().Warn("Failed to parse structured AI response, falling back to raw content", "error", parseErr.Error(), "raw", rawContent)
			trimmed := strings.TrimSpace(rawContent)
			if count <= 1 {
				helpers.Success(e, "", trimmed)
				return
			}
			helpers.Success(e, "", []string{trimmed})
			return
		}

		if count <= 1 {
			app.Logger().Info("AI Response", "response", parsedContent[0])
			helpers.Success(e, "", parsedContent[0])
			return
		}

		app.Logger().Info("AI Response", "responseCount", len(parsedContent))
		helpers.Success(e, "", parsedContent)
		return
	} else {
		app.Logger().Error("No choices found in response", "response", string(bodyBytes))
		helpers.Error(e, "No response from AI")
		return
	}

}

func GenerateImageWithAi(e *core.RequestEvent, app *pocketbase.PocketBase) {
	prompt := strings.TrimSpace(e.Request.URL.Query().Get("prompt"))
	if prompt == "" {
		helpers.Error(e, "Prompt is required")
		return
	}
	prompt = buildImageOnlyPrompt(prompt)

	apiKey := os.Getenv("CHAT_GPT_KEY")
	if strings.TrimSpace(apiKey) == "" {
		helpers.Error(e, "OpenRouter API key is missing")
		return
	}
	model := strings.TrimSpace(e.Request.URL.Query().Get("model"))
	if model == "" {
		model = imageModelName
	}

	fontInputs, err := parseFontInputs(e)
	if err != nil {
		helpers.Error(e, err.Error())
		return
	}
	if len(fontInputs) > 0 && !isSourcefulModel(model) {
		helpers.Error(e, "font_inputs are supported only for sourceful models")
		return
	}

	aspectRatio := strings.TrimSpace(e.Request.URL.Query().Get("aspect_ratio"))
	if aspectRatio == "" {
		aspectRatio = "1:1"
	}
	imageSize := strings.TrimSpace(e.Request.URL.Query().Get("image_size"))
	if imageSize == "" {
		imageSize = "1K"
	}

	reqBody := ChatRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "user", Content: prompt},
		},
		Modalities: []string{"image"},
		Stream:     false,
		ImageConfig: &ImageConfig{
			AspectRatio: aspectRatio,
			ImageSize:   imageSize,
			FontInputs:  fontInputs,
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		app.Logger().Error("Failed to marshal image request body", "error", err)
		helpers.Error(e, "Failed to prepare image generation request")
		return
	}

	req, err := http.NewRequest("POST", chatGptEndpoint+"/chat/completions", bytes.NewBuffer(payload))
	if err != nil {
		app.Logger().Error("Failed to create image request", "error", err)
		helpers.Error(e, "Failed to create image generation request")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	setOpenRouterHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		app.Logger().Error("Failed to send image generation request", "error", err)
		helpers.Error(e, "Failed to connect to AI image service")
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Logger().Error("Failed to read image generation response body", "error", err)
		helpers.Error(e, "Failed to read response from AI image service")
		return
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		app.Logger().Error("Failed to parse image generation response", "error", err, "body", string(bodyBytes))
		helpers.Error(e, "Failed to parse image generation response")
		return
	}

	if chatResp.Error != nil && chatResp.Error.Message != "" {
		app.Logger().Error("AI image provider returned error", "error", chatResp.Error.Message)
		helpers.Error(e, chatResp.Error.Message)
		return
	}

	if len(chatResp.Choices) == 0 || len(chatResp.Choices[0].Message.Images) == 0 {
		app.Logger().Error("No image found in AI response", "response", string(bodyBytes))
		helpers.Error(e, "No image returned from AI")
		return
	}

	imageURL := strings.TrimSpace(chatResp.Choices[0].Message.Images[0].ImageURL.URL)
	if imageURL == "" {
		helpers.Error(e, "Generated image URL is empty")
		return
	}

	helpers.Success(e, "", map[string]string{
		"image_url": imageURL,
	})
}

func GenerateImagePromptWithAi(e *core.RequestEvent, app *pocketbase.PocketBase) {
	topic := strings.TrimSpace(e.Request.URL.Query().Get("topic"))
	if topic == "" {
		helpers.Error(e, "Topic is required")
		return
	}

	apiKey := os.Getenv("CHAT_GPT_KEY")
	if strings.TrimSpace(apiKey) == "" {
		helpers.Error(e, "OpenRouter API key is missing")
		return
	}

	reqBody := ChatRequest{
		Model: modelName,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You create high-quality image generation prompts for social media visuals. Keep prompts concise and visual. Ensure result contains no text, captions, logos or watermarks.",
			},
			{
				Role:    "user",
				Content: "Create one image generation prompt based on: " + topic,
			},
		},
		Stream: false,
		ResponseFormat: map[string]interface{}{
			"type": "json_schema",
			"json_schema": map[string]interface{}{
				"name":   "image_prompt_response",
				"strict": true,
				"schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"prompt": map[string]interface{}{
							"type": "string",
						},
					},
					"required":             []string{"prompt"},
					"additionalProperties": false,
				},
			},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		app.Logger().Error("Failed to marshal image prompt request", "error", err)
		helpers.Error(e, "Failed to prepare image prompt request")
		return
	}

	req, err := http.NewRequest("POST", chatGptEndpoint+"/chat/completions", bytes.NewBuffer(payload))
	if err != nil {
		app.Logger().Error("Failed to create image prompt request", "error", err)
		helpers.Error(e, "Failed to create image prompt request")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	setOpenRouterHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		app.Logger().Error("Failed to send image prompt request", "error", err)
		helpers.Error(e, "Failed to connect to AI prompt service")
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Logger().Error("Failed to read image prompt response body", "error", err)
		helpers.Error(e, "Failed to read response from AI prompt service")
		return
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		app.Logger().Error("Failed to parse image prompt response", "error", err, "body", string(bodyBytes))
		helpers.Error(e, "Failed to parse image prompt response")
		return
	}

	if chatResp.Error != nil && chatResp.Error.Message != "" {
		app.Logger().Error("AI prompt provider returned error", "error", chatResp.Error.Message)
		helpers.Error(e, chatResp.Error.Message)
		return
	}

	if len(chatResp.Choices) == 0 {
		helpers.Error(e, "No prompt returned from AI")
		return
	}

	rawContent := chatResp.Choices[0].Message.Content
	var output ImagePromptOutput
	if err := json.Unmarshal([]byte(rawContent), &output); err != nil {
		app.Logger().Warn("Failed to parse structured image prompt response, using raw content", "error", err.Error())
		output.Prompt = strings.TrimSpace(rawContent)
	}

	output.Prompt = buildImageOnlyPrompt(strings.TrimSpace(output.Prompt))
	if output.Prompt == "" {
		helpers.Error(e, "Generated prompt is empty")
		return
	}

	helpers.Success(e, "", map[string]string{
		"prompt": output.Prompt,
	})
}

func buildImageOnlyPrompt(prompt string) string {
	instruction := "Generate only an image scene. Do not include any text, letters, words, captions in the image."
	return strings.TrimSpace(prompt) + " " + instruction
}

func isSourcefulModel(model string) bool {
	return model == "sourceful/riverflow-v2-fast" || model == "sourceful/riverflow-v2-pro"
}

func parseFontInputs(e *core.RequestEvent) ([]FontInput, error) {
	inputs := make([]FontInput, 0, 2)
	for index := 1; index <= 2; index++ {
		fontURL := strings.TrimSpace(e.Request.URL.Query().Get(fmt.Sprintf("font_url_%d", index)))
		text := strings.TrimSpace(e.Request.URL.Query().Get(fmt.Sprintf("font_text_%d", index)))
		if fontURL == "" && text == "" {
			continue
		}
		if fontURL == "" || text == "" {
			return nil, fmt.Errorf("font_url_%d and font_text_%d are required together", index, index)
		}
		inputs = append(inputs, FontInput{
			FontURL: fontURL,
			Text:    text,
		})
	}
	return inputs, nil
}

func parseStructuredOutput(raw string) ([]string, error) {
	var parsed StructuredOutput
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}

	if len(parsed.Outputs) > 0 {
		cleaned := make([]string, 0, len(parsed.Outputs))
		for _, item := range parsed.Outputs {
			trimmed := strings.TrimSpace(item)
			if trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
		if len(cleaned) > 0 {
			return cleaned, nil
		}
	}

	if strings.TrimSpace(parsed.Output) != "" {
		return []string{strings.TrimSpace(parsed.Output)}, nil
	}

	return nil, fmt.Errorf("structured response output is empty")
}

func parseCountParam(raw string) int {
	if raw == "" {
		return 1
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 1
	}
	if value < 1 {
		return 1
	}
	if value > 10 {
		return 10
	}
	return value
}

func setOpenRouterHeaders(req *http.Request) {
	appName := strings.TrimSpace(os.Getenv("OPENROUTER_APP_NAME"))
	if appName == "" {
		appName = strings.TrimSpace(os.Getenv("APP_NAME"))
	}
	if appName == "" {
		appName = defaultAppName
	}

	siteURL := strings.TrimSpace(os.Getenv("OPENROUTER_SITE_URL"))
	if siteURL == "" {
		siteURL = strings.TrimSpace(os.Getenv("REDIRECT_HOST"))
	}
	if siteURL == "" {
		siteURL = strings.TrimSpace(os.Getenv("API_HOST"))
	}

	req.Header.Set("X-Title", appName)
	if siteURL != "" {
		req.Header.Set("HTTP-Referer", siteURL)
	}
}
