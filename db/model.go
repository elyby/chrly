package db

type Profile struct {
	// Uuid contains user's UUID without dashes in lower case
	Uuid string
	// Username contains user's username with the original casing
	Username string
	// SkinUrl contains a valid URL to user's skin or an empty string in case the user doesn't have a skin
	SkinUrl string
	// SkinModel contains skin's model. It will be empty when the model is default
	SkinModel string
	// CapeUrl contains a valid URL to user's skin or an empty string in case the user doesn't have a cape
	CapeUrl string
	// MojangTextures contains the original textures value from Mojang's skinsystem
	MojangTextures string
	// MojangSignature contains the original textures signature from Mojang's skinsystem
	MojangSignature string
}
