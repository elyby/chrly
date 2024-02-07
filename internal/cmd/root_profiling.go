//go:build profiling

package cmd

import (
	"log"
	"os"
	"runtime/pprof"

	"github.com/spf13/cobra"
)

func init() {
	var profilePath string
	RootCmd.PersistentFlags().StringVar(&profilePath, "cpuprofile", "", "enables pprof profiling and sets its output path")

	pprofEnabled := false
	originalPersistentPreRunE := RootCmd.PersistentPreRunE
	RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if profilePath == "" {
			return nil
		}

		f, err := os.Create(profilePath)
		if err != nil {
			return err
		}

		log.Println("enabling profiling")
		err = pprof.StartCPUProfile(f)
		if err != nil {
			return err
		}

		pprofEnabled = true

		if originalPersistentPreRunE != nil {
			return originalPersistentPreRunE(cmd, args)
		}

		return nil
	}

	originalPersistentPostRun := RootCmd.PersistentPreRun
	RootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if pprofEnabled {
			log.Println("shutting down profiling")
			pprof.StopCPUProfile()
		}

		if originalPersistentPostRun != nil {
			originalPersistentPostRun(cmd, args)
		}
	}
}
