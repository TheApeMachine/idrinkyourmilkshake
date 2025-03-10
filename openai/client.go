package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/davecgh/go-spew/spew"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/theapemachine/idrinkyourmilkshake/browser"
	"github.com/theapemachine/idrinkyourmilkshake/models"
	"github.com/theapemachine/idrinkyourmilkshake/mongodb"
	"github.com/theapemachine/idrinkyourmilkshake/request"
	"github.com/theapemachine/idrinkyourmilkshake/utils"
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
func (c *Client) ProcessToolCall(
	toolCall openai.ChatCompletionMessageToolCall,
	params *openai.ChatCompletionNewParams,
) error {
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

// Replace the existing getToolExecutor method with this more concise version

// getToolExecutor returns the appropriate executor for a tool
func (c *Client) getToolExecutor(toolName string, args map[string]any) (ToolExecutor, map[string]any, error) {
	// Define tool configurations
	toolConfigs := map[string]struct {
		factory    func() ToolExecutor
		paramCheck func(args map[string]any) ([]any, error)
	}{
		"extract_page_content": {
			factory: func() ToolExecutor { return &browser.BrowserExtractor{} },
			paramCheck: func(args map[string]any) ([]any, error) {
				return []any{}, nil
			},
		},
		"browser_navigate": {
			factory: func() ToolExecutor { return &browser.BrowserNavigator{} },
			paramCheck: func(args map[string]any) ([]any, error) {
				url, ok := args["url"].(string)
				if !ok {
					return nil, fmt.Errorf("url parameter is required")
				}
				return []any{"url", url}, nil
			},
		},
		// ... similar entries for other tools
	}

	// Look up the tool configuration
	config, exists := toolConfigs[toolName]
	if !exists {
		log.Error("Unknown tool called", "tool", toolName)
		return nil, nil, fmt.Errorf("unknown tool: %s", toolName)
	}

	// Check parameters and get log details
	params, err := config.paramCheck(args)
	if err != nil {
		log.Error(err.Error())
		return nil, nil, err
	}

	return config.factory(), c.getStatusMessages(toolName, params), nil
}

func (c *Client) Execute(
	buffer *Buffer,
	maxIterations int,
) (string, error) {
	log.Info("Starting OpenAI client execution", "maxIterations", maxIterations)

	schema := utils.GenerateSchema[models.APIConfig]()

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("api_config"),
		Description: openai.F("The API configuration"),
		Schema:      openai.F(schema),
		Strict:      openai.Bool(false),
	}

	availableTools := []models.ToolType{
		models.NewTool(browser.NewBrowserExtractor()),
		models.NewTool(browser.NewBrowserNavigator()),
		models.NewTool(browser.NewBrowserJavaScriptExecutor()),
		models.NewTool(browser.NewBrowserClicker()),
		models.NewTool(request.NewHTTPRequest()),
		models.NewTool(mongodb.NewMongoDBInspector()),
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
		Model:       openai.F(openai.ChatModelGPT4o),
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

		log.Info("Messages", "messages", len(params.Messages.Value))
		spew.Dump(params)
		completion, err := c.client.Chat.Completions.New(c.ctx, params)

		if err != nil {
			log.Error("OpenAI API error", "error", err)
			return "", fmt.Errorf("OpenAI API error: %w", err)
		}

		toolCalls := completion.Choices[0].Message.ToolCalls

		if len(toolCalls) == 0 {
			log.Info("No tool calls requested, returning final result")
			continue
		}

		params.Messages.Value = append(params.Messages.Value, completion.Choices[0].Message)

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
