package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "create-go-api",
	Short: "A CLI tool to scaffold production-ready Go API services",
	Long:  `create-go-api is a CLI tool that generates Go API service boilerplate with database support (DynamoDB, Postgres), framework support (Chi, ConnectRPC), and one-click deployment.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(createCmd)
}

