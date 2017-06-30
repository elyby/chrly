package ui

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"elyby/minecraft-skinsystem/utils"
)

type signedTexturesResponse struct {
	Id    string     `json:"id"`
	Name  string     `json:"name"`
	IsEly bool       `json:"ely,omitempty"`
	Props []property `json:"properties"`
}

type property struct {
	Name string      `json:"name"`
	Signature string `json:"signature,omitempty"`
	Value string     `json:"value"`
}

func (s *uiService) SignedTextures(response http.ResponseWriter, request *http.Request) {
	s.logger.IncCounter("signed_textures.request", 1)
	username := utils.ParseUsername(mux.Vars(request)["username"])

	rec, err := s.skinsRepo.FindByUsername(username)
	if err != nil || rec.SkinId == 0 || rec.MojangTextures == "" {
		response.WriteHeader(http.StatusNoContent)
		return
	}

	responseData:= signedTexturesResponse{
		Id: strings.Replace(rec.Uuid, "-", "", -1),
		Name: rec.Username,
		Props: []property{
			{
				Name: "textures",
				Signature: rec.MojangSignature,
				Value: rec.MojangTextures,
			},
			{
				Name: "ely",
				Value: "but why are you asking?",
			},
		},
	}

	responseJson,_ := json.Marshal(responseData)
	response.Header().Set("Content-Type", "application/json")
	response.Write(responseJson)
}
