package eventsubscribers

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mono83/slf"

	"github.com/elyby/chrly/api/mojang"
)

type StatsReporter struct {
	slf.StatsReporter
	Prefix string

	timersMap   map[string]time.Time
	timersMutex sync.Mutex
}

func (s *StatsReporter) ConfigureWithDispatcher(d Subscriber) {
	s.timersMap = make(map[string]time.Time)

	// Per request events
	d.Subscribe("skinsystem:before_request", s.handleBeforeRequest)
	d.Subscribe("skinsystem:after_request", s.handleAfterRequest)

	// Authentication events
	d.Subscribe("authenticator:success", s.incCounterHandler("authentication.challenge")) // TODO: legacy, remove in v5
	d.Subscribe("authenticator:success", s.incCounterHandler("authentication.success"))
	d.Subscribe("authentication:error", s.incCounterHandler("authentication.challenge")) // TODO: legacy, remove in v5
	d.Subscribe("authentication:error", s.incCounterHandler("authentication.failed"))

	// Mojang signed textures source events
	d.Subscribe("mojang_textures:call", s.incCounterHandler("mojang_textures.request"))
	d.Subscribe("mojang_textures:usernames:after_cache", func(username string, uuid string, err error) {
		if err != nil {
			return
		}

		if uuid == "" {
			s.IncCounter("mojang_textures:usernames:cache_hit_nil", 1)
		} else {
			s.IncCounter("mojang_textures:usernames:cache_hit", 1)
		}
	})
	d.Subscribe("mojang_textures:textures:after_cache", func(uuid string, textures *mojang.SignedTexturesResponse, err error) {
		if err != nil {
			return
		}

		if textures != nil {
			s.IncCounter("mojang_textures.textures.cache_hit", 1)
		}
	})
	d.Subscribe("mojang_textures:already_processing", s.incCounterHandler("mojang_textures.already_scheduled"))
	d.Subscribe("mojang_textures:usernames:after_call", func(username string, profile *mojang.ProfileInfo, err error) {
		if err != nil {
			return
		}

		if profile == nil {
			s.IncCounter("mojang_textures.usernames.uuid_miss", 1)
		} else {
			s.IncCounter("mojang_textures.usernames.uuid_hit", 1)
		}
	})
	d.Subscribe("mojang_textures:textures:before_call", s.incCounterHandler("mojang_textures.textures.request"))
	d.Subscribe("mojang_textures:textures:after_call", func(uuid string, textures *mojang.SignedTexturesResponse, err error) {
		if err != nil {
			return
		}

		if textures == nil {
			s.IncCounter("mojang_textures.usernames.textures_miss", 1)
		} else {
			s.IncCounter("mojang_textures.usernames.textures_hit", 1)
		}
	})
	d.Subscribe("mojang_textures:before_result", func(username string, uuid string) {
		s.startTimeRecording("mojang_textures_result_time_" + username)
	})
	d.Subscribe("mojang_textures:after_result", func(username string, textures *mojang.SignedTexturesResponse, err error) {
		s.finalizeTimeRecording("mojang_textures_result_time_"+username, "mojang_textures.result_time")
	})
	d.Subscribe("mojang_textures:textures:before_call", func(uuid string) {
		s.startTimeRecording("mojang_textures_provider_time_" + uuid)
	})
	d.Subscribe("mojang_textures:textures:after_call", func(uuid string, textures *mojang.SignedTexturesResponse, err error) {
		s.finalizeTimeRecording("mojang_textures_provider_time_"+uuid, "mojang_textures.textures.request_time")
	})

	// Mojang UUIDs batch provider metrics
	d.Subscribe("mojang_textures:batch_uuids_provider:queued", s.incCounterHandler("mojang_textures.usernames.queued"))
	d.Subscribe("mojang_textures:batch_uuids_provider:round", func(usernames []string, queueSize int) {
		s.UpdateGauge("mojang_textures.usernames.iteration_size", int64(len(usernames)))
		s.UpdateGauge("mojang_textures.usernames.queue_size", int64(queueSize))
		if len(usernames) != 0 {
			s.startTimeRecording("batch_uuids_provider_round_time_" + strings.Join(usernames, "|"))
		}
	})
	d.Subscribe("mojang_textures:batch_uuids_provider:result", func(usernames []string, profiles []*mojang.ProfileInfo, err error) {
		s.finalizeTimeRecording("batch_uuids_provider_round_time_"+strings.Join(usernames, "|"), "mojang_textures.usernames.round_time")
	})
}

func (s *StatsReporter) handleBeforeRequest(req *http.Request) {
	var key string
	m := req.Method
	p := req.URL.Path
	if p == "/skins" {
		key = "skins.get_request"
	} else if strings.HasPrefix(p, "/skins/") {
		key = "skins.request"
	} else if p == "/cloaks" {
		key = "capes.get_request"
	} else if strings.HasPrefix(p, "/cloaks/") {
		key = "capes.request"
	} else if strings.HasPrefix(p, "/textures/signed/") {
		key = "signed_textures.request"
	} else if strings.HasPrefix(p, "/textures/") {
		key = "textures.request"
	} else if m == http.MethodPost && p == "/api/skins" {
		key = "api.skins.post.request"
	} else if m == http.MethodDelete && strings.HasPrefix(p, "/api/skins/") {
		key = "api.skins.delete.request"
	} else {
		return
	}

	s.IncCounter(key, 1)
}

func (s *StatsReporter) handleAfterRequest(req *http.Request, code int) {
	var key string
	m := req.Method
	p := req.URL.Path
	if m == http.MethodPost && p == "/api/skins" && code == http.StatusCreated {
		key = "api.skins.post.success"
	} else if m == http.MethodPost && p == "/api/skins" && code == http.StatusBadRequest {
		key = "api.skins.post.validation_failed"
	} else if m == http.MethodDelete && strings.HasPrefix(p, "/api/skins/") && code == http.StatusNoContent {
		key = "api.skins.delete.success"
	} else if m == http.MethodDelete && strings.HasPrefix(p, "/api/skins/") && code == http.StatusNotFound {
		key = "api.skins.delete.not_found"
	} else {
		return
	}

	s.IncCounter(key, 1)
}

func (s *StatsReporter) incCounterHandler(name string) func(...interface{}) {
	return func(...interface{}) {
		s.IncCounter(name, 1)
	}
}

func (s *StatsReporter) startTimeRecording(timeKey string) {
	s.timersMutex.Lock()
	defer s.timersMutex.Unlock()
	s.timersMap[timeKey] = time.Now()
}

func (s *StatsReporter) finalizeTimeRecording(timeKey string, statName string) {
	s.timersMutex.Lock()
	defer s.timersMutex.Unlock()
	startedAt, ok := s.timersMap[timeKey]
	if !ok {
		return
	}

	delete(s.timersMap, timeKey)

	s.RecordTimer(statName, time.Since(startedAt))
}
