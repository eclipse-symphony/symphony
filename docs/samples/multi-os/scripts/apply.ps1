param(
    [String]$DeploymentFile,
    [String]$ComponentListFile
)

# Load the JSON from the input file
$json = Get-Content -Encoding UTF8 $ComponentListFile | ConvertFrom-Json

$output = @{}

# Loop through the components
foreach ($component in $json) {
    # Print the current component being processed
    Write-Output "Processing Component: "
    Write-Output $component

    $componentName = $component.Properties."bin.name"
    $fileName = ".\scripts\" + $componentName    

    # Print the current component being processed
    Write-Output "Component name: "
    Write-Output $componentName

    Invoke-Expression ($fileName + ".cmd")
    New-Item -Path ($fileName + ".txt") -ItemType "file" -Force

    $output[$componentName] = @{
       "status" = 8004
       "message" = ""
    }
}

# Convert the output hashtable to JSON
$jsonOutput = $output | ConvertTo-Json -Compress


# Write the JSON to an output file
Out-File -Encoding ASCII -FilePath $DeploymentFile.Replace(".json", "-output.json") -InputObject $jsonOutput