package di

import (
	"github.com/defval/di"

	. "ely.by/chrly/internal/http"
	"ely.by/chrly/internal/profiles"
)

var profilesDiOptions = di.Options(
	di.Provide(newProfilesManager, di.As(new(ProfilesManager))),
	di.Provide(newProfilesProvider, di.As(new(ProfilesProvider))),
)

func newProfilesManager(r profiles.ProfilesRepository) *profiles.Manager {
	return profiles.NewManager(r)
}

func newProfilesProvider(
	finder profiles.ProfilesFinder,
	mojangProfilesProvider profiles.MojangProfilesProvider,
) (*profiles.Provider, error) {
	return profiles.NewProvider(
		finder,
		mojangProfilesProvider,
	)
}
