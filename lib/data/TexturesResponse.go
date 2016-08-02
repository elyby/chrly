package data

type TexturesResponse struct {
	Skin *Skin `json:"SKIN"`
}

type Skin struct {
	Url      string        `json:"url"`
	Hash     string        `json:"hash"`
	Metadata *SkinMetadata `json:"metadata,omitempty"`
}

type SkinMetadata struct {
	Model string `json:"model"`
}
