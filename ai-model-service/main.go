package main

import (
	"fmt"
	"os"

	"github.com/blcvn/backend/services/ai-model-service/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
