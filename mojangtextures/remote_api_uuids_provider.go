package mojangtextures

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	. "net/url"
	"path"

	"github.com/elyby/chrly/api/mojang"
	"github.com/elyby/chrly/version"
)

var HttpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 1024,
	},
}

type RemoteApiUuidsProvider struct {
	Emitter
	Url URL
}

func (ctx *RemoteApiUuidsProvider) GetUuid(username string) (*mojang.ProfileInfo, error) {
	url := ctx.Url
	url.Path = path.Join(url.Path, username)
	urlStr := url.String()

	request, _ := http.NewRequest("GET", urlStr, nil)
	request.Header.Add("Accept", "application/json")
	// Change default User-Agent to allow specify "Username -> UUID at time" Mojang's api endpoint
	request.Header.Add("User-Agent", "Chrly/"+version.Version())

	ctx.Emit("mojang_textures:remote_api_uuids_provider:before_request", urlStr)
	response, err := HttpClient.Do(request)
	ctx.Emit("mojang_textures:remote_api_uuids_provider:after_request", response, err)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == 204 {
		return nil, nil
	}

	if response.StatusCode != 200 {
		return nil, &UnexpectedRemoteApiResponse{response}
	}

	var result *mojang.ProfileInfo
	body, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type UnexpectedRemoteApiResponse struct {
	Response *http.Response
}

func (*UnexpectedRemoteApiResponse) Error() string {
	return "Unexpected remote api response"
}
