# Sentry plugin for vault

Sentry plugin for Vault manages Sentry projects and their DSN in an organization.

## Setup

The setup guide assumes some familiarity with Vault and its plugin ecosystem.
You must have a Vault server already running, unsealed, and authenticated.

1. Download and decompress the latest plugin binary from the Releases tab on GitHub. Alternatively you can also compile the plugin from source.
1. Move the compiled plugin into Vault's `plugin_directory`:

  ```sh
  $ mv vault-secret-plugin-sentry /etc/vault/plugins/vault-secret-plugin-sentry
  ```

1. Calculate the SHA256 of the plugin and register it in Vault's plugin catalog.
If you are downloading the pre-compiled binary it is highly recommended that you use the published checksums to verify integrity.

  ```sh
  $ export SHA256=$(shasum -a 256 "/etc/vault/plugins/vault-secret-plugin-sentry" | cut -d' ' -f1)
  $ vault plugin register -sha256=${SHA256} -command=vault-secret-plugin-sentry secret secret-sentry
  ```

1. Mount the auth method:

  ```sh
  $ vault secrets enable -path=sentry secret-sentry
  ```

## API

For details on API endpoints and their usage see [docs](./docs).

## License

This code is licensed under the MPLv2 license.