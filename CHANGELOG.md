# Changelog

## [Unreleased]

### Changed

- docs: SDK SemVer is independent of `sdk-dependencies`; `go.mod` `require` is the pin (this module is already `0.1.1` requiring `v0.1.2`)

## [0.1.1] - 2026-08-12

### Added

- Vault plugin SDK on the common contract: `Serve` (`VAULT_PLUGIN_*`), `ValidateDir` / `ValidateDirDiscover` (must `Supports(vault)`), `PackDir`
- Type aliases for Manifest / Envelope / pack.Result inherited from `bytedesk-sdk-dependencies@v0.1.2`

### Changed

- Identity client (`Client`, enroll, password begin/finish) unchanged; this module is now both plugin SDK and v1 client
