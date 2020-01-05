package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/elyby/chrly/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the Chrly version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version:    %s\n", version.Version())
		fmt.Printf("Commit:     %s\n", version.Commit())
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
