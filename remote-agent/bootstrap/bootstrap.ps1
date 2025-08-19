[CmdletBinding()]
param (
    [Parameter(Mandatory=$false)]
    [string]$protocol = "http",
    [Parameter(Mandatory=$false)]
    [string]$endpoint,
    [Parameter(Mandatory=$false)]
    [string]$mqtt_broker,
    [Parameter(Mandatory=$false)]
    [int]$mqtt_port,
    [Parameter(Mandatory=$true)]
    [string]$cert_path,
    [Parameter(Mandatory=$false)]
    [string]$key_path,
    [Parameter(Mandatory=$true)]
    [string]$target_name,
    [Parameter(Mandatory=$false)]
    [string]$namespace = "default",
    [Parameter(Mandatory=$true)]
    [string]$topology,
    [Parameter(Mandatory=$false)]
    [string]$run_mode = "service",
    [Parameter(Mandatory=$false)]
    [string]$ca_cert_path,
    [Parameter(Mandatory=$false)]
    [bool]$use_cert_subject = $false
)
function usage {
    Write-Host "Usage for HTTP mode:" -ForegroundColor Yellow
    Write-Host ".\bootstrap.ps1 -protocol http -endpoint <endpoint> -cert_path <cert_path> -target_name <target_name> -namespace <namespace> -topology <topology> -run_mode <service|schedule>" -ForegroundColor Yellow
    Write-Host "Usage for MQTT mode:" -ForegroundColor Yellow
    Write-Host ".\bootstrap.ps1 -protocol mqtt -mqtt_broker <broker_address> -mqtt_port <broker_port> -cert_path <cert_path> -key_path <key_path> -target_name <target_name> -namespace <namespace> -topology <topology> -run_mode <service|schedule> [-ca_cert_path <ca_cert_path>] [-use_cert_subject <true|false>]" -ForegroundColor Yellow
    exit 1
}

# Protocol-specific validations
if ($protocol -eq "http") {
    # Validate the endpoint (basic URL validation)
    if (-not $endpoint -or $endpoint -notmatch "^https?://") {
        Write-Host "Error: Invalid endpoint. Must be a valid URL starting with http:// or https://" -ForegroundColor Red
        usage
    }
    
    # For HTTP mode, validate PFX certificate
    if (-not (Test-Path $cert_path)) {
        Write-Host "Error: Certificate file not found at path: $cert_path" -ForegroundColor Red
        usage
    } elseif ($cert_path -notlike "*.pfx") {
        Write-Host "Error: For HTTP mode, the certificate file must be a .pfx file." -ForegroundColor Red
        usage
    }
    
    # Only prompt for certificate password in HTTP mode
    if (-not $cert_password) {
        $cert_password = Read-Host "Please enter the certificate password (input will be hidden)" -AsSecureString
    }
    $cert_password_plain = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto(
        [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($cert_password)
    )
    
    # Validate the certificate password for HTTP mode
    if ([string]::IsNullOrEmpty($cert_password_plain)) {
        Write-Host "Error: Certificate password must be a non-empty string for HTTP mode." -ForegroundColor Red
        usage
    }
} elseif ($protocol -eq "mqtt") {
    # MQTT mode validations
    if (-not $mqtt_broker) {
        Write-Host "Error: MQTT broker address must be provided for MQTT mode" -ForegroundColor Red
        usage
    }
    if (-not $mqtt_port -or $mqtt_port -le 0) {
        Write-Host "Error: MQTT broker port must be a valid number" -ForegroundColor Red
        usage
    }
    
    # For MQTT mode, validate certificate and key files
    if (-not (Test-Path $cert_path)) {
        Write-Host "Error: Certificate file not found at path: $cert_path" -ForegroundColor Red
        usage
    }
    
    if (-not $key_path) {
        Write-Host "Error: Key path must be provided for MQTT mode" -ForegroundColor Red
        usage
    }
    
    if (-not (Test-Path $key_path)) {
        Write-Host "Error: Key file not found at path: $key_path" -ForegroundColor Red
        usage
    }
} else {
    Write-Host "Error: Protocol must be either 'http' or 'mqtt'." -ForegroundColor Red
    usage
}

# Validate the target name (non-empty string)
if ([string]::IsNullOrEmpty($target_name)) {
    Write-Host "Error: Target name must be a non-empty string." -ForegroundColor Red
    usage
}

# Validate the namespace (non-empty string)
if ([string]::IsNullOrEmpty($namespace)) {
    $namespace = "default"
    Write-Host "Using default namespace: $namespace" -ForegroundColor Yellow
}

# Validate the topology file (non-empty string)
if (-not (Test-Path $topology)) {
    Write-Host "Error: Topology file not found at path: $topology" -ForegroundColor Red
    usage
} elseif ($topology -notlike "*.json") {
    Write-Host "Error: The topology file must be a .json file." -ForegroundColor Red
    usage
}    

Import-Module PKI

# Create the JSON configuration based on protocol
Write-Host "Creating JSON configuration for $protocol mode..." -ForegroundColor Green
if ($protocol -eq "http") {
    $configJson = @{
        requestEndpoint = "$endpoint/solution/tasks"
        responseEndpoint = "$endpoint/solution/task/getResult"
        baseUrl = "$endpoint"
    } | ConvertTo-Json
} else {
    $configJson = @{
        mqttBroker = $mqtt_broker
        mqttPort = $mqtt_port
        targetName = $target_name
        namespace = $namespace
    } | ConvertTo-Json
}

# Save the JSON configuration to a file
$configFile = "config.json"
# Use utf8NoBOM encoding if available (PowerShell Core 6.0+), otherwise use ASCII as a fallback
if ($PSVersionTable.PSVersion.Major -ge 6) {
    $configJson | Set-Content -Path $configFile -Encoding utf8NoBOM
} else {
    # For older PowerShell versions, use ASCII to avoid BOM
    $configJson | Set-Content -Path $configFile -Encoding ASCII
}
Write-Host "Successfully created config file:" -ForegroundColor Green
Write-Host $configJson -ForegroundColor Cyan

# Convert paths to absolute paths
$cert_path = Resolve-Path $cert_path
$topology = Resolve-Path $topology
$config = Resolve-Path $configFile
if ($protocol -eq "mqtt" -and $key_path) {
    $key_path = Resolve-Path $key_path
}

# Protocol-specific processing
if ($protocol -eq 'http') {
    # Import the pfx certificate for HTTP mode
    Write-Host "Start to import pfx certificate" -ForegroundColor Blue
    try{
        $flags = [System.Security.Cryptography.X509Certificates.X509KeyStorageFlags]::PersistKeySet 
        $cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2(
            $cert_path, $cert_password_plain, $flags
        )
        Write-Host "Successfully imported pfx certificate" -ForegroundColor Green
        Write-Host "Cert Subject: $($cert.Subject)"
        Write-Host "Cert Thumbprint: $($cert.Thumbprint)"
        Write-Host "Has Private Key: $($cert.HasPrivateKey)"
    }
    catch {
        Write-Host "Exception Message: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host "Error: The certificate file is not a valid PFX file, or the password is incorrect, or the file is corrupt."  -ForegroundColor Red
        exit 1
    }
    
    # HTTP mode: Get certificates from server
    try {
        $WebRequestParams = @{
            Uri = "$($endpoint)/targets/getcert/$($target_name)?namespace=$($namespace)&osPlatform=windows"
            Method = 'Post'
            Certificate = $cert  
            Headers = @{ "Content-Type" = "application/json"; "User-Agent" = "PowerShell-Debug" }
        }
        Write-Host "WebRequestParams:" -ForegroundColor Cyan
        $WebRequestParams.GetEnumerator() | ForEach-Object { Write-Host ("  {0}: {1}" -f $_.Key, $_.Value) }
        $response = Invoke-WebRequest @WebRequestParams -Verbose
        Write-Host "Successfully got working certificates from symphony server" -ForegroundColor Green
    } catch {
        Write-Host "Error: Failed to send request to endpoint." -ForegroundColor Red
        Write-Host "Error Message: $($_.Exception.Message)" -ForegroundColor Red
        exit 1
    }
    
    # Process the response to extract certificates
    $jsonResponse = $response.Content | ConvertFrom-Json
    
    # Extract and format public certificate
    $public = $jsonResponse.public
    $header = ($public -split ' ')[0..1] -join ' '
    $footer = ($public -split ' ')[-3..-1] -join ' '
    $base64_content = ($public -split ' ')[2..(($public -split ' ').Length - 4)] -join "`n"
    $corrected_public_content = "$header`n$base64_content`n$footer"
    $corrected_public_content | Set-Content -Path "public.pem" -Encoding ascii
    Write-Host "Successfully created public.pem file" -ForegroundColor Green
    
    # Extract and format private key
    $private = $jsonResponse.private
    $header = ($private -split ' ')[0..3] -join ' '
    $footer = ($private -split ' ')[-5..-1] -join ' '
    $base64_content = ($private -split ' ')[4..(($private -split ' ').Length - 6)] -join "`n"
    $corrected_private_content = "$header`n$base64_content`n$footer"
    $corrected_private_content | Set-Content -Path "private.pem" -Encoding ascii
    Write-Host "Successfully created private.pem file" -ForegroundColor Green
    
    # Download remote-agent binary
    Write-Host "Begin to download remote-agent binary file" -ForegroundColor Blue
    try {
        $WebRequestParams = @{
            Uri = "$($endpoint)/files/remote-agent.exe"
            Method = 'Get'
            Certificate = $cert
        }
        Invoke-WebRequest @WebRequestParams -OutFile "remote-agent.exe" -ErrorAction Stop
        Write-Host "Successfully downloaded remote-agent.exe" -ForegroundColor Green
    } catch {
        Write-Host "Error: Failed to download." -ForegroundColor Red
        Write-Host "Error Message: $($_.Exception.Message)" -ForegroundColor Red
        exit 1
    }
    $agent_path = Resolve-Path ".\remote-agent.exe"
    
    # Set certificate paths for HTTP mode
    $public_path = Resolve-Path ".\public.pem"
    $private_path = Resolve-Path ".\private.pem"
    
} else {
    # MQTT mode: Prompt for binary
    $agent_path = Read-Host "Please input the full absolute path to your remote-agent.exe binary (e.g. C:\path\to\remote-agent.exe)"
    $agent_path = $agent_path.Trim('"')
    if (-not (Test-Path $agent_path)) {
        Write-Host "Error: remote-agent.exe not found at $agent_path" -ForegroundColor Red
        exit 1
    }
    
    # Handle CA certificate for MQTT
    if (-not $ca_cert_path) {
        Write-Host "CA certificate path not provided for MQTT." -ForegroundColor Yellow
        $provide_ca = Read-Host "Do you want to provide a CA certificate path? [y/N]"
        if ($provide_ca -eq "y" -or $provide_ca -eq "Y") {
            $ca_cert_path = Read-Host "Please enter the CA certificate path"
            $ca_cert_path = $ca_cert_path.Trim('"')
            if (-not (Test-Path $ca_cert_path)) {
                Write-Host "Error: CA certificate file not found at $ca_cert_path" -ForegroundColor Red
                exit 1
            }
        }
    } elseif (-not (Test-Path $ca_cert_path)) {
        Write-Host "Error: CA certificate file not found at $ca_cert_path" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "Using user-supplied remote-agent binary: $agent_path" -ForegroundColor Green
    if ($ca_cert_path) {
        $ca_cert_path = (Get-Item $ca_cert_path).FullName
        Write-Host "Using CA certificate: $ca_cert_path" -ForegroundColor Green
    }
    
    # For MQTT, use the provided certificate and key files
    $public_path = $cert_path
    $private_path = $key_path
    Write-Host "Using certificate: $public_path" -ForegroundColor Green
    Write-Host "Using private key: $private_path" -ForegroundColor Green
}

Write-Host "Begin to start remote agent process" -ForegroundColor Blue

# Compose the command line arguments for the remote agent
$processArgs = "-config=`"$config`" -client-cert=`"$public_path`" -client-key=`"$private_path`" -target-name=`"$target_name`" -namespace=`"$namespace`" -topology=`"$topology`" -protocol=`"$protocol`""

# Add MQTT-specific parameters
if ($protocol -eq 'mqtt') {
    if ($ca_cert_path) {
        $processArgs += " -ca-cert=`"$ca_cert_path`""
    }
    if ($use_cert_subject) {
        $processArgs += " -use-cert-subject"
    }
}

$binPath = "`"$agent_path`" $processArgs"
$serviceName = "symphony-service"
Write-Host "Remote agent command line: $binPath" -ForegroundColor Cyan
# Setup as service or scheduled task
if ($run_mode -eq 'service') {
    Write-Host "[run_mode=service] Register and start as Windows service..." -ForegroundColor Cyan
    
    # Check if service already exists
    if (Get-Service -Name $serviceName -ErrorAction SilentlyContinue) {
        Stop-Service -Name $serviceName -Force -ErrorAction SilentlyContinue
        sc.exe delete $serviceName | Out-Null
        Write-Host "Service $serviceName deleted." -ForegroundColor Yellow
    }
    
    # Clean up any event log registry entries
    $eventLogRegPath = "HKLM:\SYSTEM\CurrentControlSet\Services\EventLog\Application\$serviceName"
    if (Test-Path $eventLogRegPath) {
        Remove-Item -Path $eventLogRegPath -Recurse -Force
        Write-Host "EventLog registry key $eventLogRegPath deleted." -ForegroundColor Yellow
    }

    # Register the service
    Write-Host "Registering $serviceName as a Windows service..." -ForegroundColor Blue
    sc.exe create $serviceName binPath= "$binPath" DisplayName= "$serviceName" start= auto
    
    # Start the service
    try {
        Start-Service -Name $serviceName -ErrorAction Stop
        Write-Host "Successfully registered and started $serviceName as a Windows service" -ForegroundColor Green
    } catch {
        Write-Host "Failed to start $serviceName. Please check the service logs." -ForegroundColor Red
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
        # Continue execution rather than throwing
    }
} elseif ($run_mode -eq 'schedule') {
    # Setup as a scheduled task
    Write-Host "[run_mode=schedule] Register and start as scheduled task..." -ForegroundColor Cyan
    
    # Create watchdog script
    $watchdogScript = @"
try {
    if (-not (Get-Process -Name "remote-agent" -ErrorAction SilentlyContinue)) {
        Start-Process -FilePath '$agent_path' -ArgumentList '$processArgs' -WindowStyle Hidden
    }
} catch {
    Write-Host "Error in watchdog: `$_.Exception.Message" -ForegroundColor Red
}
"@

    $watchdogPath = Join-Path (Split-Path $agent_path) "watchdog-remote-agent.ps1"
    $watchdogScript | Set-Content -Path $watchdogPath -Encoding UTF8

    # Setup scheduled task
    $Trigger = New-ScheduledTaskTrigger -Once -At (Get-Date).AddMinutes(1) -RepetitionInterval (New-TimeSpan -Minutes 1) -RepetitionDuration (New-TimeSpan -Days 3650)
    $Action = New-ScheduledTaskAction -Execute "pwsh.exe" -Argument "-NoProfile -WindowStyle Hidden -File `"$watchdogPath`""
    $TaskName = "RemoteAgentTask"
    
    if (Get-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue) {
        Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false
    }
    
    $currentUser = [System.Security.Principal.WindowsIdentity]::GetCurrent().Name
    Register-ScheduledTask -TaskName $TaskName -Action $Action -Trigger $Trigger -Description "Guard remote-agent.exe process" -User $currentUser -RunLevel Highest
    Start-ScheduledTask -TaskName $TaskName
    
    Write-Host "Registered and started scheduled task $TaskName" -ForegroundColor Green
} else {
    Write-Host "Error: Invalid run_mode '$run_mode'. Must be either 'service' or 'schedule'." -ForegroundColor Red
    exit 1
}

Write-Host "Setup complete!" -ForegroundColor Green
