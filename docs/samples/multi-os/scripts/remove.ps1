param(
    [String]$DeploymentFile,
    [String]$ComponentListFile
)

# Load the JSON from the input file
$json = Get-Content -Encoding UTF8 $ComponentListFile | ConvertFrom-Json

$output = @{}

# Loop through the components
foreach ($component in $json) {
    $componentName = $component.Properties."bin.name"
    Invoke-Expression ".\scripts\hello_world.cmd"
    Remove-Item (".\scripts\" + $componentName + ".txt")

    $output[$componentName] = @{
       "status" = 8004
       "message" = ""
    }
}

# Convert the output hashtable to JSON
$jsonOutput = $output | ConvertTo-Json -Compress


# Write the JSON to an output file
Out-File -Encoding ASCII -FilePath $DeploymentFile.Replace(".json", "-output.json") -InputObject $jsonOutput