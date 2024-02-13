package mojang

import "context"

type MojangUuidsStorage interface {
	// The second argument must be returned as a incoming username in case,
	// when cached result indicates that there is no Mojang user with provided username
	GetUuidForMojangUsername(ctx context.Context, username string) (foundUuid string, foundUsername string, err error)
	// An empty uuid value can be passed if the corresponding account has not been found
	StoreMojangUuid(ctx context.Context, username string, uuid string) error
}

type UuidsProviderWithCache struct {
	Provider UuidsProvider
	Storage  MojangUuidsStorage
}

func (p *UuidsProviderWithCache) GetUuid(ctx context.Context, username string) (*ProfileInfo, error) {
	uuid, foundUsername, err := p.Storage.GetUuidForMojangUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if foundUsername != "" {
		if uuid != "" {
			return &ProfileInfo{Id: uuid, Name: foundUsername}, nil
		}

		return nil, nil
	}

	profile, err := p.Provider.GetUuid(ctx, username)
	if err != nil {
		return nil, err
	}

	freshUuid := ""
	wellCasedUsername := username
	if profile != nil {
		freshUuid = profile.Id
		wellCasedUsername = profile.Name
	}

	_ = p.Storage.StoreMojangUuid(ctx, wellCasedUsername, freshUuid)

	return profile, nil
}
