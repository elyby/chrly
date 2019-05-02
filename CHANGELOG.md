# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

### Fixed
- `/textures` request no longer proxies request to Mojang in a case when there is no information about the skin,
  but there is a cape.
- [#5](https://github.com/elyby/chrly/issues/5): Return Redis connection to the pool after commands are executed

### Removed
- `hash` field from `/textures` response because the game doesn't use it and calculates hash by getting the filename
  from the textures link instead.
- `hash` field from `POST /api/skins` endpoint.

[Unreleased]: https://github.com/elyby/chrly/compare/4.2.0...HEAD
[4.2.0]: https://github.com/elyby/chrly/compare/4.1.1...4.2.0
