package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/theapemachine/idrinkyourmilkshake/browser"
	"github.com/theapemachine/idrinkyourmilkshake/models"
	"github.com/theapemachine/idrinkyourmilkshake/request"
)

// Client wraps the OpenAI API client with additional functionality
type Client struct {
	client *openai.Client
	ctx    context.Context
}

// NewClient creates a new OpenAI client with the given API key
func NewClient(apiKey string) *Client {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &Client{
		client: client,
		ctx:    context.Background(),
	}
}

// WithContext sets the context for the client
func (c *Client) WithContext(ctx context.Context) *Client {
	c.ctx = ctx
	return c
}

// ToolExecutor is the interface for all tool executors
type ToolExecutor interface {
	Execute(args map[string]any) (string, error)
}

// ProcessToolCall handles a single tool call
func (c *Client) ProcessToolCall(toolCall openai.ChatCompletionMessageToolCall, params *openai.ChatCompletionNewParams) error {
	// Parse the arguments
	var args map[string]any
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		log.Error("Error parsing arguments", "error", err)
		return fmt.Errorf("error parsing arguments: %w", err)
	}

	// Get the appropriate executor
	executor, logDetails, err := c.getToolExecutor(toolCall.Function.Name, args)
	if err != nil {
		return err
	}

	// Log the action being performed
	log.Info(logDetails["start_message"].(string), logDetails["params"].([]any)...)

	// Execute the tool
	content, err := executor.Execute(args)
	if err != nil {
		log.Error("Error executing tool", "tool", toolCall.Function.Name, "error", err)
		return fmt.Errorf("error executing tool: %w", err)
	}

	// Log success
	log.Info(logDetails["success_message"].(string))

	// Add the tool call result to the conversation
	params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, content))

	return nil
}

func (c *Client) getStatusMessages(toolName string, params []any) map[string]any {
	return map[string]any{
		"start_message":   "Running " + toolName,
		"success_message": toolName + " completed successfully",
		"params":          params,
	}
}

// getToolExecutor returns the appropriate executor for a tool
func (c *Client) getToolExecutor(toolName string, args map[string]any) (ToolExecutor, map[string]any, error) {
	switch toolName {
	case "extract_page_content":
		return &browser.BrowserExtractor{}, c.getStatusMessages(toolName, []any{}), nil

	case "browser_navigate":
		url, ok := args["url"].(string)
		if !ok {
			log.Error("URL parameter is missing")
			return nil, nil, fmt.Errorf("url parameter is required")
		}
		return &browser.BrowserNavigator{}, c.getStatusMessages(toolName, []any{"url", url}), nil

	case "browser_execute_js":
		script, ok := args["script"].(string)
		if !ok {
			log.Error("Script parameter is missing")
			return nil, nil, fmt.Errorf("script parameter is required")
		}
		return &browser.BrowserJavaScriptExecutor{}, c.getStatusMessages(toolName, []any{"script", script}), nil

	case "browser_click":
		selector, ok := args["selector"].(string)
		if !ok {
			log.Error("Selector parameter is missing")
			return nil, nil, fmt.Errorf("selector parameter is required")
		}
		return &browser.BrowserClicker{}, c.getStatusMessages(toolName, []any{"selector", selector}), nil

	case "http_request":
		method, _ := args["method"].(string)
		url, _ := args["url"].(string)
		return &request.HTTPRequest{}, c.getStatusMessages(toolName, []any{"method", method, "url", url}), nil

	default:
		log.Error("Unknown tool called", "tool", toolName)
		return nil, nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

func (c *Client) Execute(
	buffer *Buffer,
	maxIterations int,
) (string, error) {
	log.Info("Starting OpenAI client execution", "maxIterations", maxIterations)

	// Create a simplified schema for APIConfig that follows OpenAI's expectations
	apiConfigSchema := map[string]interface{}{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]interface{}{
			"integration": map[string]interface{}{
				"type":        "string",
				"description": "The name of the integration",
			},
			"account_id": map[string]interface{}{
				"type":        "string",
				"description": "The account ID",
			},
			"base_url": map[string]interface{}{
				"type":        "string",
				"description": "The base URL",
			},
			"jobs": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Name of the job",
						},
					},
					"required": []string{"name"},
				},
			},
		},
		"required": []string{"integration", "account_id", "base_url", "jobs"},
	}

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("api_config"),
		Description: openai.F("The API configuration"),
		Schema:      openai.F(any(apiConfigSchema)),
		Strict:      openai.Bool(true),
	}

	availableTools := []models.ToolType{
		models.NewTool(browser.NewBrowserExtractor()),
		models.NewTool(browser.NewBrowserNavigator()),
		models.NewTool(browser.NewBrowserJavaScriptExecutor()),
		models.NewTool(browser.NewBrowserClicker()),
		models.NewTool(request.NewHTTPRequest()),
	}

	tools := []openai.ChatCompletionToolParam{}

	for _, tool := range availableTools {
		tools = append(tools, openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(openai.FunctionDefinitionParam{
				Name:        openai.String(tool.Name()),
				Description: openai.String(tool.Description()),
				Parameters:  openai.F(schemaToFunctionParameters(tool.Schema())),
			}),
		})
	}

	params := openai.ChatCompletionNewParams{
		Model:       openai.F(openai.ChatModelGPT4oMini),
		Messages:    openai.F(buffer.Truncate().Messages),
		Tools:       openai.F(tools),
		Temperature: openai.F(0.0),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(schemaParam),
			},
		),
	}

	// Iterate until the model stops requesting tool calls
	for i := range maxIterations {
		log.Info("Executing iteration", "iteration", i+1, "of", maxIterations)

		completion, err := c.client.Chat.Completions.New(c.ctx, params)
		if err != nil {
			log.Error("OpenAI API error", "error", err)
			return "", fmt.Errorf("OpenAI API error: %w", err)
		}

		// If no tool calls are requested, return the final result
		toolCalls := completion.Choices[0].Message.ToolCalls
		if len(toolCalls) == 0 {
			log.Info("No tool calls requested, returning final result")
			return completion.Choices[0].Message.Content, nil
		}

		log.Info("Processing tool calls", "count", len(toolCalls))

		// Add the assistant's message to the conversation
		params.Messages.Value = append(params.Messages.Value, completion.Choices[0].Message)

		// Process each tool call
		for _, toolCall := range toolCalls {
			if err := c.ProcessToolCall(toolCall, &params); err != nil {
				return "", err
			}
		}
	}

	log.Error("Maximum iterations reached without resolution", "maxIterations", maxIterations)
	return "", fmt.Errorf("reached maximum iterations without resolution")
}

// schemaToFunctionParameters converts a jsonschema.Schema to the format expected by the OpenAI API
func schemaToFunctionParameters(schema any) openai.FunctionParameters {
	// Convert schema to map[string]any
	bytes, err := json.Marshal(schema)
	if err != nil {
		log.Error("Failed to marshal schema", "error", err)
		return openai.FunctionParameters{}
	}

	var params openai.FunctionParameters
	if err := json.Unmarshal(bytes, &params); err != nil {
		log.Error("Failed to unmarshal schema to FunctionParameters", "error", err)
		return openai.FunctionParameters{}
	}

	return params
}

// convertSchemaToMap converts a jsonschema.Schema to a map[string]interface{}
func convertSchemaToMap(schema any) (map[string]interface{}, error) {
	bytes, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	var schemaMap map[string]interface{}
	if err := json.Unmarshal(bytes, &schemaMap); err != nil {
		return nil, err
	}

	return schemaMap, nil
}
