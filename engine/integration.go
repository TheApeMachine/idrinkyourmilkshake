package engine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/idrinkyourmilkshake/models"
	"github.com/theapemachine/idrinkyourmilkshake/mongodb"
	"github.com/theapemachine/idrinkyourmilkshake/request"
)

/*
Integration is the execution engine for a given API config.
It will execute the API calls in the order specified by the API config.
It will also handle the pagination of the API calls if the API config specifies it.
*/
type Integration struct {
	config  models.APIConfig
	results map[string]map[string]interface{} // Store results from each step: stepID -> outputName -> value
}

func NewIntegration(config models.APIConfig) *Integration {
	return &Integration{
		config:  config,
		results: make(map[string]map[string]interface{}),
	}
}

func (integration *Integration) Execute() (err error) {
	log.Info("Starting integration execution", "integration", integration.config.Integration)

	for _, job := range integration.config.Jobs {
		log.Info("Executing job", "jobID", job.ID)

		for _, step := range job.Steps {
			if err = integration.executeStep(step); err != nil {
				log.Error("Error executing step", "stepID", step.ID, "error", err)
				return err
			}
		}
	}

	log.Info("Integration execution completed successfully", "integration", integration.config.Integration)
	return nil
}

// processVariables replaces {{variable}} patterns with actual values from previous results
func (integration *Integration) processVariables(input string) (string, error) {
	// Pattern to match {{step.outputs.key}} patterns
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	result := re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable path inside {{ }}
		varPath := strings.Trim(match, "{}")
		parts := strings.Split(varPath, ".")

		if len(parts) < 3 {
			log.Warn("Invalid variable format", "variable", match)
			return match // Return original if format is invalid
		}

		stepID := parts[0]
		if stepResults, ok := integration.results[stepID]; ok {
			if len(parts) == 3 && parts[1] == "outputs" {
				outputKey := parts[2]
				if value, ok := stepResults[outputKey]; ok {
					// Convert value to string
					valueStr, err := json.Marshal(value)
					if err != nil {
						log.Error("Failed to marshal value", "error", err)
						return match
					}
					return strings.Trim(string(valueStr), "\"")
				}
			}
		}

		log.Warn("Variable not found", "variable", match)
		return match // Return original if not found
	})

	return result, nil
}

// processMapVariables processes variables in a map's keys and values
func (integration *Integration) processMapVariables(input map[string]string) (map[string]string, error) {
	result := make(map[string]string)

	for key, value := range input {
		processedKey, err := integration.processVariables(key)
		if err != nil {
			return nil, err
		}

		processedValue, err := integration.processVariables(value)
		if err != nil {
			return nil, err
		}

		result[processedKey] = processedValue
	}

	return result, nil
}

func (integration *Integration) executeStep(step models.Step) (err error) {
	log.Info("Executing step", "id", step.ID, "type", step.Type)

	// Initialize result storage for this step
	integration.results[step.ID] = make(map[string]interface{})

	switch step.Type {
	case "http":
		// Process endpoint with variables if needed
		endpoint, err := integration.processVariables(step.Endpoint)
		if err != nil {
			return fmt.Errorf("failed to process endpoint variables: %w", err)
		}

		// Create HTTP request
		httpTool := request.NewHTTPRequest()

		// Prepare arguments with authentication if needed
		args := map[string]any{
			"url":    integration.config.BaseURL + endpoint,
			"method": step.Method,
		}

		// Add authentication details from config if available
		if integration.config.Auth.Type != "" {
			headers := make(map[string]string)
			for _, output := range integration.config.Auth.Outputs {
				headers[output.Key] = output.Value
			}

			if len(headers) > 0 {
				// Use processMapVariables instead of processing individually
				processedHeaders, err := integration.processMapVariables(headers)
				if err != nil {
					return fmt.Errorf("failed to process auth headers: %w", err)
				}
				args["headers"] = processedHeaders
			}
		}

		// Execute the HTTP request
		result, err := httpTool.Execute(args)
		if err != nil {
			return fmt.Errorf("failed to execute HTTP step %s: %w", step.ID, err)
		}

		// Process and store outputs
		var responseData map[string]interface{}
		if err := json.Unmarshal([]byte(result), &responseData); err != nil {
			log.Warn("Failed to parse response as JSON", "error", err)
			// Store raw response if not JSON
			integration.results[step.ID]["response"] = result
		} else {
			// Store parsed response
			integration.results[step.ID]["response"] = responseData

			// Store specific outputs as defined in step config
			for _, output := range step.Outputs {
				// Extract value using output.Key as path
				// For now, we'll just store the whole response
				integration.results[step.ID][output.Key] = responseData
			}
		}

		log.Info("HTTP step executed successfully", "id", step.ID)

	case "mongodb":
		// Create MongoDB tool
		mongoTool := mongodb.NewMongoDBInspector()

		// Process collection with variables if needed
		collection, err := integration.processVariables(step.Collection)
		if err != nil {
			return fmt.Errorf("failed to process collection variables: %w", err)
		}

		// Prepare arguments
		args := map[string]any{
			"collection": collection,
			"operation":  step.Operation,
		}

		// Process documents if available
		if len(step.Documents) > 0 {
			documents := make([]map[string]string, 0)
			for _, doc := range step.Documents {
				processedKey, err := integration.processVariables(doc.Key)
				if err != nil {
					return fmt.Errorf("failed to process document key variable: %w", err)
				}

				processedValue, err := integration.processVariables(doc.Value)
				if err != nil {
					return fmt.Errorf("failed to process document value variable: %w", err)
				}

				documents = append(documents, map[string]string{
					processedKey: processedValue,
				})
			}
			args["documents"] = documents
		}

		// Execute the MongoDB operation
		result, err := mongoTool.Execute(args)
		if err != nil {
			return fmt.Errorf("failed to execute MongoDB step %s: %w", step.ID, err)
		}

		// Process and store outputs
		var resultData map[string]interface{}
		if err := json.Unmarshal([]byte(result), &resultData); err != nil {
			log.Warn("Failed to parse MongoDB result as JSON", "error", err)
			// Store raw result if not JSON
			integration.results[step.ID]["result"] = result
		} else {
			// Store parsed result
			integration.results[step.ID]["result"] = resultData
		}

		log.Info("MongoDB step executed successfully", "id", step.ID)

	default:
		return fmt.Errorf("unsupported step type: %s", step.Type)
	}

	return nil
}
