package model

type Skin struct {
	UserId          int    `json:"userId"`
	Uuid            string `json:"uuid"`
	Username        string `json:"username"`
	SkinId          int    `json:"skinId"`
	Url             string `json:"url"`
	Is1_8           bool   `json:"is1_8"`
	IsSlim          bool   `json:"isSlim"`
	Hash            string `json:"hash"`
	MojangTextures  string `json:"mojangTextures"`
	MojangSignature string `json:"mojangSignature"`
	OldUsername     string
}
