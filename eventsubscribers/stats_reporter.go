package eventsubscribers

import (
	"net/http"
	"strings"

	"github.com/mono83/slf"

	"github.com/elyby/chrly/dispatcher"
)

type StatsReporter struct {
	Reporter slf.StatsReporter
	Prefix   string
}

func (s *StatsReporter) ConfigureWithDispatcher(d dispatcher.EventDispatcher) {
	d.Subscribe("skinsystem:before_request", s.handleBeforeRequest)
	d.Subscribe("skinsystem:after_request", s.handleAfterRequest)

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
	} else if m == http.MethodPost && p == "/api/skins" {
		key = "api.skins.post.request"
	} else if m == http.MethodDelete && strings.HasPrefix(p, "/api/skins/") {
		key = "api.skins.delete.request"
	} else {
		return
	}

	s.incCounter(key)
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

	s.incCounter(key)
}

func (s *StatsReporter) incCounterHandler(name string) func(...interface{}) {
	return func(...interface{}) {
		s.incCounter(name)
	}
}

func (s *StatsReporter) incCounter(name string) {
	s.Reporter.IncCounter(s.key(name), 1)
}

func (s *StatsReporter) key(name string) string {
	return strings.Join([]string{s.Prefix, name}, ".")
}
