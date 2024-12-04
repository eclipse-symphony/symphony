# Symphony Remote Agent

The Symphony Remote Agent is currently in a draft state. It is built for the validation purposes of Symphony remote agent bootstrap and Symphony cluster asynchronous operations.

## Components

The agent consists of the following components:
- **Start-up**: `main.go`
- **Binding**: `http.go`
- **Manager**: `agent.go`
- **Execution**: `providers/target/script_provider.go`

## Script Provider

The Script Provider is designed to be the most applicable provider for the remote agent. Below is the configuration for the Script Provider:

- `ApplyScript`: `"mock-apply.sh"`
- `GetScript`: `"mock-get.sh"`
- `RemoveScript`: `"mock-remove.sh"`
- `ScriptFolder`: `./script`
- `StagingFolder`: `"."`

### Known Issues

The Script Provider's apply functionality currently has a bug. As a workaround, the script output is added to the `componentSpec` message.

## Mock Implementation

The remote agent currently uses a mock to read requests from `samples/request.json` and prints the response body to the console. It should poll requests from the Symphony endpoint and patch the asynchronous operation result to the Symphony patch endpoint. This configuration is set in `config.json`.

## Run draft agent
```
go run main.go -config=./config.json -client-cert=./bootstrap/public.pem -client-key=./bootstrap/private.pem -target-name=$target_name -namespace=$namespace -topology=./bootstrap/topologies.json
```