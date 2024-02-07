package cmd

import (
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts HTTP handler for the skins system",
	RunE: func(cmd *cobra.Command, args []string) error {
		return startServer("skinsystem", "api")
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
