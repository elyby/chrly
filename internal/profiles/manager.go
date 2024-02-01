package profiles

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/elyby/chrly/internal/db"
)

type ProfilesRepository interface {
	FindProfileByUuid(uuid string) (*db.Profile, error)
	SaveProfile(profile *db.Profile) error
	RemoveProfileByUuid(uuid string) error
}

func NewManager(pr ProfilesRepository) *Manager {
	return &Manager{
		ProfilesRepository: pr,
		profileValidator:   createProfileValidator(),
	}
}

type Manager struct {
	ProfilesRepository
	profileValidator *validator.Validate
}

func (m *Manager) PersistProfile(profile *db.Profile) error {
	validationErrors := m.profileValidator.Struct(profile)
	if validationErrors != nil {
		return mapValidationErrorsToCommonError(validationErrors.(validator.ValidationErrors))
	}

	profile.Uuid = cleanupUuid(profile.Uuid)
	if profile.SkinUrl == "" || isClassicModel(profile.SkinModel) {
		profile.SkinModel = ""
	}

	return m.ProfilesRepository.SaveProfile(profile)
}

func (m *Manager) RemoveProfileByUuid(uuid string) error {
	return m.ProfilesRepository.RemoveProfileByUuid(cleanupUuid(uuid))
}

type ValidationError struct {
	Errors map[string][]string
}

func (e *ValidationError) Error() string {
	return "The profile is invalid and cannot be persisted"
}

func cleanupUuid(uuid string) string {
	return strings.ReplaceAll(strings.ToLower(uuid), "-", "")
}

func createProfileValidator() *validator.Validate {
	validate := validator.New()

	regexUuidAny := regexp.MustCompile("(?i)^[a-f0-9]{8}-?[a-f0-9]{4}-?[a-f0-9]{4}-?[a-f0-9]{4}-?[a-f0-9]{12}$")
	_ = validate.RegisterValidation("uuid_any", func(fl validator.FieldLevel) bool {
		return regexUuidAny.MatchString(fl.Field().String())
	})

	regexUsername := regexp.MustCompile(`^[-\w.!$%^&*()\[\]:;]+$`)
	_ = validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return regexUsername.MatchString(fl.Field().String())
	})

	validate.RegisterStructValidationMapRules(map[string]string{
		"Username":        "required,username,max=21",
		"Uuid":            "required,uuid_any",
		"SkinUrl":         "omitempty,url",
		"SkinModel":       "omitempty,max=20",
		"CapeUrl":         "omitempty,url",
		"MojangTextures":  "omitempty,base64",
		"MojangSignature": "required_with=MojangTextures,omitempty,base64",
	}, db.Profile{})

	return validate
}

func mapValidationErrorsToCommonError(err validator.ValidationErrors) *ValidationError {
	resultErr := &ValidationError{make(map[string][]string)}
	for _, e := range err {
		// Manager can return multiple errors per field, but the current validation implementation
		// returns only one error per field
		resultErr.Errors[e.Field()] = []string{formatValidationErr(e)}
	}

	return resultErr
}

// The go-playground/validator lib already contains tools for translated errors output.
// However, the implementation is very heavy and becomes even more so when you need to add messages for custom validators.
// So for simplicity, I've extracted validation error formatting into this simple implementation
func formatValidationErr(err validator.FieldError) string {
	switch err.Tag() {
	case "required", "required_with":
		return fmt.Sprintf("%s is a required field", err.Field())
	case "username":
		return fmt.Sprintf("%s must be a valid username", err.Field())
	case "max":
		return fmt.Sprintf("%s must be a maximum of %s in length", err.Field(), err.Param())
	case "uuid_any":
		return fmt.Sprintf("%s must be a valid UUID", err.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", err.Field())
	case "base64":
		return fmt.Sprintf("%s must be a valid Base64 string", err.Field())
	default:
		return fmt.Sprintf(`Field validation for "%s" failed on the "%s" tag`, err.Field(), err.Tag())
	}
}

func isClassicModel(model string) bool {
	return model == "" || model == "classic" || model == "default" || model == "steve"
}
