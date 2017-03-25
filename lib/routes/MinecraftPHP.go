package routes

import (
	"net/http"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/lib/services"
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
	mux.Vars(r)["converted"] = "1"
	switch required {
	case "skin":
		services.Logger.IncCounter("skins.minecraft-php-request", 1)
		Skin(w, r)
	case "cloack":
		services.Logger.IncCounter("capes.minecraft-php-request", 1)
		Cape(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
