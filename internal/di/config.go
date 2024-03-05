package di

import (
	"github.com/defval/di"
	"github.com/spf13/viper"
)

var configDiOptions = di.Options(
	di.Provide(viper.GetViper),
)
