### Running MQTT provider tests locally with Mosquitto (Docker)

This guide shows how to run the unit/integration tests for the MQTT target provider against a local Mosquitto MQTT broker running in Docker. It covers both plain TCP (port 1883) and mutual TLS (mTLS, port 8883).

Prerequisites
- Docker (or Docker Desktop)
- OpenSSL (for generating test certificates if you want to run the mTLS test)
- Go 1.21+ (matching this repo’s go.mod)

Repository layout of interest
- `api/pkg/apis/v1alpha1/providers/target/mqtt/mqtt.go`
- `api/pkg/apis/v1alpha1/providers/target/mqtt/mqtt_test.go`

By default the tests are gated behind environment variables and will be skipped unless explicitly enabled.

Plain TCP broker (1883)
1) Create a Mosquitto config that enables a listener and allows anonymous access (required for Mosquitto 2.x):

Create `mosquitto.conf` with:

```conf
listener 1883
protocol mqtt
allow_anonymous true
```

2) Start Mosquitto with the config mounted:

```bash
docker run --rm -it --name mosquitto -p 1883:1883 -v $(pwd)/mosquitto.conf:/mosquitto/config/mosquitto.conf -v $(pwd)/mtls-certs:/certs eclipse-mosquitto:2
```

3) In a separate shell, enable the tests that expect a locally running broker:

- Linux/macOS (bash/zsh):

```bash
export TEST_MQTT_LOCAL_ENABLED=1
export TEST_MQTT=1
```

- Windows PowerShell:

```powershell
$env:TEST_MQTT_LOCAL_ENABLED = "1"
$env:TEST_MQTT = "1"
```

4) Run the tests for the MQTT provider package:

```bash
cd api
go test ./pkg/apis/v1alpha1/providers/target/mqtt -v
```

Notes
- The tests publish and subscribe on topics `coa-request` and `coa-response` by default.
- The tests create a responder client within the test itself; Mosquitto simply routes messages.
- If Mosquitto logs show "Starting in local only mode... Create a configuration file which defines a listener", you did not mount a config; use the steps above.

mTLS broker (8883)

The file `mqtt_test.go` contains an optional mTLS integration-style test guarded by environment variables. To run it, stand up Mosquitto with TLS and client-certificate auth and point the test to your CA, client cert, and client key.

1) Generate a simple CA, server, and client certificates (for local testing only):

```bash
# Clean slate (optional)
rm -f ca.* server.* client.* *.srl

# 1) CA
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -sha256 -days 365 \
  -subj "/CN=test-ca" -out ca.crt

# 2) Keys & CSR config with extensions
openssl genrsa -out server.key 2048

cat > server.cnf <<'EOF'
[ req ]
distinguished_name = dn
prompt = no
req_extensions = v3_req

[ dn ]
CN = localhost

[ v3_req ]
subjectAltName = @alt_names
extendedKeyUsage = serverAuth

[ alt_names ]
DNS.1 = localhost
IP.1  = 127.0.0.1
EOF

# CSR with req extensions present
openssl req -new -key server.key -out server.csr -config server.cnf

# Sign CSR and COPY THE SAME EXTENSIONS into the cert
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -days 365 -sha256 \
  -extensions v3_req -extfile server.cnf

openssl genrsa -out client.key 2048

cat > client.cnf <<'EOF'
[ req ]
distinguished_name = dn
prompt = no
req_extensions = v3_req

[ dn ]
CN = mtls-client

[ v3_req ]
extendedKeyUsage = clientAuth
EOF

openssl req -new -key client.key -out client.csr -config client.cnf

openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out client.crt -days 365 -sha256 \
  -extensions v3_req -extfile client.cnf

chmod 600 server.key client.key
chmod 644 ca.crt server.crt client.crt

echo "---- SERVER ----"
openssl x509 -in server.crt -noout -text | sed -n '/Subject:/p;/Subject Alternative Name/,+1p;/Extended Key Usage/p'
echo "---- CLIENT ----"
openssl x509 -in client.crt -noout -text | sed -n '/Subject:/p;/Extended Key Usage/p'

```

2) Create a Mosquitto config `mosquitto.conf` next to the certs with both 1883 and 8883 enabled and require client certs on 8883:

```conf
# Plain TCP for anonymous tests
listener 1883
protocol mqtt
allow_anonymous true

# TLS (server-auth only) for TestGet_TLS
listener 8883
protocol mqtt
cafile /certs/ca.crt
certfile /certs/server.crt
keyfile /certs/server.key
allow_anonymous true

# mTLS (client certs required) for TestGet_mTLS
listener 8884
protocol mqtt
cafile /certs/ca.crt
certfile /certs/server.crt
keyfile /certs/server.key
require_certificate true
use_identity_as_username true
allow_anonymous true
```

3) Start Mosquitto with the config and certs mounted:

```bash
# From the folder containing mosquitto.conf and the *.crt/*.key files
docker run --rm -it \
  --name mosquitto \
  -p 1883:1883 \
  -p 8883:8883 \
  -p 8884:8884 \
  -v $(pwd)/mosquitto.conf:/mosquitto/config/mosquitto.conf \
  -v $(pwd)/mtls-certs:/certs \
  eclipse-mosquitto:2
```

4) In a separate shell, export the mTLS test environment variables:

```bash
export TEST_MQTT_TLS=1
export TEST_MQTT_TLS_BROKER="ssl://127.0.0.1:8883"
export TEST_MQTT_TLS_CA="$(pwd)/mtls-certs/ca.crt"
export TEST_MQTT_TLS_REQUEST_TOPIC="coa-request"
export TEST_MQTT_TLS_RESPONSE_TOPIC="coa-response"

export TEST_MQTT_MTLS=1
export TEST_MQTT_MTLS_BROKER="ssl://127.0.0.1:8884"
export TEST_MQTT_MTLS_CA="$(pwd)/mtls-certs/ca.crt"
export TEST_MQTT_MTLS_CERT="$(pwd)/mtls-certs/client.crt"
export TEST_MQTT_MTLS_KEY="$(pwd)/mtls-certs/client.key"
export TEST_MQTT_MTLS_REQUEST_TOPIC="coa-request"
export TEST_MQTT_MTLS_RESPONSE_TOPIC="coa-response"
```

5) Run tests:

```bash
cd api
go test ./pkg/apis/v1alpha1/providers/target/mqtt -v
```

Tips
- If you only want to run the mTLS test, use `-run` to filter:

```bash
go test ./pkg/apis/v1alpha1/providers/target/mqtt -run TestGet_mTLS -v
```

- If you see certificate errors, double-check that:
  - `TEST_MQTT_MTLS_CA` points to the CA that signed both the server and client certs.
  - `server.crt`/`server.key` match, and CN/SAN includes `localhost` or you connect by the same name you issued.
  - Mosquitto is actually listening on 8883 (check container logs).

Environment variables used by tests
- Plain TCP tests (skip unless set): `TEST_MQTT_LOCAL_ENABLED=1`, `TEST_MQTT=1`
- mTLS test (skip unless set): `TEST_MQTT_MTLS=1`, `TEST_MQTT_MTLS_BROKER`, `TEST_MQTT_MTLS_CA`, `TEST_MQTT_MTLS_CERT`, `TEST_MQTT_MTLS_KEY`, `TEST_MQTT_MTLS_REQUEST_TOPIC`, `TEST_MQTT_MTLS_RESPONSE_TOPIC`

Troubleshooting
- On Windows with WSL, run Docker Desktop and expose the ports to the host. The tests connect to `127.0.0.1`.
- If ports are in use, stop other MQTT brokers or change the exposed ports in `docker run` and env vars accordingly.

