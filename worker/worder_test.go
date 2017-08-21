package worker

import (
	"testing"

	"github.com/golang/mock/gomock"
	testify "github.com/stretchr/testify/assert"

	"elyby/minecraft-skinsystem/api/accounts"
	"elyby/minecraft-skinsystem/db"
	"elyby/minecraft-skinsystem/interfaces/mock_interfaces"
	"elyby/minecraft-skinsystem/interfaces/mock_wd"
	"elyby/minecraft-skinsystem/model"
)

func TestServices_HandleChangeUsername(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	services, skinRepo, _, wd := setupMocks(ctrl)

	resultModel := createSourceModel()
	resultModel.Username = "new_username"

	// Запись о скине существует, никаких осложнений
	skinRepo.EXPECT().FindByUserId(1).Return(createSourceModel(), nil)
	skinRepo.EXPECT().Save(resultModel)
	wd.EXPECT().IncCounter("worker.change_username", int64(1))

	assert.True(services.HandleChangeUsername(&model.UsernameChanged{
		AccountId: 1,
		OldUsername: "mock_user",
		NewUsername: "new_username",
	}))

	// Событие с пустым ником, т.е это регистрация, так что нужно создать запись о скине
	skinRepo.EXPECT().FindByUserId(1).Times(0)
	skinRepo.EXPECT().Save(&model.Skin{UserId: 1, Username: "new_mock"})
	wd.EXPECT().IncCounter("worker.change_username", int64(1))
	wd.EXPECT().IncCounter("worker.change_username_empty_old_username", int64(1))

	assert.True(services.HandleChangeUsername(&model.UsernameChanged{
		AccountId: 1,
		OldUsername: "",
		NewUsername: "new_mock",
	}))

	// В базе системы скинов нет записи об указанном пользователе, так что её нужно восстановить
	skinRepo.EXPECT().FindByUserId(1).Return(nil, &db.SkinNotFoundError{})
	skinRepo.EXPECT().Save(&model.Skin{UserId: 1, Username: "new_mock2"})
	wd.EXPECT().IncCounter("worker.change_username", int64(1))
	wd.EXPECT().IncCounter("worker.change_username_id_not_found", int64(1))

	assert.True(services.HandleChangeUsername(&model.UsernameChanged{
		AccountId: 1,
		OldUsername: "mock_user",
		NewUsername: "new_mock2",
	}))
}

func TestServices_HandleSkinChanged(t *testing.T) {
	assert := testify.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	services, skinRepo, accountsAPI, wd := setupMocks(ctrl)

	event := &model.SkinChanged{
		AccountId: 1,
		Uuid: "cdb907ce-84f4-4c38-801d-1e287dca2623",
		SkinId: 2,
		OldSkinId: 1,
		Hash: "f76caa016e07267a05b7daf9ebc7419c",
		Is1_8: true,
		IsSlim: false,
		Url: "http://ely.by/minecraft/skins/69c6740d2993e5d6f6a7fc92420efc29.png",
		MojangTextures: "new mocked textures base64",
		MojangSignature: "new mocked signature",
	}

	resultModel := createSourceModel()
	resultModel.SkinId = event.SkinId
	resultModel.Hash = event.Hash
	resultModel.Is1_8 = event.Is1_8
	resultModel.IsSlim = event.IsSlim
	resultModel.Url = event.Url
	resultModel.MojangTextures = event.MojangTextures
	resultModel.MojangSignature = event.MojangSignature

	// Запись о скине существует, никаких осложнений
	skinRepo.EXPECT().FindByUserId(1).Return(createSourceModel(), nil)
	skinRepo.EXPECT().Save(resultModel)
	wd.EXPECT().IncCounter("worker.skin_changed", int64(1))

	assert.True(services.HandleSkinChanged(event))

	// Записи о скине не существует, она должна быть восстановлена
	skinRepo.EXPECT().FindByUserId(1).Return(nil, &db.SkinNotFoundError{"mock_user"})
	skinRepo.EXPECT().Save(resultModel)
	accountsAPI.EXPECT().AccountInfo("id", "1").Return(&accounts.AccountInfoResponse{
		Id: 1,
		Username: "mock_user",
		Uuid: "cdb907ce-84f4-4c38-801d-1e287dca2623",
		Email: "mock-user@ely.by",
	}, nil)
	wd.EXPECT().IncCounter("worker.skin_changed", int64(1))
	wd.EXPECT().IncCounter("worker.skin_changed_id_not_found", int64(1))
	wd.EXPECT().IncCounter("worker.skin_changed_id_restored", int64(1))
	wd.EXPECT().Warning(gomock.Any())
	wd.EXPECT().Info(gomock.Any())

	assert.True(services.HandleSkinChanged(event))

	// Записи о скине не существует, и Ely.by Accounts internal API не знает о таком пользователе
	skinRepo.EXPECT().FindByUserId(1).Return(nil, &db.SkinNotFoundError{"mock_user"})
	accountsAPI.EXPECT().AccountInfo("id", "1").Return(nil, &accounts.NotFoundResponse{})
	wd.EXPECT().IncCounter("worker.skin_changed", int64(1))
	wd.EXPECT().IncCounter("worker.skin_changed_id_not_found", int64(1))
	wd.EXPECT().IncCounter("worker.skin_changed_id_not_restored", int64(1))
	wd.EXPECT().Warning(gomock.Any())
	wd.EXPECT().Error(gomock.Any())

	assert.True(services.HandleSkinChanged(event))
}

func createSourceModel() *model.Skin {
	return &model.Skin{
		UserId: 1,
		Uuid: "cdb907ce-84f4-4c38-801d-1e287dca2623",
		Username: "mock_user",
		SkinId: 1,
		Url: "http://ely.by/minecraft/skins/3a345c701f473ac08c8c5b8ecb58ecf3.png",
		Is1_8: false,
		IsSlim: false,
		Hash: "3a345c701f473ac08c8c5b8ecb58ecf3",
		MojangTextures: "mocked textures base64",
		MojangSignature: "mocked signature",
	}
}

func setupMocks(ctrl *gomock.Controller) (
	*Services,
	*mock_interfaces.MockSkinsRepository,
	*mock_interfaces.MockAccountsAPI,
	*mock_wd.MockWatchdog,
) {
	skinsRepo := mock_interfaces.NewMockSkinsRepository(ctrl)
	accountApi := mock_interfaces.NewMockAccountsAPI(ctrl)
	wd := mock_wd.NewMockWatchdog(ctrl)

	return &Services{
		SkinsRepo: skinsRepo,
		AccountsAPI: accountApi,
		Logger: wd,
	}, skinsRepo, accountApi, wd
}
