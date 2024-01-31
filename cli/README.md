<!--
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
SPDX-License-Identifier: MIT
-->
# Build up CLI and release

_(last edit: 11/14/2023)_

To build up Maestro CLI and release with Symphony api and samples. 
The build uses mage to automate the process. Please refer to  [Build CLI](../docs/symphony-book/cli/build_cli.md) if manual steps are preferrd.

## Prerequisites
1. If you are doing this from WSL/Ubuntu, install `zip`.

   ```bash
   sudo apt-get install zip unzip -y
   ```
2. Update maestro version.

   Update the `SymphonyAPIVersion` constant under `cli/cmd/up.go` to reflect the latest Symphony version.

## Mage Commands
See all commands with mage -l
```
# under symphony/cli folder
> mage -l
Use this tool to quickly build symphony api or maestro cli. It can also help generate the release package.

Targets:
  buildApi            Build Symphony api for Windoes, Mac and Linux.
  buildCli            Build maestro cli tools for Windoes, Mac and Linux.
  generatePackages    Generate packages with Symphony api, maestro cli and samples for Windoes, Mac and Linux.
```
Samples
```
# under symphony/cli folder
# Build up Symphony api for windows, mac nad linux. You can find the binary files in symphony/api folder
mage buildApi

# Build up Maestro CLI for windows, mac nad linux. You can find the binary files in symphony/cli folder
mage buildCli

# Build up Symphony api and Maestro CLI for windows, mac nad linux. Copy the binary files, samples and configuration files to specified folder and generate the release package. You will find maestro_windows_amd64.zip, maestro_darwin_amd64.tar.gz and maestro_linux_amd64.tar.gz in the specified folder.
mage generatePackages /home/usr/assemble
```

## Release 

1. Edit your release to include the three **.gz** files from the previous step:

   ![CLI release](../docs/symphony-book/images/cli-release.png)

1. Check `symphony/cli/install/install.sh` and `symphony/cli/install/install.ps1` to a public repository. Users will be instructed to use scripts from this repo for the one-command experience, for example:

   ```bash
   wget -q https://raw.githubusercontent.com/eclipse-symphony/symphony/master/cli/install/install.sh -O - | /bin/bash
   ```