# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - xxxx-xx-xx
### Added
- `/profile/{username}` endpoint, which returns a profile and its textures, equivalent of the Mojang's
  [UUID -> Profile + Skin/Cape endpoint](https://wiki.vg/Mojang_API#UUID_-.3E_Profile_.2B_Skin.2FCape).
- `/signature-verification-key.der` and `/signature-verification-key.pem` endpoints, which returns the public key in
  `DER` or `PEM` formats for signature verification.

### Fixed
- [#28](https://github.com/elyby/chrly/issues/28): Added handling of corrupted data from the Mojang's username to UUID
  cache.
- [#29](https://github.com/elyby/chrly/issues/29): If a previously cached UUID no longer exists,
  it will be invalidated and re-requested.
- Use correct status code for error about empty response from Mojang's API.

### Changed
- **BREAKING**: `/cloaks/{username}` and `/textures/{username}` endpoints will no longer return a cape if there are no
  textures for the requested username.
- All endpoints are now returns `500` status code when an error occurred during request processing.
- Increased the response timeout for Mojang's API from 3 to 10 seconds.

## [4.5.0] - 2020-05-01
### Added
- [#24](https://github.com/elyby/chrly/issues/24): Implemented a new strategy for the queue in the batch provider of
  Mojang UUIDs: `full-bus`.
- New configuration param `QUEUE_STRATEGY` with the default value `periodic`.
- New configuration params: `MOJANG_API_BASE_URL` and `MOJANG_SESSION_SERVER_BASE_URL`, that allow you to spoof
  Mojang API base addresses.
- New health checker, that ensures that response for textures provider from Mojang's API is valid.
- `dev` Docker images now have the `--cpuprofile` flag, which allows you to run the program with CPU profiling.
- New StatsD metrics:
  - Gauges:
    - `ely.skinsystem.{hostname}.app.redis.pool.available`

### Fixed
- Handle the case when there is no textures property in Mojang's response.
- Handle `SIGTERM` as a valid stop signal for a graceful shutdown since it's the default stop code for the Docker.
- Default connections pool size for Redis.

### Changed
- `ely.skinsystem.{hostname}.app.mojang_textures.usernames.round_time` timer will not be recorded if the iteration was
  empty.

## [4.4.1] - 2020-04-24
### Added
- [#20](https://github.com/elyby/chrly/issues/20): Print hostname in the `version` command output.
- [#21](https://github.com/elyby/chrly/issues/21): Print Chrly's version during server startup.

### Fixed
- [#22](https://github.com/elyby/chrly/issues/22): Correct version passing during building of the Docker image.

## [4.4.0] - 2020-04-22
### Added
- Mojang textures queue now can be completely disabled via `MOJANG_TEXTURES_ENABLED` param.
- Remote mode for Mojang's textures queue with a new configuration params: `MOJANG_TEXTURES_UUIDS_PROVIDER_DRIVER` and
  `MOJANG_TEXTURES_UUIDS_PROVIDER_URL`.

  For example, to send requests directly to [Mojang's APIs](https://wiki.vg/Mojang_API#Username_-.3E_UUID_at_time),
  set the next configuration:
  - `MOJANG_TEXTURES_UUIDS_PROVIDER_DRIVER=remote`
  - `MOJANG_TEXTURES_UUIDS_PROVIDER_URL=https://api.mojang.com/users/profiles/minecraft/`
- Implemented worker mode. The app starts with the only one API endpoint: `/api/worker/mojang-uuid/{username}`,
  which is compatible with [Mojang's endpoint](https://wiki.vg/Mojang_API#Username_-.3E_UUID_at_time) to exchange
  username to its UUID. It can be used with some load balancing software to increase throughput of Mojang's textures
  proxy by splitting the load across multiple servers with its own IPs.
- Textures extra param is now can be configured via `TEXTURES_EXTRA_PARAM_NAME` and `TEXTURES_EXTRA_PARAM_VALUE`.
- New StatsD metrics:
  - Counters:
    - `ely.skinsystem.{hostname}.app.mojang_textures.usernames.textures_hit`
    - `ely.skinsystem.{hostname}.app.mojang_textures.usernames.textures_miss`
- All incoming requests are now logging to the console in
  [Apache Common Log Format](http://httpd.apache.org/docs/2.2/logs.html#common).
- Added `/healthcheck` endpoint.
- Graceful server shutdown.
- Panics in http are now logged in Sentry.

### Fixed
- `ely.skinsystem.{hostname}.app.mojang_textures.usernames.iteration_size` and
  `ely.skinsystem.{hostname}.app.mojang_textures.usernames.queue_size` are now updates even if the queue is empty.
- Don't return an empty object if Mojang's textures don't contain any skin or cape.
- Provides a correct URL scheme for the cape link.

### Changed
- **BREAKING**: `QUEUE_LOOP_DELAY` param is now sets as a Go duration, not milliseconds.
  For example, default value is now `2s500ms`.
- **BREAKING**: Event `ely.skinsystem.{hostname}.app.mojang_textures.already_in_queue` has been renamed into
  `ely.skinsystem.{hostname}.app.mojang_textures.already_scheduled`.
- Bumped Go version to 1.14.

### Removed
- **BREAKING**: `ely.skinsystem.{hostname}.app.mojang_textures.invalid_username` counter has been removed.

## [4.3.0] - 2019-11-08
### Added
- 403 Forbidden errors from the Mojang's API are now logged.
- `QUEUE_LOOP_DELAY` configuration param to adjust Mojang's textures queue performance.

### Changed
- Mojang's textures queue loop is now has an iteration delay of 2.5 seconds (was 1).
- Bumped Go version to 1.13.

## [4.2.3] - 2019-10-03
### Changed
- Mojang's textures queue batch size [reduced to 10](https://wiki.vg/index.php?title=Mojang_API&type=revision&diff=14964&oldid=14954).
- 400 BadRequest errors from the Mojang's API are now logged.

## [4.2.2] - 2019-06-19
### Fixed
- GC for in-memory textures cache has not been initialized.

## [4.2.1] - 2019-05-06
### Changed
- Improved Keep-Alive settings for HTTP client used to perform requests to Mojang's APIs.
- Mojang's textures queue now has static delay of 1 second after each iteration to prevent strange `429` errors.
- Mojang's textures queue now caches even errored responses for signed textures to avoid `429` errors.
- Mojang's textures queue now caches textures data for 70 seconds to avoid strange `429` errors.
- Mojang's textures queue now doesn't log timeout errors.

### Fixed
- Panic when Redis connection is broken.
- Duplication of Redis connections pool for Mojang's textures queue.
- Removed validation rules for `hash` field.

## [4.2.0] - 2019-05-02
### Added
- `CHANGELOG.md` file.
- [#1](https://github.com/elyby/chrly/issues/1): Restored Mojang skins proxy.
- New StatsD metrics:
  - Counters:
    - `ely.skinsystem.{hostname}.app.mojang_textures.invalid_username`
    - `ely.skinsystem.{hostname}.app.mojang_textures.request`
    - `ely.skinsystem.{hostname}.app.mojang_textures.usernames.cache_hit_nil`
    - `ely.skinsystem.{hostname}.app.mojang_textures.usernames.queued`
    - `ely.skinsystem.{hostname}.app.mojang_textures.usernames.cache_hit`
    - `ely.skinsystem.{hostname}.app.mojang_textures.already_in_queue`
    - `ely.skinsystem.{hostname}.app.mojang_textures.usernames.uuid_miss`
    - `ely.skinsystem.{hostname}.app.mojang_textures.usernames.uuid_hit`
    - `ely.skinsystem.{hostname}.app.mojang_textures.textures.cache_hit`
    - `ely.skinsystem.{hostname}.app.mojang_textures.textures.request`
  - Gauges:
    - `ely.skinsystem.{hostname}.app.mojang_textures.usernames.iteration_size`
    - `ely.skinsystem.{hostname}.app.mojang_textures.usernames.queue_size`
  - Timers:
    - `ely.skinsystem.{hostname}.app.mojang_textures.result_time`
    - `ely.skinsystem.{hostname}.app.mojang_textures.usernames.round_time`
    - `ely.skinsystem.{hostname}.app.mojang_textures.textures.request_time`

### Changed
- Bumped Go version to 1.12.
- Bumped Alpine version to 3.9.3.

### Fixed
- `/textures` request no longer proxies request to Mojang in a case when there is no information about the skin,
  but there is a cape.
- [#5](https://github.com/elyby/chrly/issues/5): Return Redis connection to the pool after commands are executed

### Removed
- `hash` field from `/textures` response because the game doesn't use it and calculates hash by getting the filename
  from the textures link instead.
- `hash` field from `POST /api/skins` endpoint.

[Unreleased]: https://github.com/elyby/chrly/compare/4.5.0...HEAD
[4.5.0]: https://github.com/elyby/chrly/compare/4.4.1...4.5.0
[4.4.1]: https://github.com/elyby/chrly/compare/4.4.0...4.4.1
[4.4.0]: https://github.com/elyby/chrly/compare/4.3.0...4.4.0
[4.3.0]: https://github.com/elyby/chrly/compare/4.2.3...4.3.0
[4.2.3]: https://github.com/elyby/chrly/compare/4.2.2...4.2.3
[4.2.2]: https://github.com/elyby/chrly/compare/4.2.1...4.2.2
[4.2.1]: https://github.com/elyby/chrly/compare/4.2.0...4.2.1
[4.2.0]: https://github.com/elyby/chrly/compare/4.1.1...4.2.0
