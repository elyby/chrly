package worker

import (
	"elyby/minecraft-skinsystem/lib/data"
	"log"
)

func handleChangeUsername(model usernameChanged) (bool) {
	if (model.OldUsername == "") {
		record := data.SkinItem{
			UserId: model.AccountId,
			Username: model.NewUsername,
		}

		record.Save()

		return true
	}

	record, err := data.FindSkinByUsername(model.OldUsername)
	if (err != nil) {
		log.Println("Exit by not found record")
		// TODO: я не уверен, что это валидное поведение
		// Суть в том, что здесь может возникнуть ошибка в том случае, если записи в базе нету
		// а значит его нужно, как минимум, зарегистрировать
		return true
	}

	record.Username = model.NewUsername
	record.Save()

	log.Println("all saved!")

	return true
}

func handleSkinChanged(model skinChanged) (bool) {
	record, err := data.FindSkinById(model.AccountId)
	if (err != nil) {
		return true
	}

	record.SkinId = model.SkinId
	record.Hash   = model.Hash
	record.Is1_8  = model.Is1_8
	record.IsSlim = model.IsSlim
	record.Url    = model.Url

	record.Save()

	return true
}
