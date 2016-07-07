package routes

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Метод-наследие от первой версии системы скинов.
// Всё ещё иногда используется
// Просто конвертируем данные и отправляем их в основной обработчик
func MinecraftPHP(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("name")
	required := r.URL.Query().Get("type")
	if username == "" || required == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(r)["username"] = username
	switch required {
	case "skin": Skin(w, r)
	case "cloack": Cape(w, r)
	default: {
		w.WriteHeader(http.StatusNotFound)
	}
	}
}
