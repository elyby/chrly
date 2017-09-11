package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"elyby/minecraft-skinsystem/bootstrap"
	"runtime"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the Minecraft Skinsystem version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version:    %s\n",    bootstrap.GetVersion())
		fmt.Printf("Go version: %s\n",    runtime.Version())
		fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
