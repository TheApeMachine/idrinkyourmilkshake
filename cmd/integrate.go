package cmd

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/theapemachine/idrinkyourmilkshake/ai"
	"github.com/theapemachine/idrinkyourmilkshake/models"
)

var integrateCmd = &cobra.Command{
	Use:   "integrate",
	Short: "Generate API integration configuration",
	Long:  txtIntegrateLong,
	Run: func(cmd *cobra.Command, args []string) {
		docsURL, err := cmd.Flags().GetString("docs")
		if err != nil {
			log.Fatalf("Failed to get docs URL: %v", err)
		}
		if docsURL == "" {
			log.Fatal("docs URL is required")
		}

		outputFile, err := cmd.Flags().GetString("output")
		if err != nil {
			log.Fatalf("Failed to get output file: %v", err)
		}
		if outputFile == "" {
			outputFile = "integration-config.json"
		}

		// Initialize API config with docs URL
		cfg := &models.APIConfig{
			BaseURL: docsURL,
		}

		// Create AI pipeline to analyze docs and generate config
		systemPrompt := `You are an API integration expert. Your task is to create a configuration that maps data from the API at ` + docsURL + ` into our existing MongoDB collections.

First, analyze the API documentation to understand:
1. What data endpoints are available
2. What data structures they return
3. Authentication requirements

Then, for each API endpoint's data structure:
1. Use mongodb_inspector to find collections with matching schemas
2. Only include endpoints where you find a suitable collection match
3. Do not make up non-existing collections, only use those that are actually present in the MongoDB database (you can list all collections with the mongodb_inspector tool, by not providing a collection name)

Do not try to map every collection - focus only on finding the right collections for the API data you discovered.

While you are working, you are placed in an iteration loop, so you should use this to execute multiple tool calls, and fully explore the API documentation, and the MongoDB collections.
`

		// Execute AI pipeline to generate config
		cfg, err = ai.NewPipeline(systemPrompt).Execute(context.Background(), cfg)
		if err != nil {
			log.Fatalf("Failed to generate integration config: %v", err)
		}

		// Save the configuration to a file
		configJSON, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal config to JSON: %v", err)
		}

		if err := os.WriteFile(outputFile, configJSON, 0644); err != nil {
			log.Fatalf("Failed to write config file: %v", err)
		}

		log.Printf("Integration configuration generated and saved to %s", outputFile)
		log.Printf("You can now use this configuration file in your Kubernetes cron job")
	},
}

func init() {
	rootCmd.AddCommand(integrateCmd)
	integrateCmd.Flags().StringP("docs", "d", "", "URL to the API documentation")
	integrateCmd.Flags().StringP("output", "o", "", "Output file for the configuration (default: integration-config.json)")
}

const txtIntegrateLong = `
Generate an API integration configuration by analyzing API documentation.
This tool uses AI to understand the API and create a configuration file that maps API data to MongoDB collections.
The generated configuration can then be used by the engine running in Kubernetes to pull and load data.
`
