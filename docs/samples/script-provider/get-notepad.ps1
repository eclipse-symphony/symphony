param(
    [String]$InputFile
)

# Load the JSON from the input file
$json = Get-Content -Encoding UTF8 $InputFile | ConvertFrom-Json

# Loop through the components and remove those with app.package equals to "notepad" if notepad process is not running
foreach ($component in $json.Solution.Components) {
    if ($component.Properties."app.package" -eq "notepad") {
        if ((Get-Process -Name "notepad" -ErrorAction SilentlyContinue) -eq $null) {
            # Remove the component from the Components list
            $json.Solution.Components = $json.Solution.Components | Where-Object { $_ -ne $component }
        }
    }
}

# Write the updated JSON to an output file
"[" + ($json.Solution.Components | ConvertTo-Json -Compress) + "]" | Out-File -Encoding ASCII $InputFile.Replace(".json", "-output.json")