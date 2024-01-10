package eventsubscribers

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mono83/slf"
)

type StatsReporter struct {
	slf.StatsReporter
	Prefix string

	timersMap   map[string]time.Time
	timersMutex sync.Mutex
}

type Reporter interface {
	Enable(reporter slf.StatsReporter)
}

type ReporterFunc func(reporter slf.StatsReporter)

func (f ReporterFunc) Enable(reporter slf.StatsReporter) {
	f(reporter)
}

// TODO: rework all reporters in the same style as it was there: https://github.com/elyby/chrly/blob/1543e98b/di/db.go#L48-L52
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
	} else if strings.HasPrefix(p, "/profile/") {
		key = "profiles.request"
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
