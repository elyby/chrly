# Chrly

[![Written in Go][ico-lang]][link-go]
[![Build Status][ico-build]][link-build]
[![Keep a Changelog][ico-changelog]](CHANGELOG.md)
[![Software License][ico-license]](LICENSE)

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

| ENV                | Description                                                                        | Example                                   |
|--------------------|------------------------------------------------------------------------------------|-------------------------------------------|
| STORAGE_REDIS_POOL | By default, Chrly creates pool with 10 connection, but you may want to increase it | `20`                                      |
| STATSD_ADDR        | StatsD can be used to collect metrics                                              | `localhost:8125`                          |
| SENTRY_DSN         | Sentry can be used to collect app errors                                           | `https://public:private@your.sentry.io/1` |

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
| hash            | string | Skin's hash. Algorithm can be any. For example `md5`.                          |
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
To run tests execute `go test ./...`. If your Go version is older than 1.9, then run a `/script/test`.

[ico-lang]: https://img.shields.io/badge/lang-go%201.12-blue.svg?style=flat-square
[ico-build]: https://img.shields.io/travis/elyby/chrly.svg?style=flat-square
[ico-changelog]: https://img.shields.io/badge/keep%20a-changelog-orange.svg?style=flat-square
[ico-license]: https://img.shields.io/github/license/elyby/chrly.svg?style=flat-square

[link-go]: https://golang.org
[link-build]: https://travis-ci.org/elyby/chrly
