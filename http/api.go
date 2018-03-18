package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/elyby/chrly/auth"
	"github.com/elyby/chrly/db"
	"github.com/elyby/chrly/interfaces"
	"github.com/elyby/chrly/model"

	"github.com/gorilla/mux"
	"github.com/mono83/slf/wd"
	"github.com/thedevsaddam/govalidator"
)

//noinspection GoSnakeCaseUsage
const UUID_ANY = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
var regexUuidAny = regexp.MustCompile(UUID_ANY)

func init() {
	govalidator.AddCustomRule("md5", func(field string, rule string, message string, value interface{}) error {
		val := []byte(value.(string))
		if ok, _ := regexp.Match(`^[a-f0-9]{32}$`, val); !ok {
			if message == "" {
				message = fmt.Sprintf("The %s field must be a valid md5 hash", field)
			}

			return errors.New(message)
		}

		return nil
	})

	govalidator.AddCustomRule("skinUploadingNotAvailable", func(field string, rule string, message string, value interface{}) error {
		if message == "" {
			message = "Skin uploading is temporary unavailable"
		}

		return errors.New(message)
	})

	// Add ability to validate any possible uuid form
	govalidator.AddCustomRule("uuid_any", func(field string, rule string, message string, value interface{}) error {
		str := value.(string)
		if !regexUuidAny.MatchString(str) {
			if message == "" {
				message = fmt.Sprintf("The %s field must contain valid UUID", field)
			}

			return errors.New(message)
		}

		return nil
	})
}

func (cfg *Config) PostSkin(resp http.ResponseWriter, req *http.Request) {
	cfg.Logger.IncCounter("api.skins.post.request", 1)
	validationErrors := validatePostSkinRequest(req)
	if validationErrors != nil {
		cfg.Logger.IncCounter("api.skins.post.validation_failed", 1)
		apiBadRequest(resp, validationErrors)
		return
	}

	identityId, _ := strconv.Atoi(req.Form.Get("identityId"))
	username := req.Form.Get("username")

	record, err := findIdentity(cfg.SkinsRepo, identityId, username)
	if err != nil {
		cfg.Logger.Error("Error on requesting a skin from the repository: :err", wd.ErrParam(err))
		apiServerError(resp)
		return
	}

	skinId, _ := strconv.Atoi(req.Form.Get("skinId"))
	is18, _ := strconv.ParseBool(req.Form.Get("is1_8"))
	isSlim, _ := strconv.ParseBool(req.Form.Get("isSlim"))

	record.Uuid = req.Form.Get("uuid")
	record.SkinId = skinId
	record.Hash = req.Form.Get("hash")
	record.Is1_8 = is18
	record.IsSlim = isSlim
	record.Url = req.Form.Get("url")
	record.MojangTextures = req.Form.Get("mojangTextures")
	record.MojangSignature = req.Form.Get("mojangSignature")

	err = cfg.SkinsRepo.Save(record)
	if err != nil {
		cfg.Logger.Error("Unable to save record to the repository: :err", wd.ErrParam(err))
		apiServerError(resp)
		return
	}

	cfg.Logger.IncCounter("api.skins.post.success", 1)
	resp.WriteHeader(http.StatusCreated)
}

func (cfg *Config) DeleteSkinByUserId(resp http.ResponseWriter, req *http.Request) {
	cfg.Logger.IncCounter("api.skins.delete.request", 1)
	id, _ := strconv.Atoi(mux.Vars(req)["id"])
	skin, err := cfg.SkinsRepo.FindByUserId(id)
	if err != nil {
		cfg.Logger.IncCounter("api.skins.delete.not_found", 1)
		apiNotFound(resp, "Cannot find record for requested user id")
		return
	}

	cfg.deleteSkin(skin, resp)
}

func (cfg *Config) DeleteSkinByUsername(resp http.ResponseWriter, req *http.Request) {
	cfg.Logger.IncCounter("api.skins.delete.request", 1)
	username := mux.Vars(req)["username"]
	skin, err := cfg.SkinsRepo.FindByUsername(username)
	if err != nil {
		cfg.Logger.IncCounter("api.skins.delete.not_found", 1)
		apiNotFound(resp, "Cannot find record for requested username")
		return
	}

	cfg.deleteSkin(skin, resp)
}

func (cfg *Config) Authenticate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		cfg.Logger.IncCounter("authentication.challenge", 1)
		err := cfg.Auth.Check(req)
		if err != nil {
			if _, ok := err.(*auth.Unauthorized); ok {
				cfg.Logger.IncCounter("authentication.failed", 1)
				apiForbidden(resp, err.Error())
			} else {
				cfg.Logger.Error("Unknown error on validating api request: :err", wd.ErrParam(err))
				apiServerError(resp)
			}

			return
		}

		cfg.Logger.IncCounter("authentication.success", 1)
		handler.ServeHTTP(resp, req)
	})
}

func (cfg *Config) deleteSkin(skin *model.Skin, resp http.ResponseWriter) {
	err := cfg.SkinsRepo.RemoveByUserId(skin.UserId)
	if err != nil {
		cfg.Logger.Error("Cannot delete skin by error: :err", wd.ErrParam(err))
		apiServerError(resp)
		return
	}

	cfg.Logger.IncCounter("api.skins.delete.success", 1)
	resp.WriteHeader(http.StatusNoContent)
}

func validatePostSkinRequest(request *http.Request) map[string][]string {
	const maxMultipartMemory int64 = 32 << 20
	const oneOfSkinOrUrlMessage = "One of url or skin should be provided, but not both"

	request.ParseMultipartForm(maxMultipartMemory)

	validationRules := govalidator.MapData{
		"identityId": {"required", "numeric", "min:1"},
		"username":   {"required"},
		"uuid":       {"required", "uuid_any"},
		"skinId":     {"required", "numeric", "min:1"},
		"url":        {"url"},
		"file:skin":  {"ext:png", "size:24576", "mime:image/png"},
		"hash":       {"md5"},
		"is1_8":      {"bool"},
		"isSlim":     {"bool"},
	}

	shouldAppendSkinRequiredError := false
	url := request.Form.Get("url")
	_, _, skinErr := request.FormFile("skin")
	if (url != "" && skinErr == nil) || (url == "" && skinErr != nil) {
		shouldAppendSkinRequiredError = true
	} else if skinErr == nil {
		validationRules["file:skin"] = append(validationRules["file:skin"], "skinUploadingNotAvailable")
	} else if url != "" {
		validationRules["hash"] = append(validationRules["hash"], "required")
		validationRules["is1_8"] = append(validationRules["is1_8"], "required")
		validationRules["isSlim"] = append(validationRules["isSlim"], "required")
	}

	mojangTextures := request.Form.Get("mojangTextures")
	if mojangTextures != "" {
		validationRules["mojangSignature"] = []string{"required"}
	}

	validator := govalidator.New(govalidator.Options{
		Request:         request,
		Rules:           validationRules,
		RequiredDefault: false,
		FormSize:        maxMultipartMemory,
	})
	validationResults := validator.Validate()
	if shouldAppendSkinRequiredError {
		validationResults["url"] = append(validationResults["url"], oneOfSkinOrUrlMessage)
		validationResults["skin"] = append(validationResults["skin"], oneOfSkinOrUrlMessage)
	}

	if len(validationResults) != 0 {
		return validationResults
	}

	return nil
}

func findIdentity(repo interfaces.SkinsRepository, identityId int, username string) (*model.Skin, error) {
	var record *model.Skin
	record, err := repo.FindByUserId(identityId)
	if err != nil {
		if _, isSkinNotFound := err.(*db.SkinNotFoundError); !isSkinNotFound {
			return nil, err
		}

		record, err = repo.FindByUsername(username)
		if err == nil {
			repo.RemoveByUsername(username)
			record.UserId = identityId
		} else {
			record = &model.Skin{
				UserId:   identityId,
				Username: username,
			}
		}
	} else if record.Username != username {
		repo.RemoveByUserId(identityId)
		record.Username = username
	}

	return record, nil
}

func apiBadRequest(resp http.ResponseWriter, errorsPerField map[string][]string) {
	resp.WriteHeader(http.StatusBadRequest)
	resp.Header().Set("Content-Type", "application/json")
	result, _ := json.Marshal(map[string]interface{}{
		"errors": errorsPerField,
	})
	resp.Write(result)
}

func apiForbidden(resp http.ResponseWriter, reason string) {
	resp.WriteHeader(http.StatusForbidden)
	resp.Header().Set("Content-Type", "application/json")
	result, _ := json.Marshal(map[string]interface{}{
		"error": reason,
	})
	resp.Write(result)
}

func apiNotFound(resp http.ResponseWriter, reason string) {
	resp.WriteHeader(http.StatusNotFound)
	resp.Header().Set("Content-Type", "application/json")
	result, _ := json.Marshal([]interface{}{
		reason,
	})
	resp.Write(result)
}

func apiServerError(resp http.ResponseWriter) {
	resp.WriteHeader(http.StatusInternalServerError)
}
