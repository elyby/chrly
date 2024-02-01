package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"ely.by/chrly/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the Chrly version information",
	Run: func(cmd *cobra.Command, args []string) {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "<unknown>"
		}

		fmt.Printf("Version:    %s\n", version.Version())
		fmt.Printf("Commit:     %s\n", version.Commit())
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Hostname:   %s\n", hostname)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
