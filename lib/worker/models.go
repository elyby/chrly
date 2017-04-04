package worker

type usernameChanged struct {
	AccountId   int    `json:"accountId"`
	OldUsername string `json:"oldUsername"`
	NewUsername string `json:"newUsername"`
}

type skinChanged struct {
	AccountId       int    `json:"userId"`
	Uuid            string `json:"uuid"`
	SkinId          int    `json:"skinId"`
	OldSkinId       int    `json:"oldSkinId"`
	Hash            string `json:"hash"`
	Is1_8           bool   `json:"is1_8"`
	IsSlim          bool   `json:"isSlim"`
	Url             string `json:"url"`
	MojangTextures  string `json:"mojangTextures"`
	MojangSignature string `json:"mojangSignature"`
}
