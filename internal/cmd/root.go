package cmd

import (
	"strings"

	. "github.com/defval/di"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"ely.by/chrly/internal/di"
	"ely.by/chrly/internal/version"
)

var RootCmd = &cobra.Command{
	Use:     "chrly",
	Short:   "Implementation of the Minecraft skins system server",
	Version: version.Version(),
}

func shouldGetContainer() *Container {
	container, err := di.New()
	if err != nil {
		panic(err)
	}

	return container
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
}
