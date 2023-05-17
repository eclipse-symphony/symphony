param (
    [Parameter(Mandatory=$true)]
    [string]$file1,

    [Parameter(Mandatory=$true)]
    [string]$file2
)

# Load the JSON from the input files
$json1 = Get-Content -Path $file1 -Raw | ConvertFrom-Json
$json2 = Get-Content -Path $file2 -Raw | ConvertFrom-Json

# Convert the JSON objects to strings and compare them for equivalence
if ($json1 | ConvertTo-Json -Compress -Depth 100 -ErrorAction SilentlyContinue -WarningAction SilentlyContinue -InformationAction SilentlyContinue -OutVariable str1 | Out-Null) {
    $str1 = $json1 | Out-String
}
if ($json2 | ConvertTo-Json -Compress -Depth 100 -ErrorAction SilentlyContinue -WarningAction SilentlyContinue -InformationAction SilentlyContinue -OutVariable str2 | Out-Null) {
    $str2 = $json2 | Out-String
}

if ($str1 -eq $str2) {
    Write-Output "false"
}
else {
    Write-Output "true"
}