package di

import (
	"github.com/defval/di"

	. "github.com/elyby/chrly/internal/http"
	"github.com/elyby/chrly/internal/profiles"
)

var profilesDi = di.Options(
	di.Provide(newProfilesManager, di.As(new(ProfilesManager))),
	di.Provide(newProfilesProvider, di.As(new(ProfilesProvider))),
)

func newProfilesManager(r profiles.ProfilesRepository) *profiles.Manager {
	return profiles.NewManager(r)
}

func newProfilesProvider(
	finder profiles.ProfilesFinder,
	mojangProfilesProvider profiles.MojangProfilesProvider,
) *profiles.Provider {
	return &profiles.Provider{
		ProfilesFinder:         finder,
		MojangProfilesProvider: mojangProfilesProvider,
	}
}
