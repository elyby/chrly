package data

type TexturesResponse struct {
	Skin *Skin `json:"SKIN"`
	Cape *Cape `json:"CAPE,omitempty"`
}

type Skin struct {
	Url      string        `json:"url"`
	Hash     string        `json:"hash"`
	Metadata *SkinMetadata `json:"metadata,omitempty"`
}

type SkinMetadata struct {
	Model string `json:"model"`
}

type Cape struct {
	Url      string        `json:"url"`
	Hash     string        `json:"hash"`
}
