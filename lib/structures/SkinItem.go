package structures

type SkinItem struct {
	UserId   int    `json:"userId"`
	Nickname string `json:"nickname"`
	SkinId   int    `json:"skinId"`
	Url      string `json:"url"`
	Is1_8    bool   `json:"is1_8"`
	IsSlim   bool   `json:"isSlim"`
	Hash     string `json:"hash"`
}
