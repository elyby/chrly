package cmd

import (
	"github.com/spf13/cobra"

	"ely.by/chrly/internal/di"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts HTTP handler for the skins system",
	RunE: func(cmd *cobra.Command, args []string) error {
		return startServer(di.ModuleSkinsystem, di.ModuleProfiles, di.ModuleSigner)
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
