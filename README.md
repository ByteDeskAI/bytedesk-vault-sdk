# bytedesk-vault-sdk

Vault **plugin SDK** plus the Vault **v1** identity client.

Inherits **all** common objects and plugin requirements from
`bytedesk-sdk-dependencies` (Manifest, Validate, LoadDir, pack, serve,
bus.Envelope, semver). This module only orchestrates vault host differences:

- env: `VAULT_PLUGIN_SOCKET`, `VAULT_PLUGIN_ID`
- `plugin.json` `targets` must include `vault` (empty targets stay gateway-only)
- identity client: enroll + password begin/finish

```go
import vaultsdk "github.com/ByteDeskAI/bytedesk-vault-sdk"

// Process plugin (same Manifest as gateway, vault target required)
vaultsdk.Serve(vaultsdk.Config{Handler: mux})

// Identity client (login still belongs on the gateway host before /p/)
c := vaultsdk.Client{BaseURL: vaultURL, GatewayID: id}
ok, user, err := c.PasswordLogin(ctx, user, pass)
pkg, err := c.RedeemEnroll(ctx, token)
```

Gateway plugins use `bytedesk-remote-gateway-plugin-sdk` (same inherited
Manifest, `GATEWAY_PLUGIN_*`). A plugin that lists both targets can import
either SDK for serve/validate; pack once.

See gateway ADR 0014.
