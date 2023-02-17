#!/usr/bin/env bash

# ------------------------------------------------------------
#   MIT License
#
#   Copyright (c) Microsoft Corporation.
#
#   Permission is hereby granted, free of charge, to any person obtaining a copy
#   of this software and associated documentation files (the "Software"), to deal
#   in the Software without restriction, including without limitation the rights
#   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
#   copies of the Software, and to permit persons to whom the Software is
#   furnished to do so, subject to the following conditions:
#
#   The above copyright notice and this permission notice shall be included in all
#   copies or substantial portions of the Software.
#
#   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
#   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
#   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
#   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
#   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
#   SOFTWARE
# ------------------------------------------------------------

# Symphony CLI location
SYMPHONY_INSTALL_DIR="/usr/local/bin"
SYMPHONY_HOME_DIR="${HOME}/.symphony"
USE_SUDO="false"
SYMPHONY_HTTP_REQUEST_CLI=curl
SYMPHONY_GITHUB_ORG=Haishi2016
SYMPHONY_GITHUB_REPO=Vault818
SYMPHONY_CLI_FILENAME=maestro
SYMPHONY_CLI_FILE="${SYMPHONY_INSTALL_DIR}/${SYMPHONY_CLI_FILENAME}"
getSystemInfo() {
    ARCH=$(uname -m)
    case $ARCH in
        armv7*) ARCH="arm";;
        aarch64) ARCH="arm64";;
        x86_64) ARCH="amd64";;
    esac

    OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

    if [[ $OS == "linux" || $OS == "darwin" ]] && [ "$SYMPHONY_INSTALL_DIR" == "/usr/local/bin" ]; then
        USE_SUDO="true"
    fi
}

onError() {
    result=$?
    if [ "$result" != "0" ]; then
        echo "Failed to install Maestro CLI"        
    fi
    cleanUp
    exit $result
}

cleanUp() {
    if [[ -d "${SYMPHONY_TMP_ROOT:-}" ]]; then
        rm -rf "$SYMPHONY_TMP_ROOT"
    fi
}

checkDownloader() {
    if type "curl" > /dev/null; then
        SYMPHONY_HTTP_REQUEST_CLI=curl
    elif type "wget" > /dev/null; then
        SYMPHONY_HTTP_REQUEST_CLI=wget
    else
        echo "Either curl or wget is required to download artifacts"
        exit 1
    fi
}

getLatestRelease() {
    local symphonyReleaseUrl="https://api.github.com/repos/${SYMPHONY_GITHUB_ORG}/${SYMPHONY_GITHUB_REPO}/releases"
    local latestRelease=""

    if [ "$SYMPHONY_HTTP_REQUEST_CLI" == "curl" ]; then
        latestRelease=$(curl -s $symphonyReleaseUrl -H "Accept: application/json"  | grep \"tag_name\" | grep -v rc | awk 'NR==1{print $2}' |  sed -n 's/\"\(.*\)\",/\1/p')
    else
        latestRelease=$(wget -q --header="Accept: application/json" -O - $symphonyReleaseUrl | grep \"tag_name\" | grep -v rc | awk 'NR==1{print $2}' |  sed -n 's/\"\(.*\)\",/\1/p')
    fi

    retVal=$latestRelease
}

verifySupported() {
    releaseTag=$1
    local supported=(darwin-amd64 linux-amd64 linux-arm linux-arm64)
    local current_osarch="${OS}-${ARCH}"

    for osarch in "${supported[@]}"; do
        if [ "$osarch" == "$current_osarch" ]; then
            echo "Your system is ${OS}_${ARCH}"
            return
        fi
    done

    if [ "$current_osarch" == "darwin-arm64" ]; then
        if isReleaseAvailable $releaseTag; then
            return
        else
            echo "The darwin_arm64 arch has no native binary for this version of Symphony, however you can use the amd64 version so long as you have rosetta installed"
            echo "Use 'softwareupdate --install-rosetta' to install rosetta if you don't already have it"
            ARCH="amd64"
            return
        fi
    fi

    echo "No prebuilt binary for ${current_osarch}"
    exit 1
}

checkExistingMaestro() {
    if [ -f "$SYMPHONY_CLI_FILE" ]; then
        echo -e "\nMaestro is detected:"
        $SYMPHONY_CLI_FILE version
        echo -e "Reinstalling Maestro - ${SYMPHONY_CLI_FILE}...\n"
    else
        echo -e "Installing Maestro...\n"
    fi
}

downloadFile() {
    LATEST_RELEASE_TAG=$1

    SYMPHONY_CLI_ARTIFACT="${SYMPHONY_CLI_FILENAME}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_BASE="https://github.com/${SYMPHONY_GITHUB_ORG}/${SYMPHONY_GITHUB_REPO}/releases/download"
    DOWNLOAD_URL="${DOWNLOAD_BASE}/${LATEST_RELEASE_TAG}/${SYMPHONY_CLI_ARTIFACT}"

    # Create the temp directory
    SYMPHONY_TMP_ROOT=$(mktemp -dt symphony-install-XXXXXX)
    ARTIFACT_TMP_FILE="$SYMPHONY_TMP_ROOT/$SYMPHONY_CLI_ARTIFACT"

    echo "Downloading $DOWNLOAD_URL ..."
    if [ "$SYMPHONY_HTTP_REQUEST_CLI" == "curl" ]; then
        curl -SsL "$DOWNLOAD_URL" -o "$ARTIFACT_TMP_FILE"
    else
        wget -q -O "$ARTIFACT_TMP_FILE" "$DOWNLOAD_URL"
    fi

    if [ ! -f "$ARTIFACT_TMP_FILE" ]; then
        echo "failed to download $DOWNLOAD_URL ..."
        exit 1
    fi
}

runAsRoot() {
    local CMD="$*"

    if [ $EUID -ne 0 -a $USE_SUDO = "true" ]; then
        CMD="sudo $CMD"
    fi

    $CMD || {
        echo "Please visit TBD for instructions on how to install without sudo."
        exit 1
    }
}

installFile() {
    tar xf "$ARTIFACT_TMP_FILE" -C "$SYMPHONY_TMP_ROOT"
    local tmp_root_symphony_cli="$SYMPHONY_TMP_ROOT/$SYMPHONY_CLI_FILENAME"

    if [ ! -f "$tmp_root_symphony_cli" ]; then
        echo "Failed to unpack Symphony executable."
        exit 1
    fi

    if [ -f "$SYMPHONY_CLI_FILE" ]; then
        runAsRoot rm "$SYMPHONY_CLI_FILE"
    fi
    chmod o+x $tmp_root_symphony_cli
    runAsRoot cp "$tmp_root_symphony_cli" "$SYMPHONY_INSTALL_DIR"

    if [ -f "$SYMPHONY_CLI_FILE" ]; then
        echo "$SYMPHONY_CLI_FILENAME installed into $SYMPHONY_INSTALL_DIR successfully."

        $SYMPHONY_CLI_FILE --help
    else 
        echo "Failed to install $SYMPHONY_CLI_FILENAME"
        exit 1
    fi

    mkdir -p $SYMPHONY_HOME_DIR

    mv $SYMPHONY_TMP_ROOT/* $SYMPHONY_HOME_DIR/    
}

cleanUp() {
    if [[ -d "${SYMPHONY_TMP_ROOT:-}" ]]; then
        rm -rf "$SYMPHONY_TMP_ROOT"
    fi
}

welcomeMessage() {
    echo -e "\nThank you for installing Maestro, the Symphony CLI. To configure Symphony on your K8s cluster, use: maestro up. If you want to run Symphony in standalone mode without Kubernetes, use: maestro up --no-k8s."
}


# ------------------------------------------------------------
# main
# ------------------------------------------------------------

trap "onError" EXIT

getSystemInfo
checkDownloader

if [ -z "$1" ]; then
    echo "Getting the latest Maestro CLI..."
    getLatestRelease
else
    retVal=v$1
fi

verifySupported $retVal
checkExistingMaestro
downloadFile $retVal
installFile
cleanUp
welcomeMessage