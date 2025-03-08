package main

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/idrinkyourmilkshake/openai"
)

func main() {
	log.Info("Starting application")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	log.Info("Initializing OpenAI client")
	client := openai.NewClient(apiKey)

	log.Info("Creating background context")
	ctx := context.Background()
	client = client.WithContext(ctx)

	log.Info("Creating conversation buffer with system and user prompts")
	buffer := openai.NewBuffer(
		`
		You are an advanced API integration expert.
		You work with a specialized API Integration Engine that relies on a configuration file to drive all parts of the integration.
		You will be given a URL to a page of API documentation and your job is to extract the API endpoints and data models from the documentation and generate a configuration object.
		You have access to a full Chrome browser as a tool, so you can navigate the documentation and do whatever is needed to extract the information.
		You also have access to an HTTP request tool, so you can interact with APIs when needed.
		`,
		`
		Here is the documentation URL for the API: https://developer.dyflexis.com/v3
		`,
	)

	log.Info("Starting OpenAI client execution with max iterations", "maxIterations", 20)
	result, err := client.Execute(buffer, 20)
	if err != nil {
		log.Fatal("Error executing OpenAI client", "error", err)
	}

	log.Info("Execution completed successfully", "resultLength", len(result))
}
