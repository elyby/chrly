package profiles

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elyby/chrly/db"
)

type ProfilesRepositoryMock struct {
	mock.Mock
}

func (m *ProfilesRepositoryMock) FindProfileByUuid(uuid string) (*db.Profile, error) {
	args := m.Called(uuid)
	var result *db.Profile
	if casted, ok := args.Get(0).(*db.Profile); ok {
		result = casted
	}

	return result, args.Error(1)
}

func (m *ProfilesRepositoryMock) SaveProfile(profile *db.Profile) error {
	return m.Called(profile).Error(0)
}

func (m *ProfilesRepositoryMock) RemoveProfileByUuid(uuid string) error {
	return m.Called(uuid).Error(0)
}

type ManagerTestSuite struct {
	suite.Suite

	Manager *Manager

	ProfilesRepository *ProfilesRepositoryMock
}

func (t *ManagerTestSuite) SetupSubTest() {
	t.ProfilesRepository = &ProfilesRepositoryMock{}
	t.Manager = NewManager(t.ProfilesRepository)
}

func (t *ManagerTestSuite) TearDownSubTest() {
	t.ProfilesRepository.AssertExpectations(t.T())
}

func (t *ManagerTestSuite) TestPersistProfile() {
	t.Run("valid profile (full)", func() {
		profile := &db.Profile{
			Uuid:            "ba866a9c-c839-4268-a30f-7b26ae604c51",
			Username:        "mock-username",
			SkinUrl:         "https://example.com/skin.png",
			SkinModel:       "slim",
			CapeUrl:         "https://example.com/cape.png",
			MojangTextures:  "eyJ0aW1lc3RhbXAiOjE0ODYzMzcyNTQ4NzIsInByb2ZpbGVJZCI6ImM0ZjFlNTZmNjFkMTQwYTc4YzMyOGQ5MTY2ZWVmOWU3IiwicHJvZmlsZU5hbWUiOiJXaHlZb3VSZWFkVGhpcyIsInRleHR1cmVzIjp7IlNLSU4iOnsidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS83Mzk1NmE4ZTY0ZWU2ZDhlYzY1NmFkYmI0NDA0ZjhlYmZmMzQxMWIwY2I5MGIzMWNiNDc2ZWNiOTk2ZDNiOCJ9fX0=",
			MojangSignature: "QH+1rlQJYk8tW+8WlSJnzxZZUL5RIkeOO33dq84cgNoxwCkzL95Zy5pbPMFhoiMXXablqXeqyNRZDQa+OewgDBSZxm0BmkNmwdTLzCPHgnlNYhwbO4sirg3hKjCZ82ORZ2q7VP2NQIwNvc3befiCakhDlMWUuhjxe7p/HKNtmKA7a/JjzmzwW7BWMv8b88ZaQaMaAc7puFQcu2E54G2Zk2kyv3T1Bm7bV4m7ymbL8McOmQc6Ph7C95/EyqIK1a5gRBUHPEFIEj0I06YKTHsCRFU1U/hJpk98xXHzHuULJobpajqYXuVJ8QEVgF8k8dn9VkS8BMbXcjzfbb6JJ36v7YIV6Rlt75wwTk2wr3C3P0ij55y0iXth1HjwcEKsg54n83d9w8yQbkUCiTpMbOqxTEOOS7G2O0ZDBJDXAKQ4n5qCiCXKZ4febv4+dWVQtgfZHnpGJUD3KdduDKslMePnECOXMjGSAOQou//yze2EkL2rBpJtAAiOtvBlm/aWnDZpij5cQk+pWmeHWZIf0LSSlsYRUWRDk/VKBvUTEAO9fqOxWqmSgQRUY2Ea56u0ZsBb4vEa1UY6mlJj3+PNZaWu5aP2E9Unh0DIawV96eW8eFQgenlNXHMmXd4aOra4sz2eeOnY53JnJP+eVE4cB1hlq8RA2mnwTtcy3lahzZonOWc=",
		}
		t.ProfilesRepository.On("SaveProfile", profile).Once().Return(nil)

		err := t.Manager.PersistProfile(profile)
		t.NoError(err)
	})

	t.Run("valid profile (minimal)", func() {
		profile := &db.Profile{
			Uuid:     "ba866a9c-c839-4268-a30f-7b26ae604c51",
			Username: "mock-username",
		}
		t.ProfilesRepository.On("SaveProfile", profile).Once().Return(nil)

		err := t.Manager.PersistProfile(profile)
		t.NoError(err)
	})

	t.Run("normalize uuid and skin model", func() {
		profile := &db.Profile{
			Uuid:      "BA866A9C-C839-4268-A30F-7B26AE604C51",
			Username:  "mock-username",
			SkinUrl:   "https://example.com/skin.png",
			SkinModel: "default",
		}
		expectedProfile := *profile
		expectedProfile.Uuid = "ba866a9cc8394268a30f7b26ae604c51"
		expectedProfile.SkinModel = ""
		t.ProfilesRepository.On("SaveProfile", &expectedProfile).Once().Return(nil)

		err := t.Manager.PersistProfile(profile)
		t.NoError(err)
	})

	t.Run("require mojangSignature when mojangTexturesProvided", func() {
		profile := &db.Profile{
			Uuid:           "ba866a9c-c839-4268-a30f-7b26ae604c51",
			Username:       "mock-username",
			MojangTextures: "eyJ0aW1lc3RhbXAiOjE0ODYzMzcyNTQ4NzIsInByb2ZpbGVJZCI6ImM0ZjFlNTZmNjFkMTQwYTc4YzMyOGQ5MTY2ZWVmOWU3IiwicHJvZmlsZU5hbWUiOiJXaHlZb3VSZWFkVGhpcyIsInRleHR1cmVzIjp7IlNLSU4iOnsidXJsIjoiaHR0cDovL3RleHR1cmVzLm1pbmVjcmFmdC5uZXQvdGV4dHVyZS83Mzk1NmE4ZTY0ZWU2ZDhlYzY1NmFkYmI0NDA0ZjhlYmZmMzQxMWIwY2I5MGIzMWNiNDc2ZWNiOTk2ZDNiOCJ9fX0=",
		}

		err := t.Manager.PersistProfile(profile)
		t.Error(err)
		t.IsType(&ValidationError{}, err)
		castedErr := err.(*ValidationError)
		mojangSignatureErr, mojangSignatureErrExists := castedErr.Errors["MojangSignature"]
		t.True(mojangSignatureErrExists)
		t.Contains(mojangSignatureErr[0], "required")
	})

	t.Run("validate username", func() {
		profile := &db.Profile{
			Uuid:     "ba866a9c-c839-4268-a30f-7b26ae604c51",
			Username: "invalid\"username",
		}

		err := t.Manager.PersistProfile(profile)
		t.Error(err)
		t.IsType(&ValidationError{}, err)
		castedErr := err.(*ValidationError)
		usernameErrs, usernameErrExists := castedErr.Errors["Username"]
		t.True(usernameErrExists)
		t.Contains(usernameErrs[0], "valid")
	})

	t.Run("empty profile", func() {
		profile := &db.Profile{}

		err := t.Manager.PersistProfile(profile)
		t.Error(err)
		t.IsType(&ValidationError{}, err)
		// TODO: validate errors
	})
}

func TestManager(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}
