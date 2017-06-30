package ui

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Метод-наследие от первой версии системы скинов.
// Всё ещё иногда используется
// Просто конвертируем данные и отправляем их в основной обработчик
func (s *uiService) MinecraftPHP(response http.ResponseWriter, request *http.Request) {
	username := request.URL.Query().Get("name")
	required := request.URL.Query().Get("type")
	if username == "" || required == "" {
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	mux.Vars(request)["username"] = username
	mux.Vars(request)["converted"] = "1"

	switch required {
	case "skin":
		s.logger.IncCounter("skins.minecraft-php-request", 1)
		s.Skin(response, request)
	case "cloack":
		s.logger.IncCounter("capes.minecraft-php-request", 1)
		s.Cape(response, request)
	default:
		response.WriteHeader(http.StatusNotFound)
	}
}
