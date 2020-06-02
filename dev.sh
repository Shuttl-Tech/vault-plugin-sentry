#!/usr/bin/env bash
set -e

DIR="$(cd "$(dirname "$(readlink "$0")")" && pwd)"

echo "==> Starting dev"
echo "    --> Creating Scratch dir"

SCRATCH="$DIR/tmp"
mkdir -p "$SCRATCH/plugins"

echo "    --> Writing Vault server config"

tee "$SCRATCH/vault.hcl" > /dev/null <<EOF
plugin_directory = "$SCRATCH/plugins"
EOF

echo "    --> Configuring Shell Environment"
export VAULT_DEV_ROOT_TOKEN_ID="root"
export VAULT_ADDR="http://127.0.0.1:8200"

echo "    --> Starting Vault"
vault server -dev -log-level="debug" -config="$SCRATCH/vault.hcl" > "$SCRATCH/vault.log" 2>&1 &
sleep 3
VAULT_PID=$!

echo "    --> Building test sentry server"
make localserver

echo "    --> Starting test sentry server"
random_local_port=$(seq 25000 30000 | shuf | head -n 1)
./internal/testing/pkg/localserver "${random_local_port}" > "$SCRATCH/localserver.log" 2>&1 &
LOCALSERVER_PID=$!
echo "      --> Server is running on 127.0.0.1:${random_local_port}"

function cleanup {
  echo ""
  echo "  ==> Cleaning up"
  kill -INT "$VAULT_PID" "$LOCALSERVER_PID"
  rm -rf "$SCRATCH"
}
trap cleanup EXIT

echo "    --> Authenticating with vault"
vault login root &>/dev/null

cp "pkg/vault-plugin-sentry_$(go env GOOS)_$(go env GOARCH)" "$SCRATCH/plugins/vault-secrets-sentry"
SHASUM=$(shasum -a 256 "$SCRATCH/plugins/vault-secrets-sentry" | cut -d " " -f1)

echo "    --> Registering plugin"
vault plugin register -sha256="$SHASUM" -command="vault-secrets-sentry" secret sentry | awk '{print "        " $0}'

echo "    --> Mounting plugin"
vault secrets enable -path=sentry sentry | awk '{print "        " $0}'

echo "    --> Reading out"
vault read sentry/info | awk '{print "        " $0}'

echo ""
echo "    --> Vault is available:"
awk '/Unseal Key:|Root Token:/ { print "        " $0 }' "$SCRATCH/vault.log"
echo ""
echo "    --> See vault logs in $SCRATCH/vault.log"
echo "    --> See localserver logs in $SCRATCH/localserver.log"
echo ""
echo "    ==> Ready!"

# Only hold control if not being sourced
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    wait $!
fi