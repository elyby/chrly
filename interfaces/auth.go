package interfaces

import "net/http"

type AuthChecker interface {
	Check(req *http.Request) error
}
