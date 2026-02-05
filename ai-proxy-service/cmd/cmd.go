package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "ai-proxy-service",
	Short: "AI Proxy Service",
	Long:  `AI Proxy Service for centralized LLM access, logging and quota management`,
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
