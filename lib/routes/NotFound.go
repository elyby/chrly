package routes

import (
	"net/http"
	"encoding/json"
)

func NotFound(w http.ResponseWriter, r *http.Request)  {
	json, _ := json.Marshal(map[string]string{
		"status": "404",
		"message": "Not Found",
		"link": "http://docs.ely.by/skin-system.html",
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write(json)
}
