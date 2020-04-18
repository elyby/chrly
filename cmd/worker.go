package cmd

import (
	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Starts HTTP handler for the Mojang usernames to UUIDs worker",
	Run: func(cmd *cobra.Command, args []string) {
		startServer([]string{"worker"})
	},
}

func init() {
	RootCmd.AddCommand(workerCmd)
}
