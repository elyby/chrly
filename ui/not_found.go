package ui

import (
	"encoding/json"
	"net/http"
)

func NotFound(response http.ResponseWriter, request *http.Request)  {
	json, _ := json.Marshal(map[string]string{
		"status": "404",
		"message": "Not Found",
		"link": "http://docs.ely.by/skin-system.html",
	})

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusNotFound)
	response.Write(json)
}
