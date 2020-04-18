package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	. "github.com/goava/di"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/elyby/chrly/di"
	"github.com/elyby/chrly/http"
	"github.com/elyby/chrly/version"
)

var RootCmd = &cobra.Command{
	Use:     "chrly",
	Short:   "Implementation of Minecraft skins system server",
	Version: version.Version(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func shouldGetContainer() *Container {
	container, err := di.New()
	if err != nil {
		panic(err)
	}

	return container
}

func startServer(modules []string) {
	container := shouldGetContainer()

	var config *viper.Viper
	err := container.Resolve(&config)
	if err != nil {
		log.Fatal(err)
	}

	config.Set("modules", modules)

	err = container.Invoke(http.StartServer)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
}
