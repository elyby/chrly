package structures

type Skin struct {
	Url      string `json:"url"`
	Hash     string `json:"hash"`
	Metadata *SkinMetadata `json:"metadata,omitempty"`
}
