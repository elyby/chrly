package interfaces

import (
	"elyby/minecraft-skinsystem/api/accounts"
)

type AccountsAPI interface {
	AccountInfo(attribute string, value string) (*accounts.AccountInfoResponse, error)
}
