# Build and releasing Maestro CLI

_(last edit: 10/24/2023)_

## Manually create a new CLI release

1. If you are doing this from WSL/Ubuntu, install `zip`.

   ```bash
   sudo apt-get install zip unzip -y
   ```

1. Build symphony-api and maestro for Windows, Mac and Linux.

   ```bash
   # under symphony/api folder
   go build -o symphony-api
   GOOS=windows GOARCH=amd64 go build -o symphony-api.exe
   GOOS=darwin GOARCH=amd64 go build -o symphony-api-mac
   ```

1. Update maestro version.

   Update the `SymphonyAPIVersion` constant under `cli/cmd/up.go` to reflect the latest Symphony version.

1. Build maestro

   ```bash
   # under symphony/cli folder
   go build -o maestro
   GOOS=windows GOARCH=amd64 go build -o maestro.exe
   GOOS=darwin GOARCH=amd64 go build -o maestro-mac
   ```

1. In a separate working folder:

   ```bash
   # This example assumes you are under ~/assemble folder, and you've checked out
   # Symphony repos to the ~/projects/go/src/github.com/azure folder
   # To support the original Percept OSS project, you also need to clone the PerceptOSS repo

   # remove previous packages, if any
   rm maestro_linux_amd64.tar.gz
   rm maestro_windows_amd64.tar.gz
   rm maestro_darwin_amd64.tar.gz

   # copy new binary files, configuration files and scripts
   cp ../projects/go/src/github.com/azure/symphony/api/symphony-api .
   cp ../projects/go/src/github.com/azure/symphony/api/symphony-api.exe .
   cp ../projects/go/src/github.com/azure/symphony/api/symphony-api-mac .
   cp ../projects/go/src/github.com/azure/symphony/api/symphony-api-no-k8s.json .
   cp ../projects/go/src/github.com/azure/symphony/cli/maestro .
   cp ../projects/go/src/github.com/azure/symphony/cli/maestro.exe .
   cp ../projects/go/src/github.com/azure/symphony/cli/maestro-mac .
   
   # Copy over samples
   cp ../projects/go/src/github.com/azure/symphony/docs/samples/samples.json .
   mkdir -p ./k8s
   mkdir -p ./iot-edge
   cp -r ../projects/go/src/github.com/azure/symphony/docs/samples/k8s/hello-world/ ./k8s/
   cp -r ../projects/go/src/github.com/azure/symphony/docs/samples/k8s/staged/ ./k8s/
   cp -r ../projects/go/src/github.com/azure/symphony/docs/samples/iot-edge/simulated-temperature-sensor/ ./iot-edge/

   # package Linux
   tar -czvf maestro_linux_amd64.tar.gz maestro symphony-api symphony-api-no-k8s.json samples.json k8s iot-edge
   # package Windows
   zip -r maestro_windows_amd64.zip maestro.exe symphony-api.exe symphony-api-no-k8s.json samples.json k8s iot-edge
   # package Mac
   rm maestro
   rm symphony-api
   mv maestro-mac maestro
   mv symphony-api-mac symphony-api
   tar -czvf maestro_darwin_amd64.tar.gz maestro symphony-api symphony-api-no-k8s.json samples.json k8s iot-edge
   ```

1. Edit your release to include the three **.gz** files from the previous step:

   ![CLI release](../images/cli-release.png)

1. Check `symphony/cli/install/install.sh` and `symphony/cli/install/install.ps1` to a public repository. Users will be instructed to use scripts from this repo for the one-command experience, for example:

   ```bash
   wget -q https://raw.githubusercontent.com/Haishi2016/Vault818/master/cli/install/install.sh -O - | /bin/bash
   ```

## Appendix

* Build for x64 (on an x64 machine)

  ```bash
  go build -o maestro-x64-<version>
  ```

* Build for ARM64

  ```bash
  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o maestro-arm64-<version>
  ```

* Build for Windows AMD64

  ```bash
  GOOS=windows GOARCH=amd64 go build -o maestro.exe
  ```

* Build for Mac

  ```bash
  GOOS=darwin GOARCH=amd64 go build -o maestro-mac
  ```
