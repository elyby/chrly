package http

import (
	"encoding/json"
	"net/http"
)

func (cfg *Config) NotFound(response http.ResponseWriter, request *http.Request) {
	data, _ := json.Marshal(map[string]string{
		"status":  "404",
		"message": "Not Found",
	})

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusNotFound)
	response.Write(data)
}
