module github.com/elyby/chrly

go 1.21

replace github.com/asaskevich/EventBus v0.0.0-20200330115301-33b3bc6a7ddc => github.com/erickskrauch/EventBus v0.0.0-20200330115301-33b3bc6a7ddc

// Main dependencies
require (
	github.com/SermoDigital/jose v0.9.2-0.20161205224733-f6df55f235c2
	github.com/asaskevich/EventBus v0.0.0-20200330115301-33b3bc6a7ddc
	github.com/defval/di v1.12.0
	github.com/etherlabsio/healthcheck/v2 v2.0.0
	github.com/getsentry/raven-go v0.2.1-0.20190419175539-919484f041ea
	github.com/gorilla/mux v1.8.1
	github.com/mediocregopher/radix/v4 v4.1.4
	github.com/mono83/slf v0.0.0-20170919161409-79153e9636db
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.18.1
	github.com/thedevsaddam/govalidator v1.9.10
)

// Dev dependencies
require (
	github.com/h2non/gock v1.2.0
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/certifi/gocertifi v0.0.0-20210507211836-431795d63e8d // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/h2non/parth v0.0.0-20190131123155-b4df798d6542 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mono83/udpwriter v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tilinna/clock v1.0.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20231206192017-f3f8817b8deb // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
