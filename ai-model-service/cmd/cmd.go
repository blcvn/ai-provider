package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd represents the base command
var RootCmd = &cobra.Command{
	Use:   "ai-model-service",
	Short: "AI Model Service",
	Long:  `AI Model Service for managing AI model configurations and credentials`,
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
