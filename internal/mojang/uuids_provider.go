package mojang

type MojangUuidsStorage interface {
	// The second argument must be returned as a incoming username in case,
	// when cached result indicates that there is no Mojang user with provided username
	GetUuidForMojangUsername(username string) (foundUuid string, foundUsername string, err error)
	// An empty uuid value can be passed if the corresponding account has not been found
	StoreMojangUuid(username string, uuid string) error
}

type UuidsProviderWithCache struct {
	Provider UuidsProvider
	Storage  MojangUuidsStorage
}

func (p *UuidsProviderWithCache) GetUuid(username string) (*ProfileInfo, error) {
	uuid, foundUsername, err := p.Storage.GetUuidForMojangUsername(username)
	if err != nil {
		return nil, err
	}

	if foundUsername != "" {
		if uuid != "" {
			return &ProfileInfo{Id: uuid, Name: foundUsername}, nil
		}

		return nil, nil
	}

	profile, err := p.Provider.GetUuid(username)
	if err != nil {
		return nil, err
	}

	freshUuid := ""
	wellCasedUsername := username
	if profile != nil {
		freshUuid = profile.Id
		wellCasedUsername = profile.Name
	}

	_ = p.Storage.StoreMojangUuid(wellCasedUsername, freshUuid)

	return profile, nil
}
