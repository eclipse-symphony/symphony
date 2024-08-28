param(
    [String]$DeploymentFile,
    [String]$ComponentListFile
)

# Load the JSON from the input file
$json = Get-Content -Encoding UTF8 $ComponentListFile | ConvertFrom-Json

# Loop through the components and remove those with app.package equals to "notepad" if notepad process is not running
foreach ($component in $json) {
    $fileName = ".\scripts\" + $component.Properties."bin.name" + ".txt"
    # check if file not exists
    if (-not (Test-Path $fileName)) {
        # Remove the component from the Components list
        $json= $json | Where-Object { $_ -ne $component }
    }    
}

# Write the updated JSON to an output file
"[" + ($json | ForEach-Object {$_.Component} | ConvertTo-Json -Compress) + "]" | Out-File -Encoding ASCII $DeploymentFile.Replace(".json", "-output.json")