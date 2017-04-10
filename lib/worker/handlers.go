package worker

import (
	"fmt"
	"elyby/minecraft-skinsystem/lib/data"
	"elyby/minecraft-skinsystem/lib/services"
)

func handleChangeUsername(model usernameChanged) (bool) {
	if (model.OldUsername == "") {
		services.Logger.IncCounter("worker.change_username.empty_old_username", 1)
		record := data.SkinItem{
			UserId: model.AccountId,
			Username: model.NewUsername,
		}

		record.Save()

		return true
	}

	record, err := data.FindSkinById(model.AccountId)
	if (err != nil) {
		services.Logger.IncCounter("worker.change_username.id_not_found", 1)
		fmt.Println("Cannot find user id. Trying to search.")
		response, err := getById(model.AccountId)
		if err != nil {
			services.Logger.IncCounter("worker.change_username.id_not_restored", 1)
			fmt.Printf("Cannot restore user info. %T\n", err)
			// TODO: логгировать в какой-нибудь Sentry, если там не 404
			return true
		}

		services.Logger.IncCounter("worker.change_username.id_restored", 1)
		fmt.Println("User info successfully restored.")
		record = data.SkinItem{
			UserId: response.Id,
		}
	}

	record.Username = model.NewUsername
	record.Save()

	services.Logger.IncCounter("worker.change_username.processed", 1)

	return true
}

func handleSkinChanged(model skinChanged) bool {
	record, err := data.FindSkinById(model.AccountId)
	if err != nil {
		services.Logger.IncCounter("worker.skin_changed.id_not_found", 1)
		fmt.Println("Cannot find user id. Trying to search.")
		response, err := getById(model.AccountId)
		if err != nil {
			services.Logger.IncCounter("worker.skin_changed.id_not_restored", 1)
			fmt.Printf("Cannot restore user info. %T\n", err)
			// TODO: логгировать в какой-нибудь Sentry, если там не 404
			return true
		}

		services.Logger.IncCounter("worker.skin_changed.id_restored", 1)
		fmt.Println("User info successfully restored.")
		record.UserId = response.Id
		record.Username = response.Username
	}

	record.Uuid = model.Uuid
	record.SkinId = model.SkinId
	record.Hash = model.Hash
	record.Is1_8 = model.Is1_8
	record.IsSlim = model.IsSlim
	record.Url = model.Url
	record.MojangTextures = model.MojangTextures
	record.MojangSignature = model.MojangSignature

	record.Save()

	services.Logger.IncCounter("worker.skin_changed.processed", 1)

	return true
}
