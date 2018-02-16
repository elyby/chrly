package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/elyby/chrly/bootstrap"
	"runtime"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the Chrly version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version:    %s\n",    bootstrap.GetVersion())
		fmt.Printf("Go version: %s\n",    runtime.Version())
		fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
