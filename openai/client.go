package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/davecgh/go-spew/spew"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/theapemachine/idrinkyourmilkshake/browser"
	"github.com/theapemachine/idrinkyourmilkshake/models"
	"github.com/theapemachine/idrinkyourmilkshake/mongodb"
	"github.com/theapemachine/idrinkyourmilkshake/utils"
)

// Client wraps the OpenAI API client with additional functionality
type Client struct {
	client *openai.Client
	ctx    context.Context
	tools  map[string]models.ToolType
}

// NewClient creates a new OpenAI client with the given API key
func NewClient(apiKey string) *Client {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &Client{
		client: &client,
		ctx:    context.Background(),
		tools:  make(map[string]models.ToolType),
	}
}

// WithContext sets the context for the client
func (c *Client) WithContext(ctx context.Context) *Client {
	c.ctx = ctx
	return c
}

// RegisterTools registers tools with the client
func (c *Client) RegisterTools(tools []models.ToolType) {
	for _, tool := range tools {
		c.tools[tool.Name()] = tool
	}
}

// CreateCompletionParams creates OpenAI completion parameters with tools
func (c *Client) CreateCompletionParams(systemPrompt, userPrompt string) *openai.ChatCompletionNewParams {
	toolParams := []openai.ChatCompletionToolParam{}

	for _, tool := range c.tools {
		toolParams = append(toolParams, openai.ChatCompletionToolParam{
			Type: "function",
			Function: openai.FunctionDefinitionParam{
				Name:        tool.Name(),
				Description: param.NewOpt(tool.Description()),
				Parameters:  schemaToFunctionParameters(tool.Schema()),
			},
		})
	}

	// Generate schema for APIConfig
	schema := utils.GenerateSchema[models.APIConfig]()
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "api_config",
		Description: param.NewOpt("The API configuration"),
		Schema:      param.NewOpt(schema),
		Strict:      param.NewOpt(true),
	}

	return &openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4o,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		Tools: toolParams,
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				Type:       "json_schema",
				JSONSchema: schemaParam,
			},
		},
	}
}

// ExecuteWithTools executes the OpenAI completion with tool support
func (c *Client) ExecuteWithTools(params *openai.ChatCompletionNewParams, maxIterations int) (string, error) {
	for i := 0; i < maxIterations; i++ {
		log.Info("Executing iteration", "iteration", i+1)

		completion, err := c.client.Chat.Completions.New(c.ctx, *params)
		if err != nil {
			log.Error("OpenAI API error", "error", err)
			return "", fmt.Errorf("OpenAI API error: %w", err)
		}

		toolCalls := completion.Choices[0].Message.ToolCalls

			if len(toolCalls) == 0 {
				// Attempt to validate JSON response; if invalid, log warning and return raw content
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(completion.Choices[0].Message.Content), &result); err != nil {
					log.Warn("Invalid JSON response, returning raw content", "error", err)
					return completion.Choices[0].Message.Content, nil
				}
				return completion.Choices[0].Message.Content, nil
			}

		// Add assistant message
		message := openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: param.NewOpt(completion.Choices[0].Message.Content),
				},
				ToolCalls: convertToolCalls(completion.Choices[0].Message.ToolCalls),
			},
		}
		params.Messages = append(params.Messages, message)

		// Process tool calls
		for _, toolCall := range toolCalls {
			if err := c.ProcessToolCall(toolCall, params); err != nil {
				return "", err
			}
		}
	}

	return "", fmt.Errorf("reached maximum iterations without resolution")
}

// ProcessToolCall handles a single tool call
func (c *Client) ProcessToolCall(toolCall openai.ChatCompletionMessageToolCall, params *openai.ChatCompletionNewParams) error {
   // Parse the arguments
   var args map[string]any
   if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
       log.Warn("Error parsing tool call arguments, skipping tool call", "error", err, "arguments", toolCall.Function.Arguments)
       // Skip this tool call and continue
       return nil
   }

	tool, exists := c.tools[toolCall.Function.Name]
	if !exists {
		log.Error("Unknown tool called", "tool", toolCall.Function.Name)
		return fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
	}

	// Execute the tool
	result, err := tool.Execute(args)
	if err != nil {
		log.Error("Error executing tool", "tool", toolCall.Function.Name, "error", err)
		return fmt.Errorf("error executing tool: %w", err)
	}

	// Add the tool call result to the conversation
	params.Messages = append(params.Messages, openai.ToolMessage(result, toolCall.ID))

	return nil
}

// schemaToFunctionParameters converts a schema to OpenAI function parameters
func schemaToFunctionParameters(schema any) openai.FunctionParameters {
	bytes, err := json.Marshal(schema)
	if err != nil {
		log.Error("Failed to marshal schema", "error", err)
		return openai.FunctionParameters{}
	}

	var params openai.FunctionParameters
	if err := json.Unmarshal(bytes, &params); err != nil {
		log.Error("Failed to unmarshal schema", "error", err)
		return openai.FunctionParameters{}
	}

	return params
}

// Helper function to convert tool calls
func convertToolCalls(calls []openai.ChatCompletionMessageToolCall) []openai.ChatCompletionMessageToolCallParam {
	params := make([]openai.ChatCompletionMessageToolCallParam, len(calls))
	for i, call := range calls {
		params[i] = openai.ChatCompletionMessageToolCallParam{
			Function: openai.ChatCompletionMessageToolCallFunctionParam{
				Name:      call.Function.Name,
				Arguments: call.Function.Arguments,
			},
			ID: call.ID,
		}
	}
	return params
}

// ToolExecutor is the interface for all tool executors
type ToolExecutor interface {
	Execute(args map[string]any) (string, error)
}

func (c *Client) Execute(
	buffer *Buffer,
	maxIterations int,
) (string, error) {
	log.Info("Starting OpenAI client execution", "maxIterations", maxIterations)

	schema := utils.GenerateSchema[models.APIConfig]()

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "api_config",
		Description: param.NewOpt("The API configuration"),
		Schema:      param.NewOpt(schema),
		Strict:      param.NewOpt(false),
	}

	availableTools := []models.ToolType{
		models.NewTool(browser.NewBrowserExtractor()),
		models.NewTool(browser.NewBrowserNavigator()),
		models.NewTool(browser.NewBrowserJavaScriptExecutor()),
		models.NewTool(browser.NewBrowserClicker()),
		models.NewTool(mongodb.NewMongoDBInspector()),
	}

	tools := []openai.ChatCompletionToolParam{}

	for _, tool := range availableTools {
		tools = append(tools, openai.ChatCompletionToolParam{
			Type: "function",
			Function: openai.FunctionDefinitionParam{
				Name:        tool.Name(),
				Description: param.NewOpt(tool.Description()),
				Parameters:  schemaToFunctionParameters(tool.Schema()),
			},
		})
	}

	params := openai.ChatCompletionNewParams{
		Model:       openai.ChatModelGPT4o,
		Messages:    buffer.Truncate().Messages,
		Tools:       tools,
		Temperature: param.NewOpt(0.0),
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				Type:       "json_schema",
				JSONSchema: schemaParam,
			},
		},
	}

	// Iterate until the model stops requesting tool calls
	for i := range maxIterations {
		log.Info("Executing iteration", "iteration", i+1, "of", maxIterations)

		log.Info("Messages", "messages", len(params.Messages))
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

		// Convert the message to the correct type before appending
		message := openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: param.NewOpt(completion.Choices[0].Message.Content),
				},
				ToolCalls: convertToolCalls(completion.Choices[0].Message.ToolCalls),
			},
		}
		params.Messages = append(params.Messages, message)

		for _, toolCall := range toolCalls {
			if err := c.ProcessToolCall(toolCall, &params); err != nil {
				return "", err
			}
		}
	}

	log.Error("Maximum iterations reached without resolution", "maxIterations", maxIterations)
	return "", fmt.Errorf("reached maximum iterations without resolution")
}
