# Chrly

[![Written in Go][ico-lang]][link-go]
[![Build Status][ico-build]][link-build]
[![Coverage][ico-coverage]][link-coverage]
[![Keep a Changelog][ico-changelog]](CHANGELOG.md)
[![Software License][ico-license]](LICENSE)
[![FOSSA Status][ico-fossa]][link-fossa]

Chrly is a lightweight implementation of Minecraft skins system server with ability to proxy requests to Mojang's
skins system. It's packaged and distributed as a Docker image and can be downloaded from
[Dockerhub](https://hub.docker.com/r/elyby/chrly/). App is written in Go, can withstand heavy loads and is
production ready.

## Installation

You can easily install Chrly using [docker-compose](https://docs.docker.com/compose/). The configuration below (save
it as `docker-compose.yml`) can be used to start a Chrly server. It relies on `CHRLY_SECRET` environment variable
that you must set before running `docker-compose up -d`. Other possible variables are described below.

```yml
version: '2'
services:
  app:
    image: elyby/chrly
    hostname: chrly0
    restart: always
    links:
      - redis
    volumes:
      - ./data/capes:/data/capes
    ports:
      - "80:80"
    environment:
      CHRLY_SECRET: replace_this_value_in_production

  redis:
    image: redis:4.0-32bit
    restart: always
    volumes:
      - ./data/redis:/data
```

Chrly uses some volumes to persist storage for capes and Redis database. The configuration above mounts them to
the host machine to do not lose data on container recreations.

### Config

Application's configuration is based on the environment variables. You can adjust config by modifying `environment` key
inside your `docker-compose.yml` file. After value will have been changed, container should be stopped and recreated.
If environment variables have been changed, Docker will automatically recreate the container, so you only need to `stop`
and `up` it:

```sh
docker-compose stop app
docker-compose up -d app
```

**Variables to adjust:**

<table>
<thead>
    <tr>
        <th>ENV</th>
        <th>Description</th>
        <th>Example</th>
    </tr>
</thead>
<tbody>
    <tr>
        <td>STORAGE_REDIS_HOST</td>
        <td>
            By default, Chrly tries to connect to the <code>redis</code> host
            (by service name in docker-compose configuration).
        </td>
        <td><code>localhost</code></td>
    </tr>
    <tr>
        <td>STORAGE_REDIS_PORT</td>
        <td>
            Specifies the Redis connection port.
        </td>
        <td><code>6379</code></td>
    </tr>
    <tr>
        <td>STORAGE_REDIS_POOL</td>
        <td>By default, Chrly creates pool with 10 connection, but you may want to increase it</td>
        <td><code>20</code></td>
    </tr>
    <tr>
        <td>STATSD_ADDR</td>
        <td>StatsD can be used to collect metrics</td>
        <td><code>localhost:8125</code></td>
    </tr>
    <tr>
        <td>SENTRY_DSN</td>
        <td>Sentry can be used to collect app errors</td>
        <td><code>https://public:private@your.sentry.io/1</code></td>
    </tr>
    <tr>
        <td>QUEUE_STRATEGY</td>
        <td>
            Sets the strategy for the queue in the batch provider of Mojang UUIDs. Allowed values are <code>periodic</code>
            and <code>full-bus</code> (see <a href="https://github.com/elyby/chrly/issues/24">#24</a>).
        </td>
        <td><code>periodic</code></td>
    </tr>
    <tr>
        <td>QUEUE_LOOP_DELAY</td>
        <td>
            Parameter is sets the delay before each iteration of the Mojang's textures queue
            (<a href="https://golang.org/pkg/time/#ParseDuration">Go's duration</a>)
        </td>
        <td><code>3s200ms</code></td>
    </tr>
    <tr>
        <td>QUEUE_BATCH_SIZE</td>
        <td>
            Sets the count of usernames, which will be sent to the
            <a href="https://wiki.vg/Mojang_API#Playernames_-.3E_UUIDs">Mojang's API to exchange them to their UUIDs</a>.
            The current limit is <code>10</code>, but it may change in the future, so you may want to adjust it.
        </td>
        <td><code>10</code></td>
    </tr>
    <tr>
        <td>MOJANG_TEXTURES_ENABLED</td>
        <td>
            Allows to completely disable Mojang textures provider for unknown usernames. Enabled by default.
        </td>
        <td><code>true</code></td>
    </tr>
    <tr>
        <td id="remote-mojang-uuids-provider">MOJANG_TEXTURES_UUIDS_PROVIDER_DRIVER</td>
        <td>
            Specifies the preferred provider of the Mojang's UUIDs. Takes <code>remote</code> value.
            In any other case, the local queue will be used.
        </td>
        <td><code>remote</code></td>
    </tr>
    <tr>
        <td>MOJANG_TEXTURES_UUIDS_PROVIDER_URL</td>
        <td>
            When the UUIDs driver set to <code>remote</code>, sets the remote URL.
            The trailing slash won't cause any problems.
        </td>
        <td><code>http://remote-provider.com/api/worker/mojang-uuid</code></td>
    </tr>
    <tr>
        <td>MOJANG_API_BASE_URL</td>
        <td>
            Allows you to spoof the Mojang's API server address.
        </td>
        <td><code>https://api.mojang.com</code></td>
    </tr>
    <tr>
        <td>MOJANG_SESSION_SERVER_BASE_URL</td>
        <td>
            Allows you to spoof the Mojang's Session server address.
        </td>
        <td><code>https://sessionserver.mojang.com</code></td>
    </tr>
    <tr>
        <td>TEXTURES_EXTRA_PARAM_NAME</td>
        <td>
            Sets the name of the extra property in the
            <a href="#get-texturessignedusername">signed textures</a> response.
        </td>
        <td><code>your-name</code></td>
    </tr>
    <tr>
        <td>TEXTURES_EXTRA_PARAM_VALUE</td>
        <td>
            Sets the value of the extra property in the
            <a href="#get-texturessignedusername">signed textures</a> response.
        </td>
        <td><code>your awesome joke!</code></td>
    </tr>
</tbody>
</table>

If something goes wrong, you can always access logs by executing `docker-compose logs -f app`.

## Endpoints

Each endpoint that accepts `username` as a part of an url takes it case insensitive. `.png` part can be omitted too.

#### `GET /skins/{username}.png`

This endpoint responds to requested `username` with a skin texture. If user's skin was set as texture's link, then it'll
respond with the `301` redirect to that url. If the skin entry isn't found, it'll request textures information from
Mojang's API and if it has a skin, than it'll return a `301` redirect to it.

#### `GET /cloaks/{username}.png`

It responds to requested `username` with a cape texture. If the cape entry isn't found, it'll request textures
information from Mojang's API and if it has a cape, than it'll return a `301` redirect to it.

#### `GET /textures/{username}`

This endpoint forms response payloads as if it was the `textures`' property, but without base64 encoding. For example:

```json
{
    "SKIN": {
        "url": "http://example.com/skin.png",
        "metadata": {
            "model": "slim"
        }
    },
    "CAPE": {
        "url": "http://example.com/cape.png"
    }
}
```

If both the skin and the cape entries aren't found, it'll request textures information from Mojang's API and if it has
a textures property, than it'll return decoded contents.

That request is handy in case when your server implements authentication for a game server (e.g. join/hasJoined
operation) and you have to respond with hasJoined request with an actual user textures. You have to simply send request
to the Chrly server and put the result in your hasJoined response.

#### `GET /textures/signed/{username}`

Actually, it's [Ely.by](http://ely.by) feature called [Server Skins System](http://ely.by/server-skins-system), but if
you have your own source of Mojang's signatures, then you can pass it with textures and it'll be displayed in response
of this endpoint. Received response should be directly sent to the client without any modification via game server API.

Response example:

```json
{
    "id": "0f657aa8bfbe415db7005750090d3af3",
    "name": "username",
    "properties": [
        {
            "name": "textures",
            "signature": "signature value",
            "value": "base64 encoded value"
        },
        {
            "name": "chrly",
            "value": "how do you tame a horse in Minecraft?"
        }
    ]
}
```

If there is no requested `username` or `mojangSignature` field isn't set, `204` status code will be sent.

You can adjust URL to `/textures/signed/{username}?proxy=true` to obtain textures information for provided username
from Mojang's API. The textures will contain unmodified json with addition property with name "chrly" as shown in
the example above.

#### `GET /skins?name={username}`

Equivalent of the `GET /skins/{username}.png`, but constructed especially for old Minecraft versions, where username
placeholder wasn't used.

#### `GET /cloaks?name={username}`

Equivalent of the `GET /cloaks/{username}.png`, but constructed especially for old Minecraft versions, where username
placeholder wasn't used.

### Records manipulating API

Each request to the internal API should be performed with the Bearer authorization header. Example curl request:

```sh
curl -X POST -i http://chrly.domain.com/api/skins \
  -H "Authorization: Bearer Ym9zY236Ym9zY28="
```

You can obtain token by executing `docker-compose run --rm app token`.

#### `POST /api/skins`

> **Warning**: skin uploading via `skin` field is not implemented for now.

Endpoint allows you to create or update skin record for a username. To upload skin, you have to send multipart
form data. `form-urlencoded` also supported, but, as you may know, it doesn't support files uploading.

**Request params:**

| Field           | Type   | Description                                                                    |
|-----------------|--------|--------------------------------------------------------------------------------|
| identityId      | int    | Unique record identifier.                                                      |
| username        | string | Username. Case insensitive.                                                    |
| uuid            | uuid   | UUID of the user.                                                              |
| skinId          | int    | Skin identifier.                                                               |
| is1_8           | bool   | Does the skin have the new format (64x64).                                     |
| isSlim          | bool   | Does skin have slim arms (Alex model).                                         |
| mojangTextures  | string | Mojang textures field. It must be a base64 encoded json string. Not required.  |
| mojangSignature | string | Signature for Mojang textures, which is required when `mojangTextures` passed. |
| url             | string | Actual url of the skin. You have to pass this parameter or `skin`.             |
| skin            | file   | Skin file. You have to pass this parameter or `url`.                           |

If successful you'll receive `201` status code. In the case of failure there will be `400` status code and errors list
as json:

```json
{
    "errors": {
        "identityId": [
            "The identityId field must be numeric"
        ]
    }
}
```

#### `DELETE /api/skins/id:{identityId}`

Performs record removal by identity id. Request body is not required. On success you will receive `204` status code.
On failure it'll be `404` with the json body:

```json
{
    "error": "Cannot find record for requested user id"
}
```

#### `DELETE /api/skins/{username}`

Same endpoint as above but it removes record by identity's username. Have the same behavior, but in case of failure
response will be:

```json
{
    "error": "Cannot find record for requested username"
}
```

### Worker mode

The worker mode can be used in cooperation with the [remote server mode](#remote-mojang-uuids-provider)
to exchange Mojang usernames to UUIDs. This mode by itself doesn't solve the problem of
[extremely strict limits](https://github.com/elyby/chrly/issues/10) on the number of requests to the Mojang's API.
But with a proxying load balancer (e.g. HAProxy, Nginx, etc.) it's easy to build a cluster of workers,
which will multiply the bandwidth of the exchanging usernames to its UUIDs.

The instructions for setting up a proxy load balancer are outside the context of this documentation,
but you get the idea ;)

#### `GET /api/worker/mojang-uuid/{username}`

Performs [batch usernames exchange to UUIDs](https://github.com/elyby/chrly/issues/1) and returns the result in the
[same format as it returns from the Mojang's API](https://wiki.vg/Mojang_API#Username_-.3E_UUID_at_time):

```json
{
    "id": "3e3ee6c35afa48abb61e8cd8c42fc0d9",
    "name": "ErickSkrauch"
}
```

> **Note**: the results aren't cached.

### Health check

#### `GET /healthcheck`

This endpoint can be used to programmatically check the status of the server.
If all internal checks are successful, the server will return `200` status code with the following body:

```json
{
    "status": "OK"
}
```

If any of the checks fails, the server will return `503` status code with the following body:

```json
{
    "status": "Service Unavailable",
    "errors": {
        "mojang-batch-uuids-provider-queue-length": "the maximum number of tasks in the queue has been exceeded"
    }
}
```

## Development

First of all you should install the [latest stable version of Go](https://golang.org/doc/install) and set `GOPATH`
environment variable.

This project uses [`dep`](https://github.com/golang/dep) for dependencies management, so it
[should be installed](https://github.com/golang/dep#installation) too.

Then you must fork this repository. Now follow these steps:

```sh
# Get the source code
go get github.com/elyby/chrly
# Switch to the project folder
cd $GOPATH/src/github.com/elyby/chrly
# Install dependencies (it can take a while)
dep ensure
# Add your fork link as a remote
git remote add fork git@github.com:your-username/chrly.git
# Create a new branch for your task
git checkout -b iss-123
```

You only need to execute `go run main.go` to run the project, but without Redis database and a secret key it won't work
for very long. You have to export `CHRLY_SECRET` environment variable globally or pass it via `env`:

```sh
env CHRLY_SECRET=some_local_secret go run main.go serve
```

Redis can be installed manually, but if you have [Docker installed](https://docs.docker.com/install/), you can run
predefined docker-compose service. Simply execute the next commands:

```sh
cp docker-compose.dev.yml docker-compose.yml
docker-compose up -d
```

If your Redis instance isn't located at the `localhost`, you can change host by editing environment variable
`STORAGE_REDIS_HOST`.

After all of that `go run main.go serve` should successfully start the application.
To run tests execute `go test ./...`.

## License
[![FOSSA Status][ico-fossa-big]][link-fossa]

[ico-lang]: https://img.shields.io/badge/lang-go%201.14-blue.svg?style=flat-square
[ico-build]: https://img.shields.io/travis/elyby/chrly.svg?style=flat-square
[ico-coverage]: https://img.shields.io/codecov/c/github/elyby/chrly.svg?style=flat-square
[ico-changelog]: https://img.shields.io/badge/keep%20a-changelog-orange.svg?style=flat-square
[ico-license]: https://img.shields.io/github/license/elyby/chrly.svg?style=flat-square
[ico-fossa]: https://app.fossa.io/api/projects/git%2Bgithub.com%2Felyby%2Fchrly.svg?type=shield
[ico-fossa-big]: https://app.fossa.io/api/projects/git%2Bgithub.com%2Felyby%2Fchrly.svg?type=large

[link-go]: https://golang.org
[link-build]: https://travis-ci.org/elyby/chrly
[link-coverage]: https://codecov.io/gh/elyby/chrly
[link-fossa]: https://app.fossa.io/projects/git%2Bgithub.com%2Felyby%2Fchrly
