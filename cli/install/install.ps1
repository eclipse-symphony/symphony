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

param (
    [string]$Version,
    [string]$SymphonyRoot = "$Env:UserProfile\.symphony",
    [string]$SymphonyReleaseUrl = "",
    [scriptblock]$CustomAssetFactory = $null
)

Write-Output ""
$ErrorActionPreference = 'stop'

#Escape space of SymphonyRoot path
$SymphonyRoot = $SymphonyRoot -replace ' ', '` '

# Constants
$SymphonyCliFileName = "maestro.exe"
$SymphonyCliFilePath = "${SymphonyRoot}\${SymphonyCliFileName}"

# GitHub Org and repo hosting Dapr CLI
$GitHubOrg = "Haishi2016"
$GitHubRepo = "Vault818"

# Set Github request authentication for basic authentication.
if ($Env:GITHUB_USER) {
    $basicAuth = [System.Convert]::ToBase64String([System.Text.Encoding]::ASCII.GetBytes($Env:GITHUB_USER + ":" + $Env:GITHUB_TOKEN));
    $githubHeader = @{"Authorization" = "Basic $basicAuth" }
}
else {
    $githubHeader = @{}
}

if ((Get-ExecutionPolicy) -gt 'RemoteSigned' -or (Get-ExecutionPolicy) -eq 'ByPass') {
    Write-Output "PowerShell requires an execution policy of 'RemoteSigned'."
    Write-Output "To make this change please run:"
    Write-Output "'Set-ExecutionPolicy RemoteSigned -scope CurrentUser'"
    break
}

# Change security protocol to support TLS 1.2 / 1.1 / 1.0 - old powershell uses TLS 1.0 as a default protocol
[Net.ServicePointManager]::SecurityProtocol = "tls12, tls11, tls"

# Check if Dapr CLI is installed.
if (Test-Path $SymphonyCliFilePath -PathType Leaf) {
    Write-Warning "Maestro is detected - $SymphonyCliFilePath"
    Invoke-Expression "$SymphonyCliFilePath --version"
    Write-Output "Reinstalling Maestro..."
}
else {
    Write-Output "Installing Maestro..."
}

# Create Dapr Directory
Write-Output "Creating $SymphonyRoot directory"
New-Item -ErrorAction Ignore -Path $SymphonyRoot -ItemType "directory"
if (!(Test-Path $SymphonyRoot -PathType Container)) {
    Write-Warning "Please visit https://github.com/azure/symphony-docs/ for instructions on how to install without admin rights."
    throw "Cannot create $SymphonyRoot"
}

# Get the list of release from GitHub
$releaseJsonUrl = $SymphonyReleaseUrl
if (!$releaseJsonUrl) {
    $releaseJsonUrl = "https://api.github.com/repos/${GitHubOrg}/${GitHubRepo}/releases"
}

$releases = Invoke-RestMethod -Headers $githubHeader -Uri $releaseJsonUrl -Method Get
if ($releases.Count -eq 0) {
    throw "No releases from github.com/${GitHubOrg}/${GitHubRepo} repo"
}

# get latest or specified version info from releases
function GetVersionInfo {
    param (
        [string]$Version,
        $Releases
    )
    # Filter windows binary and download archive
    if (!$Version) {
        $release = $Releases | Where-Object { $_.tag_name -notlike "*rc*" } | Select-Object -First 1
    }
    else {
        $release = $Releases | Where-Object { $_.tag_name -eq "v$Version" } | Select-Object -First 1
    }

    return $release
}

# get info about windows asset from release
function GetWindowsAsset {
    param (
        $Release
    )
    if ($CustomAssetFactory) {
        Write-Output "CustomAssetFactory dectected, try to invoke it"
        return $CustomAssetFactory.Invoke($Release)
    }
    else {
        $windowsAsset = $Release | Select-Object -ExpandProperty assets | Where-Object { $_.name -Like "*windows_amd64.zip" }
        if (!$windowsAsset) {
            throw "Cannot find the windows Maestro CLI binary"
        }
        [hashtable]$return = @{}
        $return.url = $windowsAsset.url
        $return.name = $windowsAsset.name
        return $return
    }`
}

$release = GetVersionInfo -Version $Version -Releases $releases
if (!$release) {
    throw "Cannot find the specified Maestro CLI binary version"
}
$asset = GetWindowsAsset -Release $release
$zipFileUrl = $asset.url
$assetName = $asset.name

$zipFilePath = $SymphonyRoot + "\" + $assetName
Write-Output "Downloading $zipFileUrl ..."

$githubHeader.Accept = "application/octet-stream"
$oldProgressPreference = $progressPreference;
$progressPreference = 'SilentlyContinue';
Invoke-WebRequest -Headers $githubHeader -Uri $zipFileUrl -OutFile $zipFilePath
$progressPreference = $oldProgressPreference;
if (!(Test-Path $zipFilePath -PathType Leaf)) {
    throw "Failed to download Maestro Cli binary - $zipFilePath"
}

# Extract Maestro CLI to $SymphonyRoot
Write-Output "Extracting $zipFilePath..."
Microsoft.Powershell.Archive\Expand-Archive -Force -Path $zipFilePath -DestinationPath $SymphonyRoot
if (!(Test-Path $SymphonyCliFilePath -PathType Leaf)) {
    throw "Failed to download Maestro Cli archieve - $zipFilePath"
}

# Check the Maestro CLI version
Invoke-Expression "$SymphonyCliFilePath version"

# Clean up zipfile
Write-Output "Clean up $zipFilePath..."
Remove-Item $zipFilePath -Force

# Add SymphonyRoot directory to User Path environment variable
Write-Output "Try to add $SymphonyRoot to User Path Environment variable..."
$UserPathEnvironmentVar = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPathEnvironmentVar -like '*symphony*') {
    Write-Output "Skipping to add $SymphonyRoot to User Path - $UserPathEnvironmentVar"
}
else {
    [System.Environment]::SetEnvironmentVariable("PATH", $UserPathEnvironmentVar + ";$SymphonyRoot", "User")
    $UserPathEnvironmentVar = [Environment]::GetEnvironmentVariable("PATH", "User")
    Write-Output "Added $SymphonyRoot to User Path - $UserPathEnvironmentVar"
}

Write-Output "`r`nMaestro CLI is installed successfully."
Write-Output "To get started with Maestro, use maestro up to set up Symphony."
Write-Output "Then, browse samples using maestro samples list and run a select sample with maestro samples run <sample name>."