package routes

import (
	"net/http"
	"strings"
	"strconv"

	"elyby/minecraft-skinsystem/lib/data"
)

func SetSkin(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("X-Ely-key")
	if key != "43fd2ce61b3f5704dfd729c1f2d6ffdb" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Nice try"))
		return
	}

	skin := new(data.SkinItem)
	skin.Nickname  = strings.ToLower(r.PostFormValue("nickname"))
	skin.UserId, _ = strconv.Atoi(r.PostFormValue("userId"))
	skin.SkinId, _ = strconv.Atoi(r.PostFormValue("skinId"))
	skin.Hash      = r.PostFormValue("hash")
	skin.Is1_8, _  = strconv.ParseBool(r.PostFormValue("is1_8"))
	skin.IsSlim, _ = strconv.ParseBool(r.PostFormValue("isSlim"))
	skin.Url       = r.PostFormValue("url")
	skin.Save()

	w.Write([]byte("OK"))
}
