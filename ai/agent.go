package ai

import (
	"context"
	"os"

	"github.com/openai/openai-go"
	"github.com/theapemachine/idrinkyourmilkshake/browser"
	"github.com/theapemachine/idrinkyourmilkshake/models"
	"github.com/theapemachine/idrinkyourmilkshake/mongodb"
	oai "github.com/theapemachine/idrinkyourmilkshake/openai"
)

type Agent struct {
	SystemPrompt string
	UserPrompt   string
	client       *oai.Client
	params       *openai.ChatCompletionNewParams
}

func NewAgent(systemPrompt, userPrompt string) *Agent {
	// Initialize OpenAI client
	client := oai.NewClient(os.Getenv("OPENAI_API_KEY"))
	client.WithContext(context.Background())

	// Initialize tools
	availableTools := []models.ToolType{
		browser.NewBrowserExtractor(),
		browser.NewBrowserNavigator(),
		browser.NewBrowserJavaScriptExecutor(),
		browser.NewBrowserClicker(),
		mongodb.NewMongoDBInspector(),
	}

	// Register tools with the client
	client.RegisterTools(availableTools)

	// Create completion parameters
	params := client.CreateCompletionParams(systemPrompt, userPrompt)

	return &Agent{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		client:       client,
		params:       params,
	}
}

func (a *Agent) Execute() (string, error) {
	maxIterations := 50
	return a.client.ExecuteWithTools(a.params, maxIterations)
}
