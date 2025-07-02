package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pocketbase/pocketbase"
)

// Universal HTTP request function
func MakeHTTPRequest[T any](
	app *pocketbase.PocketBase,
	method string,
	fullURL string,
	headers map[string]string,
	queryParams url.Values,
	body interface{},
) (T, error) {
	var result T

	var bodyReader io.Reader

	// Prepare request body based on Content-Type
	if body != nil {
		contentType := headers["Content-Type"]

		switch contentType {
		case "application/x-www-form-urlencoded":
			formValues, ok := body.(url.Values)
			if !ok {
				return result, fmt.Errorf("body must be url.Values when using application/x-www-form-urlencoded")
			}
			bodyReader = strings.NewReader(formValues.Encode())

		case "application/json", "":
			b, err := json.Marshal(body)
			if err != nil {
				return result, err
			}
			bodyReader = bytes.NewBuffer(b)

		default:
			return result, fmt.Errorf("unsupported Content-Type: %s", contentType)
		}
	}

	// Add query parameters
	u, err := url.Parse(fullURL)
	if err != nil {
		return result, err
	}
	if len(queryParams) > 0 {
		q := u.Query()
		for k, v := range queryParams {
			q[k] = v
		}
		u.RawQuery = q.Encode()
	}

	// Create HTTP request
	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return result, err
	}

	// Set headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if body != nil && headers["Content-Type"] == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	// Read and decode response
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	app.Logger().Debug("HTTP Request", "url", u.String(), "body", string(respBytes))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return result, &httpError{Status: resp.Status, Body: string(respBytes)}
	}

	// Try to unmarshal the response into result
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return result, err
	}

	return result, nil
}

type httpError struct {
	Status string
	Body   string
}

func (e *httpError) Error() string {
	return e.Status + ": " + e.Body
}
