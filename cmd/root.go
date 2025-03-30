package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "idrinkyourmilkshake",
	Short: "idrinkyourmilkshake",
	Long:  txtRootLong,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

const txtRootLong = `
idrinkyourmilkshake is an AI-powered CLI tool that helps you integrate with APIs.
It uses the power of LLMs to understand API documentation, and maps it to your data model.
It produces a configuration object which is used by the engine to execute the integration.
`
