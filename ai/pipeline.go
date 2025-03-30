package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theapemachine/idrinkyourmilkshake/models"
)

// Pipeline processes the API documentation and builds the configuration
type Pipeline struct {
	agent *Agent
}

func NewPipeline(systemPrompt string) *Pipeline {
	return &Pipeline{
		agent: NewAgent(systemPrompt, ""),
	}
}

func (p *Pipeline) Execute(ctx context.Context, apiConfig *models.APIConfig) (*models.APIConfig, error) {
	// Set the user prompt with the API docs URL
	p.agent.UserPrompt = "Please analyze the API documentation and extract endpoints, authentication, and job configurations."

	// Execute the agent to analyze the documentation
	result, err := p.agent.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to execute agent: %w", err)
	}

	// Parse the agent's response into our APIConfig structure
	if err := json.Unmarshal([]byte(result), apiConfig); err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
	}

	return apiConfig, nil
}
