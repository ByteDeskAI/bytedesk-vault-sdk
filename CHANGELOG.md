# Changelog

## [0.1.1] - 2026-08-12

### Added

- Vault plugin SDK on the common contract: `Serve` (`VAULT_PLUGIN_*`), `ValidateDir` / `ValidateDirDiscover` (must `Supports(vault)`), `PackDir`
- Type aliases for Manifest / Envelope / pack.Result inherited from `bytedesk-sdk-dependencies@v0.1.2`

### Changed

- Identity client (`Client`, enroll, password begin/finish) unchanged; this module is now both plugin SDK and v1 client
