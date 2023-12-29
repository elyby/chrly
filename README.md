# Chrly

[![Written in Go][ico-lang]][link-go]
[![Build Status][ico-build]][link-build]
[![Coverage][ico-coverage]][link-coverage]
[![Keep a Changelog][ico-changelog]](CHANGELOG.md)
[![Software License][ico-license]](LICENSE)

Chrly is a lightweight implementation of Minecraft skins system server with ability to proxy requests to Mojang's
skins system. It's packaged and distributed as a Docker image and can be downloaded from
[Dockerhub](https://hub.docker.com/r/elyby/chrly/). App is written in Go, can withstand heavy loads and is
production ready.

## Installation

You can easily install Chrly using [docker-compose](https://docs.docker.com/compose/). The configuration below (save
it as `docker-compose.yml`) can be used to start a Chrly server. It relies on `CHRLY_SECRET` and `CHRLY_SIGNING_KEY`
environment variables that you must set before running `docker-compose up -d`. Other possible variables are described
below.

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
      CHRLY_SIGNING_KEY: base64:LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlCT3dJQkFBSkJBTmJVcFZDWmtNS3BmdllaMDhXM2x1bWRBYVl4TEJubVVEbHpIQlFIM0RwWWVmNVdDTzMyClREVTZmZUlKNThBMGxBeXdndFo0d3dpMmRHSE96LzFoQXZjQ0F3RUFBUUpBSXRheFNIVGU2UEtieUVVLzlweGoKT05kaFlSWXdWTExvNTZnbk1ZaGt5b0VxYWFNc2ZvdjhoaG9lcGtZWkJNdlpGQjJiRE9zUTJTYUorRTJlaUJPNApBUUloQVBzc1MwK0JSOXcwYk9kbWpHcW1kRTlOck41VUpRY09XMTNzMjkrNlF6VUJBaUVBMnZXT2VwQTVBcGl1CnBFQTNwd29HZGtWQ3JOU25uS2pEUXpEWEJucGQzL2NDSUVGTmQ5c1k0cVVHNEZXZFhONlJubVhMN1NqMHVaZkgKRE13enU4ckVNNXNCQWlFQWh2ZG9ETnFMbWJNZHEzYytGc1BTT2VMMWQyMVpwL0pLOGtiUHRGbUhOZjhDSVFEVgo2RlNaRHd2V2Z1eGFNN0JzeWNRT05rakRCVFBOdStscWN0SkJHbkJ2M0E9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=

  redis:
    image: redis:4.0-32bit
    restart: always
    volumes:
      - ./data/redis:/data
```

**Tip**: to generate a value for the `CHRLY_SIGNING_KEY` use the command below and then join it with a `base64:` prefix.
```sh
openssl genrsa 4096 | base64 -w0
```

Chrly uses some volumes to persist storage for capes and Redis database. The configuration above mounts them to
the host machine to do not lose data on container recreations.

### Config

Application's configuration is based on the environment variables. You can adjust config by modifying `environment` key
inside your `docker-compose.yml` file. After value will have been changed, container should be stopped and recreated.
If environment variables have been changed, Docker will automatically recreate the container, so you only need to `up`
it again:

```sh
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

Each endpoint that accepts `username` as a part of an url takes it case-insensitive. The `.png` postfix can be omitted.

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

#### `GET /profile/{username}`

This endpoint behaves exactly like the
[Mojang's UUID -> Profile + Skin/Cape endpoint](https://wiki.vg/Mojang_API#UUID_-.3E_Profile_.2B_Skin.2FCape), but using
a username instead of the UUID. Just like in the Mojang's API, you can append `?unsigned=false` part to URL to sign
the `textures` property. If the textures for the requested username aren't found, it'll request them through the
Mojang's API, but the Mojang's signature will be discarded and the textures will be re-signed using the signature key
for your Chrly instance.

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

The base64 `value` string for the `textures` property decoded:

```json
{
    "timestamp": 1614387238630,
    "profileId": "0f657aa8bfbe415db7005750090d3af3",
    "profileName": "username",
    "textures": {
        "SKIN": {
            "url": "http://example.com/skin.png"
        },
        "CAPE": {
            "url": "http://example.com/cape.png"
        }
    }
}
```

If username can't be found locally and can't be obtained from the Mojang's API, empty response with `204` status code
will be sent.

Note that this endpoint will try to use the UUID for the stored profile in the database. This is an edge case, related
to the situation where the user is available in the database but has no textures, which caused them to be retrieved
from the Mojang's API.

#### `GET /signature-verification-key.der`

This endpoint returns a public key that can be used to verify textures signatures. The key is provided in `DER` format,
so it can be used directly in the Authlib, without modifying the signature checking algorithm.

#### `GET /signature-verification-key.pem`

The same endpoint as the previous one, except that it returns the key in `PEM` format.

#### `GET /textures/signed/{username}`

Actually, this is the [Ely.by](https://ely.by)'s feature called
[Server Skins System](https://ely.by/server-skins-system), but if you have your own source of Mojang's signatures,
then you can pass it with textures and it'll be displayed in response of this endpoint. Received response should be
directly sent to the client without any modification via game server API.

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

Endpoint allows you to create or update skin record for a username.

The request body must be encoded as `application/x-www-form-urlencoded`.

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
| url             | string | Actual url of the skin.                                                        |

**Important**: all parameters are always read at least as their default values. So, if you only want to update the username and not pass the skin data it will reset all skin information. If you want to keep the data, you should always pass the full set of parameters.

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

Then you must fork this repository. Now follow these steps:

```sh
# Get the source code
git clone https://github.com/elyby/chrly.git
# Switch to the project folder
cd chrly
# Install dependencies
go mod download
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

[ico-lang]: https://img.shields.io/github/go-mod/go-version/elyby/chrly?style=flat-square
[ico-build]: https://img.shields.io/github/actions/workflow/status/elyby/chrly/build.yml?style=flat-square
[ico-coverage]: https://img.shields.io/codecov/c/github/elyby/chrly.svg?style=flat-square
[ico-changelog]: https://img.shields.io/badge/keep%20a-changelog-orange.svg?style=flat-square
[ico-license]: https://img.shields.io/github/license/elyby/chrly.svg?style=flat-square

[link-go]: https://golang.org
[link-build]: https://github.com/elyby/chrly/actions
[link-coverage]: https://codecov.io/gh/elyby/chrly
