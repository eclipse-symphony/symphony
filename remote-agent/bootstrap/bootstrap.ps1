[CmdletBinding()]
param (
    [string]$endpoint,
    [string]$cert_path,
    [SecureString]$cert_password,
    [string]$target_name,
    [string]$namespace = "default",
    [string]$topology
)

function usage {
    Write-Host "Usage: .\script.ps1 -endpoint <endpoint> -cert_path <cert_path>  -target_name <target_name> -namespace <namespace> -topology <topology> -config <config>"
    exit 1
}

# Check if the correct number of parameters are provided
$requiredParams = @('endpoint', 'cert_path', 'cert_password', 'target_name', 'namespace', 'topology')
$providedParams = $PSBoundParameters.Keys | Where-Object { $_ -in $requiredParams }
Write-Verbose "Debug: Number of required parameters provided: $($providedParams.Count)"
if ($providedParams.Count -ne 6) {
    Write-Host "Error: Invalid number of parameters." -ForegroundColor Red
    usage
}

# Validate the endpoint (basic URL validation)
Write-Verbose "Debug: Endpoint: $endpoint"
if ($endpoint -notmatch "^https?://") {
    Write-Host "Error: Invalid endpoint. Must be a valid URL starting with http:// or https://" -ForegroundColor Red
    usage
}

# Validate the certificate path (check if the file exists)
Write-Verbose "Debug: Cert Path: $cert_path"
if (-not (Test-Path $cert_path)) {
    Write-Host "Error: Certificate file not found at path: $cert_path" -ForegroundColor Red
    usage
} elseif ($cert_path -notlike "*.pfx") {
        Write-Host "Error: The certificate file must be a .pfx file." -ForegroundColor Red
        usage
}    

# Validate the certificate password (check if the file exists)
Write-Verbose "Debug: Cert Path: $cert_password"
if ([string]::IsNullOrEmpty($cert_password)) {
    Write-Host "Error: Certificate password must be a non-empty string." -ForegroundColor Red
    usage
}

# Validate the target name (non-empty string)
Write-Verbose "Debug: Target Name: $target_name"
if ([string]::IsNullOrEmpty($target_name)) {
    Write-Host "Error: Target name must be a non-empty string." -ForegroundColor Red
    usage
}

# Validate the namespace (non-empty string)
Write-Verbose "Debug: Namespace: $namespace"
if ([string]::IsNullOrEmpty($namespace)) {
    Write-Host "Error: Namespace must be a non-empty string." -ForegroundColor Red
    $namespace = "default"
}

# Validate the topology file (non-empty string)
Write-Verbose "Debug: Topology: $topology"
if (-not (Test-Path $topology)) {
    Write-Host "Error: topology file not found at path: $topology" -ForegroundColor Red
    usage
} elseif ($topology -notlike "*.json") {
        Write-Host "Error: The topology file must be a .json file." -ForegroundColor Red
        usage
}    

Import-Module PKI
# Create the JSON configuration
$configJson = @{
    requestEndpoint = "$endpoint/solution/tasks"
    responseEndpoint = "$endpoint/solution/task/getResult"
    baseUrl = "$endpoint/"
} | ConvertTo-Json

# Save the JSON configuration to a file
$configFile = "config.json"
$configJson | Set-Content -Path $configFile
# Convert cert_path, topology_path, config to absolute paths
$cert_path = Resolve-Path $cert_path
$topology = Resolve-Path $topology
$config = Resolve-Path $configFile
Write-Host "Successfully create config file" -ForegroundColor Yellow
# for pfx verify
Write-Host "Start to import pfx certificate" -ForegroundColor Blue
try{
    $cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2
    $cert.Import($cert_path, $cert_password, "Exportable, PersistKeySet")
    Write-Host "Successfully import pfx certificate" -ForegroundColor Blue
}
catch {
    Write-Host "Error: The certificate file is not a valid PFX file or the password is incorrect."  -ForegroundColor Red
    exit 1
}
Write-Host "Start to get working cert from symphony server" -ForegroundColor Blue
try {
    $WebRequestParams = @{
        Uri = "$($endpoint)/targets/bootstrap/$($target_name)?namespace=$($namespace)&osPlatform=windows"
        Method = 'Post'
        Certificate = $cert
    }
    $response = Invoke-WebRequest @WebRequestParams -ErrorAction Stop
    Write-Host "Successfully get working cert from symphony server" -ForegroundColor Yellow
} catch {
    Write-Host "Error: Failed to send request to endpoint."  -ForegroundColor Red
    exit 1
}
$jsonResponse = $response.Content | ConvertFrom-Json
# Parse JSON response and extract public field
$public = $jsonResponse.public
# Extract header and footer of the public field
$header = ($public -split ' ')[0..1] -join ' '
$footer = ($public -split ' ')[-3..-1] -join ' '
# Extract Base64 encoded content and replace spaces with newlines
$base64_content = ($public -split ' ')[2..(($public -split ' ').Length - 4)] -join "`n"
# Combine header, Base64 content, and footer
$corrected_public_content = "$header`n$base64_content`n$footer"
# Write corrected_public_content to public.pem
$corrected_public_content | Set-Content -Path "public.pem" -Encoding ascii
Write-Host "Successfully create public.pem file" -ForegroundColor Yellow
# Extract private field
$private = $response.Content | ConvertFrom-Json | Select-Object -ExpandProperty private
# Extract header and footer of the private field
$header = ($private -split ' ')[0..3] -join ' '
$footer = ($private -split ' ')[-5..-1] -join ' '
# Extract Base64 content and replace spaces with newlines
$base64_content = ($private -split ' ')[4..(($private -split ' ').Length - 6)] -join "`n"
# Combine header, Base64 content, and footer
$corrected_private_content = "$header`n$base64_content`n$footer"

# Write corrected_private_content to private.pem
$corrected_private_content |  Set-Content -Path "private.pem" -Encoding ascii
Write-Host "Successfully create private.pem file" -ForegroundColor Yellow
# Download remote-agent binary file
Write-Host "Begin to download remote-agent binary file" -ForegroundColor Blue

try {
    $WebRequestParams = @{
        Uri = "$($endpoint)/files/remote-agent.exe"
        Method = 'Get'
        Certificate = $cert
    }
    $result = Invoke-WebRequest @WebRequestParams -OutFile "remote-agent.exe" -ErrorAction Stop
    Write-Host $result.Content
} catch {
    Write-Host "Error: Failed to download." -ForegroundColor Red
    Write-Host "Error Message: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Error Details: $($_ | Out-String)" -ForegroundColor Red
    exit 1
}

Write-Host "Begin to start remote agent process" -ForegroundColor Blue

# Convert public.pem, private.pem, remote-agent to absolute paths
$public_path = Resolve-Path "./public.pem"
$private_path = Resolve-Path "./private.pem"
$agent_path = Resolve-Path "./remote-agent.exe"
$serviceName = "symphony-service"
$serviceDescription = "Remote Agent Service"
# Create remote agent process
$processArgs = "-config=$config -client-cert=$public_path -client-key=$private_path -target-name=$target_name -namespace=$namespace -topology=$topology"
Write-Host "Process Args: $processArgs"
Start-Process -FilePath $agent_path -ArgumentList $processArgs
Write-Host "Successfully start remote agent process" -ForegroundColor Yellow
