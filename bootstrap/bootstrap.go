package bootstrap

import (
	"os"

	"github.com/mono83/slf/rays"
	"github.com/mono83/slf/recievers/ansi"
	"github.com/mono83/slf/recievers/statsd"
	"github.com/mono83/slf/wd"
)

func CreateLogger(statsdAddr string) (wd.Watchdog, error) {
	wd.AddReceiver(ansi.New(true, true, false))
	if statsdAddr != "" {
		hostname, _ := os.Hostname()
		statsdReceiver, err := statsd.NewReceiver(statsd.Config{
			Address: statsdAddr,
			Prefix: "ely.skinsystem." + hostname + ".app.",
			FlushEvery: 1,
		})

		if err != nil {
			return nil, err
		}

		wd.AddReceiver(statsdReceiver)
	}

	return wd.New("", "").WithParams(rays.Host), nil
}
