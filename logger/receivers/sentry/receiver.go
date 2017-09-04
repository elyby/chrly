package sentry

import (
	"fmt"

	"github.com/getsentry/raven-go"
	"github.com/mono83/slf"
	"github.com/mono83/slf/filters"
)

// Config holds information for filtered receiver
type Config struct {
	MinLevel        string
	ParamsWhiteList []string
	ParamsBlackList []string
}

// NewReceiver allows you to create a new receiver in the Sentry
// using the fastest and easiest way.
// The Config parameter can be passed as nil if you do not need additional filtration.
func NewReceiver(dsn string, cfg *Config) (slf.Receiver, error) {
	client, err := raven.New(dsn)
	if err != nil {
		return nil, err
	}

	return NewReceiverWithCustomRaven(client, cfg)
}

// NewReceiverWithCustomRaven allows you to create a new receiver in the Sentry
// configuring raven.Client by yourself. This can be useful if you need to set
// additional parameters, such as release and environment, that will be sent
// with each Packet in the Sentry:
//
// client, err := raven.New("https://some:sentry@dsn.sentry.io/1")
// if err != nil {
//     return nil, err
// }
//
// client.SetRelease("1.3.2")
// client.SetEnvironment("production")
// client.SetDefaultLoggerName("sentry-watchdog-receiver")
//
// sentryReceiver, err := sentry.NewReceiverWithCustomRaven(client, &sentry.Config{
//     MinLevel: "warn",
// })
//
// The Config parameter allows you to add additional filtering, such as the minimum
// message level and the exclusion of private parameters. If you do not need additional
// filtering, nil can passed.
func NewReceiverWithCustomRaven(client *raven.Client, cfg *Config) (slf.Receiver, error) {
	out, err := buildReceiverForClient(client)
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		return out, nil
	}

	// Resolving level
	level, ok := slf.ParseType(cfg.MinLevel)
	if !ok {
		return nil, fmt.Errorf("Unknown level %s", cfg.MinLevel)
	}

	if len(cfg.ParamsWhiteList) > 0 {
		out.filter = slf.NewWhiteListParamsFilter(cfg.ParamsWhiteList)
	} else {
		out.filter = slf.NewBlackListParamsFilter(cfg.ParamsBlackList)
	}

	return filters.MinLogLevel(level, out), nil
}

func buildReceiverForClient(client *raven.Client) (*sentryLogReceiver, error) {
	return &sentryLogReceiver{target: client, filter: slf.NewBlackListParamsFilter(nil)}, nil
}

type sentryLogReceiver struct {
	target *raven.Client
	filter slf.ParamsFilter
}

func (l sentryLogReceiver) Receive(p slf.Event) {
	if !p.IsLog() {
		return
	}

	pkt := raven.NewPacket(
		slf.ReplacePlaceholders(p.Content, p.Params, false),
		// First 5 means, that first N elements will be skipped before actual app trace
		// This is needed to exclude watchdog calls from stack trace
		raven.NewStacktrace(5, 5, []string{}),
	)

	if len(p.Params) > 0 {
		shownParams := l.filter(p.Params)
		for _, param := range shownParams {
			value := param.GetRaw()
			if e, ok := value.(error); ok && e != nil {
				value = e.Error()
			}

			pkt.Extra[param.GetKey()] = value
		}
	}

	pkt.Level = convertType(p.Type)
	pkt.Timestamp = raven.Timestamp(p.Time)

	l.target.Capture(pkt, map[string]string{})
}

func convertType(wdType byte) raven.Severity {
	switch wdType {
	case slf.TypeTrace:
	case slf.TypeDebug:
		return raven.DEBUG
	case slf.TypeInfo:
		return raven.INFO
	case slf.TypeWarning:
		return raven.WARNING
	case slf.TypeError:
		return raven.ERROR
	case slf.TypeAlert:
	case slf.TypeEmergency:
		return raven.FATAL
	}

	panic("Unknown wd type " + string(wdType))
}
