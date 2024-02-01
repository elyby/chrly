package cmd

import (
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts HTTP handler for the skins system",
	Run: func(cmd *cobra.Command, args []string) {
		startServer([]string{"skinsystem", "api"})
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
