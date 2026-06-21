package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{Use: "couchdev", Short: "Claude Code RC session launcher"}

	var configPath string
	root.PersistentFlags().StringVarP(&configPath, "config", "c", "/etc/couchdev/config.json", "config file")

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("serve: not yet implemented")
			return nil
		},
	}

	tokenCmd := &cobra.Command{Use: "token", Short: "Token management"}
	tokenGenCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a bearer token and print its SHA-256 hash",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("token generate: not yet implemented")
			return nil
		},
	}
	tokenCmd.AddCommand(tokenGenCmd)
	root.AddCommand(serveCmd, tokenCmd)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
