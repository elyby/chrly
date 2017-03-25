package worker

import (
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

	record, err := data.FindSkinByUsername(model.OldUsername)
	if (err != nil) {
		services.Logger.IncCounter("worker.change_username.username_not_found", 1)
		// TODO: я не уверен, что это валидное поведение
		// Суть в том, что здесь может возникнуть ошибка в том случае, если записи в базе нету
		// а значит его нужно, как минимум, зарегистрировать
		return true
	}

	record.Username = model.NewUsername
	record.Save()

	services.Logger.IncCounter("worker.change_username.processed", 1)

	return true
}

func handleSkinChanged(model skinChanged) (bool) {
	record, err := data.FindSkinById(model.AccountId)
	if (err != nil) {
		services.Logger.IncCounter("worker.skin_changed.id_not_found", 1)
		return true
	}

	record.SkinId = model.SkinId
	record.Hash   = model.Hash
	record.Is1_8  = model.Is1_8
	record.IsSlim = model.IsSlim
	record.Url    = model.Url

	record.Save()

	services.Logger.IncCounter("worker.skin_changed.processed", 1)

	return true
}
