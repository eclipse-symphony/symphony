param(
    [String]$InputFile
)

# Load the JSON from the input file
$json = Get-Content -Encoding UTF8 $InputFile | ConvertFrom-Json

# Loop through the components and remove those with app.package equals to "notepad" if notepad process is not running
foreach ($component in $json.SolutionVersion.Components) {
    if ($component.Properties."app.package" -eq "notepad") {
        if ((Get-Process -Name "notepad" -ErrorAction SilentlyContinue) -eq $null) {
            # Remove the component from the Components list
            $json.SolutionVersion.Components = $json.SolutionVersion.Components | Where-Object { $_ -ne $component }
        }
    }
}

# Write the updated JSON to an output file
"[" + ($json.SolutionVersion.Components | ConvertTo-Json -Compress) + "]" | Out-File -Encoding ASCII $InputFile.Replace(".json", "-output.json")