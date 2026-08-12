# bytedesk-vault-sdk

Vault **v1** HTTP client for gateway plugins (and other ByteDesk clients).
Login still belongs on the **gateway host** (auth before `/p/`). This module
is for enroll, discovery, and password begin/finish from a plugin or agent.

Depends on `bytedesk-sdk-dependencies`. See gateway ADR 0012 / 0014.

```go
c := vaultsdk.Client{BaseURL: vaultURL, GatewayID: id}
ok, user, err := c.PasswordLogin(ctx, user, pass)
pkg, err := c.RedeemEnroll(ctx, token)
```
