package request

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/idrinkyourmilkshake/models"
)

type HTTPRequest struct {
	ToolName        string           `json:"name" jsonschema:"description=The name of the tool,required"`
	ToolDescription string           `json:"description" jsonschema:"description=The description of the tool,required"`
	ToolParameters  models.Parameter `json:"parameters" jsonschema:"description=The parameters of the tool,required"`
	Required        []string         `json:"required" jsonschema:"description=The required parameters of the tool,required"`
}

func NewHTTPRequest() models.ToolType {
	return &HTTPRequest{
		ToolName:        "http_request",
		ToolDescription: "Makes an HTTP request to the specified URL",
		ToolParameters: models.Parameter{
			Type: "object",
			Properties: []models.Property{
				{
					Name:        "url",
					Type:        "string",
					Description: "The URL to request",
				},
				{
					Name:        "body",
					Type:        "string",
					Description: "The body of the request",
				},
				{
					Name:        "headers",
					Type:        "object",
					Description: "The headers of the request",
				},
			},
			Required: true,
		},
		Required: []string{"url"},
	}
}

func (h *HTTPRequest) Name() string {
	return h.ToolName
}

func (h *HTTPRequest) Description() string {
	return h.ToolDescription
}

func (h *HTTPRequest) Execute(args map[string]any) (string, error) {
	log.Info("Starting HTTP request execution")

	// Extract request parameters from args
	method, ok := args["method"].(string)
	if !ok {
		method = "GET" // Default to GET if not specified
		log.Info("No method specified, defaulting to GET")
	}

	url, ok := args["url"].(string)
	if !ok {
		log.Error("URL is required but not provided")
		return "", fmt.Errorf("url is required")
	}

	var bodyStr string
	body, ok := args["body"].(string)
	if ok {
		bodyStr = body
		log.Info("Request body provided", "size", len(bodyStr))
	} else {
		log.Info("No request body provided")
	}

	var headers map[string]string
	headersMap, ok := args["headers"].(map[string]any)
	if ok {
		headers = make(map[string]string)
		for k, v := range headersMap {
			if strVal, ok := v.(string); ok {
				headers[k] = strVal
			}
		}
		log.Info("Request headers provided", "count", len(headers))
	} else {
		log.Info("No request headers provided")
	}

	// Create HTTP client and request
	log.Info("Creating HTTP request", "method", method, "url", url)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, strings.NewReader(bodyStr))
	if err != nil {
		log.Error("Error creating HTTP request", "error", err)
		return "", err
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	log.Info("Sending HTTP request")
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error executing HTTP request", "error", err)
		return "", err
	}
	defer resp.Body.Close()

	log.Info("Received HTTP response", "status", resp.Status, "statusCode", resp.StatusCode)

	// Read response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error reading response body", "error", err)
		return "", err
	}

	responseSize := len(bodyBytes)
	log.Info("Successfully read response body", "size", responseSize)

	// Check if the status code indicates success (200-299)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error("Request failed with non-success status code", "statusCode", resp.StatusCode)
		return "", fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return string(bodyBytes), nil
}

func (h *HTTPRequest) Schema() any {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "The URL to request",
			},
			"body": map[string]interface{}{
				"type":        "string",
				"description": "The body of the request",
			},
			"headers": map[string]interface{}{
				"type":        "object",
				"description": "The headers of the request",
			},
		},
		"required": []string{"url"},
	}
}
