# vaultd
Hashicorp Vault Recursive Destroyer / Deleter

## How to use
Export Vault variables to the environment
```bash
export VAULT_ADDR="https://vault.example.com"       # Change it
export VAULT_TOKEN="s.uI5SCK9ro2FQ1N8Rn4pdbMsr"     # Change it
```
The full list of variables you can add
```bash
VAULT_ADDR
VAULT_RATE_LIMIT
VAULT_AGENT_ADDR

# Deprecated values
VAULT_CACERT
VAULT_CAPATH
VAULT_CLIENT_CERT
VAULT_CLIENT_KEY
VAULT_CLIENT_TIMEOUT
VAULT_SKIP_VERIFY
VAULT_NAMESPACE
VAULT_TLS_SERVER_NAME
VAULT_WRAP_TTL
VAULT_MAX_RETRIES
VAULT_TOKEN
VAULT_MFA
```

## Command Line Tool Usage
#### Simple usage:
```bash
./vaultd -path /secret/database/users
```


## License
Vaultd is provided under the [MIT License](https://github.com/maxkulish/vaultd/blob/master/LICENSE)
