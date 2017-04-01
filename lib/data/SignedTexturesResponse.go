package data

type SignedTexturesResponse struct {
	Id    string     `json:"id"`
	Name  string     `json:"name"`
	IsEly bool       `json:"ely,omitempty"`
	Props []Property `json:"properties"`
}

type Property struct {
	Name string      `json:"name"`
	Signature string `json:"signature"`
	Value string     `json:"value"`
}
