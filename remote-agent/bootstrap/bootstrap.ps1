[CmdletBinding()]
param (
    [string]$endpoint,
    [string]$cert_path,
    [string]$target_name,
    [string]$namespace = "default",
    [string]$topology,
    [ValidateSet('service','schedule')][string]$run_mode = 'schedule'
)

if (-not $cert_password) {
    $cert_password = Read-Host "Please enter the certificate password (input will be hidden)" -AsSecureString
}
$cert_password_plain = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto(
    [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($cert_password)
)

[System.Environment]::SetEnvironmentVariable("DOTNET_SYSTEM_NET_HTTP_SOCKETSHTTPHANDLER_LOGFILE", "$pwd\http.log", "Process")
[System.Environment]::SetEnvironmentVariable("DOTNET_SYSTEM_NET_HTTP_SOCKETSHTTPHANDLER_LOGLEVEL", "Debug", "Process")
function usage {
    Write-Host "Usage: .\script.ps1 -endpoint <endpoint> -cert_path <cert_path>  -target_name <target_name> -namespace <namespace> -topology <topology> -config <config>"
    exit 1
}

# Check if the correct number of parameters are provided
$requiredParams = @('endpoint', 'cert_path', 'target_name', 'namespace', 'topology', 'run_mode')
$providedParams = $PSBoundParameters.Keys | Where-Object { $_ -in $requiredParams }
Write-Verbose "Debug: Number of required parameters provided: $($providedParams.Count)"
if ($providedParams.Count -lt 5) {
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
$DebugPreference = "Continue"
$VerbosePreference = "Continue"
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12
[System.Net.ServicePointManager]::ServerCertificateValidationCallback = { $true }
# [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls, [Net.SecurityProtocolType]::Tls11, [Net.SecurityProtocolType]::Tls12, [Net.SecurityProtocolType]::Ssl3

Write-Host "Successfully create config file" -ForegroundColor Yellow
# for pfx verify
Write-Host "Start to import pfx certificate" -ForegroundColor Blue
try{
    $flags = $flags -bor [System.Security.Cryptography.X509Certificates.X509KeyStorageFlags]::PersistKeySet 
    $cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2(
        $cert_path, $cert_password_plain, $flags
    )
    Write-Host "Successfully import pfx certificate" -ForegroundColor Blue
    Write-Host "Cert Subject: $($cert.Subject)"
    Write-Host "Cert Thumbprint: $($cert.Thumbprint)"
    Write-Host "Has Private Key: $($cert.HasPrivateKey)"
}
catch {
    Write-Host "Exception Message: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Exception Type: $($_.Exception.GetType().FullName)" -ForegroundColor Red
    Write-Host "Stack Trace: $($_.Exception.StackTrace)" -ForegroundColor Red
    if ($_.Exception.InnerException) {
        Write-Host "Inner Exception: $($_.Exception.InnerException.Message)" -ForegroundColor Red
        Write-Host "Inner Stack Trace: $($_.Exception.InnerException.StackTrace)" -ForegroundColor Red
    }
    Write-Host "Error: The certificate file is not a valid PFX file, or the password is incorrect, or the file is corrupt."  -ForegroundColor Red
    exit 1
}

try {
    $WebRequestParams = @{
        Uri = "$($endpoint)/targets/bootstrap/$($target_name)?namespace=$($namespace)&osPlatform=windows"
        Method = 'Post'
        Certificate = $cert  
        Headers = @{ "Content-Type" = "application/json"; "User-Agent" = "PowerShell-Debug" }
        Body = (Get-Content $topology -Raw)
    }
    Write-Host "WebRequestParams:"
    $WebRequestParams.GetEnumerator() | ForEach-Object { Write-Host ("  {0}: {1}" -f $_.Key, $_.Value) }
    $response = Invoke-WebRequest @WebRequestParams -Verbose
    Write-Host "Successfully get working cert from symphony server" -ForegroundColor Yellow
} catch {
    Write-Host "Error: Failed to send request to endpoint."  -ForegroundColor Red
    Write-Host "Error Message: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Error Details: $($_ | Out-String)" -ForegroundColor Red
    
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
# Ensure remote-agent.exe can be overwritten: stop service and kill process
Stop-Service -Name symphony-service -Force -ErrorAction SilentlyContinue
sc.exe delete symphony-service | Out-Null
Start-Sleep -Seconds 2
Get-Process remote-agent -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Seconds 1
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

# Set the paths to the public and private keys, agent binary, and topology file
$public_path = Resolve-Path "./public.pem"
$private_path = Resolve-Path "./private.pem"
$agent_path = Resolve-Path "./remote-agent.exe"
$config = Resolve-Path "./config.json"
$topology = Resolve-Path $topology
$serviceName = "symphony-service"
$serviceDescription = "Remote Agent Service"

#  compose the command line arguments for the remote agent
$processArgs = "-config=`"$config`" -client-cert=`"$public_path`" -client-key=`"$private_path`" -target-name=`"$target_name`" -namespace=`"$namespace`" -topology=`"$topology`""
$binPath = "`"$agent_path`" $processArgs"

if ($run_mode -eq 'service') {
    Write-Host "[run_mode=service] Register and start as Windows service..." -ForegroundColor Cyan
    # check if the service already exists
    if (Get-Service -Name $serviceName -ErrorAction SilentlyContinue) {
        Write-Host "Service $serviceName already exists. Stopping and removing..." -ForegroundColor Yellow
        Stop-Service -Name $serviceName -Force -ErrorAction SilentlyContinue
        sc.exe delete $serviceName | Out-Null
        Write-Host "Service $serviceName deleted." -ForegroundColor Yellow
    } else {
        Write-Host "Service $serviceName does not exist, nothing to delete." -ForegroundColor Green
    }

    # register the service
    Write-Host "Registering $serviceName as a Windows service..." -ForegroundColor Blue
    New-Service -Name $serviceName -BinaryPathName $binPath -Description $serviceDescription -DisplayName $serviceName -StartupType Automatic
    # Start the service
    try {
        Start-Service -Name $serviceName
        Write-Host "Successfully registered and started $serviceName as a Windows service" -ForegroundColor Yellow
    } catch {
        Write-Host "Failed to start $serviceName. Please check the service logs and Windows Event Viewer." -ForegroundColor Red
        throw
    }
} else {
    Write-Host "[run_mode=schedule] Register and start as scheduled task..." -ForegroundColor Cyan
    $watchdogProcessName = "remote-agent"
    $watchdogLogPath = Join-Path (Split-Path $agent_path) "watchdog.log"
    $watchdogExePath = $agent_path
    $watchdogArgs = '-config="' + $config + '" -client-cert="' + $public_path + '" -client-key="' + $private_path + '" -target-name="' + $target_name + '" -namespace="' + $namespace + '" -topology="' + $topology + '"'
    $watchdogScript = @"
try {
    if (-not (Get-Process -Name $watchdogProcessName -ErrorAction SilentlyContinue)) {
        Start-Process -FilePath '$watchdogExePath' -ArgumentList '$watchdogArgs' -WindowStyle Hidden
    }
} catch {
    Write-Host "Error in watchdog: $_.Exception.Message" -ForegroundColor Red
}
"@

    Write-Host "script content for watchdog:"
    Write-Host $watchdogScript
    $watchdogPath = Join-Path (Split-Path $agent_path) "watchdog-remote-agent.ps1"
    $watchdogScript | Set-Content -Path $watchdogPath -Encoding UTF8

    # trigger every 1 minutes, start 1 minute later, repeat every 1 minute, for 10 years
    $Trigger = New-ScheduledTaskTrigger -Once -At (Get-Date).AddMinutes(1) -RepetitionInterval (New-TimeSpan -Minutes 1) -RepetitionDuration  (New-TimeSpan -Days 3650) 

    # task action to run the watchdog script
    $Action = New-ScheduledTaskAction -Execute "pwsh.exe" -Argument "-NoProfile -WindowStyle Hidden -File `"$watchdogPath`""

    $TaskName = "RemoteAgentTask"
    if (Get-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue) {
        Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false
    }
    Register-ScheduledTask -TaskName $TaskName -Action $Action -Trigger $Trigger -Description "Guard remote-agent.exe, auto-restart, unlimited retries" -User "redmond\jiaxinyan" -RunLevel Highest
    Start-ScheduledTask -TaskName $TaskName
    Write-Host "Already registered and started scheduled task $TaskName, set to run at startup and auto-retry on failure." -ForegroundColor Yellow
}
