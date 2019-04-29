# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
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
- `/textures` request now doesn't proxies request to Mojang in case when there is no information about the skin,
  but there is a cape.

### Removed
- `hash` field from `/textures` response because the game doesn't use it and calculates hash by getting filename
  from the textures link.

[Unreleased]: https://github.com/elyby/chrly/compare/4.1.1...HEAD
