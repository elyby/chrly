package mojangtextures

import (
	"encoding/json"
	"github.com/elyby/chrly/version"
	"io/ioutil"
	"net/http"
	. "net/url"
	"path"
	"time"

	"github.com/mono83/slf/wd"

	"github.com/elyby/chrly/api/mojang"
)

var HttpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 1024,
	},
}

type RemoteApiUuidsProvider struct {
	Url    URL
	Logger wd.Watchdog
}

func (ctx *RemoteApiUuidsProvider) GetUuid(username string) (*mojang.ProfileInfo, error) {
	ctx.Logger.IncCounter("mojang_textures.usernames.request", 1)

	url := ctx.Url
	url.Path = path.Join(url.Path, username)

	request, _ := http.NewRequest("GET", url.String(), nil)
	request.Header.Add("Accept", "application/json")
	// Change default User-Agent to allow specify "Username -> UUID at time" Mojang's api endpoint
	request.Header.Add("User-Agent", "Chrly/"+version.Version())

	start := time.Now()
	response, err := HttpClient.Do(request)
	ctx.Logger.RecordTimer("mojang_textures.usernames.request_time", time.Since(start))
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
