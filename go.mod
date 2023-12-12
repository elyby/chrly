module github.com/elyby/chrly

go 1.21

replace github.com/asaskevich/EventBus v0.0.0-20200330115301-33b3bc6a7ddc => github.com/erickskrauch/EventBus v0.0.0-20200330115301-33b3bc6a7ddc

// Main dependencies
require (
	github.com/SermoDigital/jose v0.9.2-0.20161205224733-f6df55f235c2
	github.com/asaskevich/EventBus v0.0.0-20200330115301-33b3bc6a7ddc
	github.com/etherlabsio/healthcheck v0.0.0-20191224061800-dd3d2fd8c3f6
	github.com/getsentry/raven-go v0.2.1-0.20190419175539-919484f041ea
	github.com/goava/di v1.1.1-0.20200420103225-1eb6eb721bf0
	github.com/gorilla/mux v1.7.1
	github.com/mediocregopher/radix.v2 v0.0.0-20181115013041-b67df6e626f9
	github.com/mono83/slf v0.0.0-20170919161409-79153e9636db
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.3.2
	github.com/thedevsaddam/govalidator v1.9.6
)

// Dev dependencies
require (
	github.com/stretchr/testify v1.8.4
	gopkg.in/h2non/gock.v1 v1.0.15
)

require (
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/certifi/gocertifi v0.0.0-20210507211836-431795d63e8d // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.4.7 // indirect
	github.com/h2non/parth v0.0.0-20190131123155-b4df798d6542 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/levenlabs/golib v0.0.0-20180911183212-0f8974794783 // indirect
	github.com/magiconair/properties v1.8.0 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/mono83/udpwriter v1.0.2 // indirect
	github.com/oschwald/geoip2-golang v1.9.0 // indirect
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/afero v1.1.2 // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	golang.org/x/sys v0.9.0 // indirect
	golang.org/x/text v0.3.8 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
