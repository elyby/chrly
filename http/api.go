package http

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/thedevsaddam/govalidator"

	"github.com/elyby/chrly/model"
)

//noinspection GoSnakeCaseUsage
const UUID_ANY = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"

var regexUuidAny = regexp.MustCompile(UUID_ANY)

func init() {
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

type Api struct {
	Emitter
	SkinsRepo SkinsRepository
}

func (ctx *Api) Handler() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/skins", ctx.postSkinHandler).Methods(http.MethodPost)
	router.HandleFunc("/skins/id:{id:[0-9]+}", ctx.deleteSkinByUserIdHandler).Methods(http.MethodDelete)
	router.HandleFunc("/skins/{username}", ctx.deleteSkinByUsernameHandler).Methods(http.MethodDelete)

	return router
}

func (ctx *Api) postSkinHandler(resp http.ResponseWriter, req *http.Request) {
	validationErrors := validatePostSkinRequest(req)
	if validationErrors != nil {
		apiBadRequest(resp, validationErrors)
		return
	}

	identityId, _ := strconv.Atoi(req.Form.Get("identityId"))
	username := req.Form.Get("username")

	record, err := ctx.findIdentityOrCleanup(identityId, username)
	if err != nil {
		ctx.Emit("skinsystem:error", fmt.Errorf("error on requesting a skin from the repository: %w", err))
		apiServerError(resp)
		return
	}

	if record == nil {
		record = &model.Skin{
			UserId:   identityId,
			Username: username,
		}
	}

	skinId, _ := strconv.Atoi(req.Form.Get("skinId"))
	is18, _ := strconv.ParseBool(req.Form.Get("is1_8"))
	isSlim, _ := strconv.ParseBool(req.Form.Get("isSlim"))

	record.Uuid = req.Form.Get("uuid")
	record.SkinId = skinId
	record.Is1_8 = is18
	record.IsSlim = isSlim
	record.Url = req.Form.Get("url")
	record.MojangTextures = req.Form.Get("mojangTextures")
	record.MojangSignature = req.Form.Get("mojangSignature")

	err = ctx.SkinsRepo.Save(record)
	if err != nil {
		ctx.Emit("skinsystem:error", fmt.Errorf("unable to save record to the repository: %w", err))
		apiServerError(resp)
		return
	}

	resp.WriteHeader(http.StatusCreated)
}

func (ctx *Api) deleteSkinByUserIdHandler(resp http.ResponseWriter, req *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(req)["id"])
	skin, err := ctx.SkinsRepo.FindByUserId(id)
	ctx.deleteSkin(skin, err, resp)
}

func (ctx *Api) deleteSkinByUsernameHandler(resp http.ResponseWriter, req *http.Request) {
	username := mux.Vars(req)["username"]
	skin, err := ctx.SkinsRepo.FindByUsername(username)
	ctx.deleteSkin(skin, err, resp)
}

func (ctx *Api) deleteSkin(skin *model.Skin, err error, resp http.ResponseWriter) {
	if err != nil {
		ctx.Emit("skinsystem:error", fmt.Errorf("unable to find skin info from the repository: %w", err))
		apiServerError(resp)
		return
	}

	if skin == nil {
		apiNotFound(resp, "Cannot find record for the requested identifier")
		return
	}

	err = ctx.SkinsRepo.RemoveByUserId(skin.UserId)
	if err != nil {
		ctx.Emit("skinsystem:error", fmt.Errorf("cannot delete skin by error: %w", err))
		apiServerError(resp)
		return
	}

	resp.WriteHeader(http.StatusNoContent)
}

func (ctx *Api) findIdentityOrCleanup(identityId int, username string) (*model.Skin, error) {
	record, err := ctx.SkinsRepo.FindByUserId(identityId)
	if err != nil {
		return nil, err
	}

	if record != nil {
		// The username may have changed in the external database,
		// so we need to remove the old association
		if record.Username != username {
			_ = ctx.SkinsRepo.RemoveByUserId(identityId)
			record.Username = username
		}

		return record, nil
	}

	// If the requested id was not found, then username was reassigned to another user
	// who has not uploaded his data to Chrly yet
	record, err = ctx.SkinsRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}

	// If the target username does exist, clear it as it will be reassigned to the new user
	if record != nil {
		_ = ctx.SkinsRepo.RemoveByUsername(username)
		record.UserId = identityId

		return record, nil
	}

	return nil, nil
}

func validatePostSkinRequest(request *http.Request) map[string][]string {
	const maxMultipartMemory int64 = 32 << 20
	const oneOfSkinOrUrlMessage = "One of url or skin should be provided, but not both"

	_ = request.ParseMultipartForm(maxMultipartMemory)

	validationRules := govalidator.MapData{
		"identityId": {"required", "numeric", "min:1"},
		"username":   {"required"},
		"uuid":       {"required", "uuid_any"},
		"skinId":     {"required", "numeric", "min:1"},
		"url":        {"url"},
		"file:skin":  {"ext:png", "size:24576", "mime:image/png"},
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
