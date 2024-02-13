package profiles

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"ely.by/chrly/internal/db"
	"ely.by/chrly/internal/mojang"
	"ely.by/chrly/internal/utils"
)

type ProfilesFinderMock struct {
	mock.Mock
}

func (m *ProfilesFinderMock) FindProfileByUsername(ctx context.Context, username string) (*db.Profile, error) {
	args := m.Called(ctx, username)
	var result *db.Profile
	if casted, ok := args.Get(0).(*db.Profile); ok {
		result = casted
	}

	return result, args.Error(1)
}

type MojangProfilesProviderMock struct {
	mock.Mock
}

func (m *MojangProfilesProviderMock) GetForUsername(ctx context.Context, username string) (*mojang.ProfileResponse, error) {
	args := m.Called(ctx, username)
	var result *mojang.ProfileResponse
	if casted, ok := args.Get(0).(*mojang.ProfileResponse); ok {
		result = casted
	}

	return result, args.Error(1)
}

type CombinedProfilesProviderSuite struct {
	suite.Suite

	Provider *Provider

	ProfilesFinder         *ProfilesFinderMock
	MojangProfilesProvider *MojangProfilesProviderMock
}

func (t *CombinedProfilesProviderSuite) SetupSubTest() {
	t.ProfilesFinder = &ProfilesFinderMock{}
	t.MojangProfilesProvider = &MojangProfilesProviderMock{}
	t.Provider = &Provider{
		ProfilesFinder:         t.ProfilesFinder,
		MojangProfilesProvider: t.MojangProfilesProvider,
	}
}

func (t *CombinedProfilesProviderSuite) TearDownSubTest() {
	t.ProfilesFinder.AssertExpectations(t.T())
	t.MojangProfilesProvider.AssertExpectations(t.T())
}

func (t *CombinedProfilesProviderSuite) TestFindByUsername() {
	t.Run("exists profile with a skin", func() {
		ctx := context.Background()
		profile := &db.Profile{
			Uuid:     "mock-uuid",
			Username: "Mock",
			SkinUrl:  "https://example.com/skin.png",
		}
		t.ProfilesFinder.On("FindProfileByUsername", ctx, "Mock").Return(profile, nil)

		foundProfile, err := t.Provider.FindProfileByUsername(ctx, "Mock", true)
		t.NoError(err)
		t.Same(profile, foundProfile)
	})

	t.Run("exists profile with a cape", func() {
		profile := &db.Profile{
			Uuid:     "mock-uuid",
			Username: "Mock",
			CapeUrl:  "https://example.com/cape.png",
		}
		t.ProfilesFinder.On("FindProfileByUsername", mock.Anything, "Mock").Return(profile, nil)

		foundProfile, err := t.Provider.FindProfileByUsername(context.Background(), "Mock", true)
		t.NoError(err)
		t.Same(profile, foundProfile)
	})

	t.Run("exists profile without textures (no proxy)", func() {
		profile := &db.Profile{
			Uuid:     "mock-uuid",
			Username: "Mock",
		}
		t.ProfilesFinder.On("FindProfileByUsername", mock.Anything, "Mock").Return(profile, nil)

		foundProfile, err := t.Provider.FindProfileByUsername(context.Background(), "Mock", false)
		t.NoError(err)
		t.Same(profile, foundProfile)
	})

	t.Run("not exists profile (no proxy)", func() {
		t.ProfilesFinder.On("FindProfileByUsername", mock.Anything, "Mock").Return(nil, nil)

		foundProfile, err := t.Provider.FindProfileByUsername(context.Background(), "Mock", false)
		t.NoError(err)
		t.Nil(foundProfile)
	})

	t.Run("handle error from profiles repository", func() {
		expectedError := errors.New("mock error")
		t.ProfilesFinder.On("FindProfileByUsername", mock.Anything, "Mock").Return(nil, expectedError)

		foundProfile, err := t.Provider.FindProfileByUsername(context.Background(), "Mock", false)
		t.Same(expectedError, err)
		t.Nil(foundProfile)
	})

	t.Run("exists profile without textures (with proxy)", func() {
		profile := &db.Profile{
			Uuid:     "mock-uuid",
			Username: "Mock",
		}
		mojangProfile := createMojangProfile(true, true)
		ctx := context.Background()
		t.ProfilesFinder.On("FindProfileByUsername", ctx, "Mock").Return(profile, nil)
		t.MojangProfilesProvider.On("GetForUsername", ctx, "Mock").Return(mojangProfile, nil)

		foundProfile, err := t.Provider.FindProfileByUsername(ctx, "Mock", true)
		t.NoError(err)
		t.Equal(&db.Profile{
			Uuid:            "mock-mojang-uuid",
			Username:        "mOcK",
			SkinUrl:         "https://mojang/skin.png",
			SkinModel:       "slim",
			CapeUrl:         "https://mojang/cape.png",
			MojangTextures:  mojangProfile.Props[0].Value,
			MojangSignature: mojangProfile.Props[0].Signature,
		}, foundProfile)
	})

	t.Run("not exists profile (with proxy)", func() {
		mojangProfile := createMojangProfile(true, true)
		t.ProfilesFinder.On("FindProfileByUsername", mock.Anything, "Mock").Return(nil, nil)
		t.MojangProfilesProvider.On("GetForUsername", mock.Anything, "Mock").Return(mojangProfile, nil)

		foundProfile, err := t.Provider.FindProfileByUsername(context.Background(), "Mock", true)
		t.NoError(err)
		t.Equal(&db.Profile{
			Uuid:            "mock-mojang-uuid",
			Username:        "mOcK",
			SkinUrl:         "https://mojang/skin.png",
			SkinModel:       "slim",
			CapeUrl:         "https://mojang/cape.png",
			MojangTextures:  mojangProfile.Props[0].Value,
			MojangSignature: mojangProfile.Props[0].Signature,
		}, foundProfile)
	})

	t.Run("should return known profile without textures when received an error from the mojang", func() {
		profile := &db.Profile{
			Uuid:     "mock-uuid",
			Username: "Mock",
		}
		t.ProfilesFinder.On("FindProfileByUsername", mock.Anything, "Mock").Return(profile, nil)
		t.MojangProfilesProvider.On("GetForUsername", mock.Anything, "Mock").Return(nil, errors.New("mock error"))

		foundProfile, err := t.Provider.FindProfileByUsername(context.Background(), "Mock", true)
		t.NoError(err)
		t.Same(profile, foundProfile)
	})

	t.Run("should not return an error when passed the invalid username", func() {
		t.ProfilesFinder.On("FindProfileByUsername", mock.Anything, "Mock").Return(nil, nil)
		t.MojangProfilesProvider.On("GetForUsername", mock.Anything, "Mock").Return(nil, mojang.InvalidUsername)

		foundProfile, err := t.Provider.FindProfileByUsername(context.Background(), "Mock", true)
		t.NoError(err)
		t.Nil(foundProfile)
	})

	t.Run("should return an error from mojang provider", func() {
		expectedError := errors.New("mock error")
		t.ProfilesFinder.On("FindProfileByUsername", mock.Anything, "Mock").Return(nil, nil)
		t.MojangProfilesProvider.On("GetForUsername", mock.Anything, "Mock").Return(nil, expectedError)

		foundProfile, err := t.Provider.FindProfileByUsername(context.Background(), "Mock", true)
		t.Same(expectedError, err)
		t.Nil(foundProfile)
	})

	t.Run("should correctly handle invalid textures from mojang", func() {
		mojangProfile := &mojang.ProfileResponse{
			Props: []*mojang.Property{
				{
					Name:      "textures",
					Value:     "this is invalid base64",
					Signature: "mojang signature",
				},
			},
		}
		t.ProfilesFinder.On("FindProfileByUsername", mock.Anything, "Mock").Return(nil, nil)
		t.MojangProfilesProvider.On("GetForUsername", mock.Anything, "Mock").Return(mojangProfile, nil)

		foundProfile, err := t.Provider.FindProfileByUsername(context.Background(), "Mock", true)
		t.ErrorContains(err, "illegal base64 data")
		t.Nil(foundProfile)
	})

	t.Run("should correctly handle missing textures property from Mojang", func() {
		mojangProfile := &mojang.ProfileResponse{
			Id:    "mock-mojang-uuid",
			Name:  "mOcK",
			Props: []*mojang.Property{},
		}
		t.ProfilesFinder.On("FindProfileByUsername", mock.Anything, "Mock").Return(nil, nil)
		t.MojangProfilesProvider.On("GetForUsername", mock.Anything, "Mock").Return(mojangProfile, nil)

		foundProfile, err := t.Provider.FindProfileByUsername(context.Background(), "Mock", true)
		t.NoError(err)
		t.Equal(&db.Profile{
			Uuid:     "mock-mojang-uuid",
			Username: "mOcK",
		}, foundProfile)
	})
}

func TestProvider(t *testing.T) {
	suite.Run(t, new(CombinedProfilesProviderSuite))
}

func createMojangProfile(withSkin bool, withCape bool) *mojang.ProfileResponse {
	timeZone, _ := time.LoadLocation("Europe/Warsaw")
	textures := &mojang.TexturesProp{
		Timestamp:   utils.UnixMillisecond(time.Date(2024, 1, 29, 13, 34, 12, 0, timeZone)),
		ProfileID:   "mock-mojang-uuid",
		ProfileName: "mOcK",
		Textures:    &mojang.TexturesResponse{},
	}

	if withSkin {
		textures.Textures.Skin = &mojang.SkinTexturesResponse{
			Url: "https://mojang/skin.png",
			Metadata: &mojang.SkinTexturesMetadata{
				Model: "slim",
			},
		}
	}

	if withCape {
		textures.Textures.Cape = &mojang.CapeTexturesResponse{
			Url: "https://mojang/cape.png",
		}
	}

	response := &mojang.ProfileResponse{
		Id:   textures.ProfileID,
		Name: textures.ProfileName,
		Props: []*mojang.Property{
			{
				Name:      "textures",
				Value:     mojang.EncodeTextures(textures),
				Signature: "mojang signature",
			},
		},
	}

	return response
}
