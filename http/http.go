package http

import (
	"encoding/json"
	"net/http"
)

func NotFound(response http.ResponseWriter, _ *http.Request) {
	data, _ := json.Marshal(map[string]string{
		"status":  "404",
		"message": "Not Found",
	})

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusNotFound)
	_, _ = response.Write(data)
}

func apiBadRequest(resp http.ResponseWriter, errorsPerField map[string][]string) {
	resp.WriteHeader(http.StatusBadRequest)
	resp.Header().Set("Content-Type", "application/json")
	result, _ := json.Marshal(map[string]interface{}{
		"errors": errorsPerField,
	})
	_, _ = resp.Write(result)
}

func apiForbidden(resp http.ResponseWriter, reason string) {
	resp.WriteHeader(http.StatusForbidden)
	resp.Header().Set("Content-Type", "application/json")
	result, _ := json.Marshal(map[string]interface{}{
		"error": reason,
	})
	_, _ = resp.Write(result)
}

func apiNotFound(resp http.ResponseWriter, reason string) {
	resp.WriteHeader(http.StatusNotFound)
	resp.Header().Set("Content-Type", "application/json")
	result, _ := json.Marshal([]interface{}{
		reason,
	})
	_, _ = resp.Write(result)
}

func apiServerError(resp http.ResponseWriter) {
	resp.WriteHeader(http.StatusInternalServerError)
}
