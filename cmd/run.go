package cmd

import (
	"encoding/json"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/theapemachine/idrinkyourmilkshake/engine"
	"github.com/theapemachine/idrinkyourmilkshake/models"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run an integration using a configuration file",
	Long:  `Execute the integration engine with the provided API configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatalf("Failed to read config flag: %v", err)
		}
		if configFile == "" {
			configFile = "integration-config.json"
		}

		data, err := os.ReadFile(configFile)
		if err != nil {
			log.Fatalf("Failed to read config file %s: %v", configFile, err)
		}

		var cfg models.APIConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("Failed to parse config file: %v", err)
		}

		eng := engine.NewIntegration(cfg)
		if err := eng.Execute(); err != nil {
			log.Fatalf("Integration failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("config", "c", "integration-config.json", "Path to the integration configuration file")
}
