#!/usr/bin/env pwsh
Set-Location $PSScriptRoot

function Get-JsonContent {
    [CmdletBinding()]
    param (
        [Parameter(Mandatory = $true, ValueFromPipelineByPropertyName = $true, Position = 0)]
        [string]$Path
    )

    process {
        if (-not (Test-Path -Path $Path -PathType Leaf)) {
            Write-Error "File not found at path: $Path"
            return
        }
        try {
            # Read the file and convert it from JSON into a PowerShell object.
            # PowerShell automatically creates an ordered object when parsing from JSON text.
            return Get-Content -Raw -Path $Path | ConvertFrom-Json
        }
        catch {
            Write-Error "Failed to read or parse JSON from '$Path'. Error: $_"
        }
    }
}

function Set-JsonContent {
    [CmdletBinding()]
    param (
        [Parameter(Mandatory = $true, Position = 0)]
        [string]$Path,

        [Parameter(Mandatory = $true, ValueFromPipeline = $true, Position = 1)]
        [psobject]$Value
    )

    process {
        try {
            # Convert the incoming object to a JSON string and write it to the specified file.
            # The -Depth parameter prevents truncation of nested objects.
            $Value | ConvertTo-Json -Depth 100 | Out-File -FilePath $Path -Encoding utf8 -NoNewline
        }
        catch {
            Write-Error "Failed to write JSON to '$Path'. Error: $_"
        }
    }
}

# # --- Example Usage ---

# # 1. Create an ordered PowerShell object.
# $myData = [ordered]@{
#     "name"      = "markdown-reader"
#     "version"   = "1.0.0"
#     "author"    = "GitHub Copilot"
#     "features"  = @(
#         "Read Markdown",
#         "Render HTML",
#         "Syntax Highlighting"
#     )
#     "enabled"   = $true
#     "builds"    = 5
# }

# $jsonFilePath = ".\example.json"

# # 2. Write the ordered object to a JSON file.
# Set-JsonContent -Path $jsonFilePath -Value $myData

# # 3. Read the data back from the JSON file.
# $readData = Get-JsonContent -Path $jsonFilePath

# # 4. Display the object to verify the data and order.
# Write-Output "--- Data read from file ---"
# $readData | Format-List

# # 5. Display the raw JSON file content to verify formatting.
# Write-Output "`n--- Raw content of $($jsonFilePath) ---"
# Get-Content $jsonFilePath

$Date = $(Get-Date -AsUTC -Format "yyyy-MM-ddTHH:mm:ssZ")
$Commit = $(git rev-parse HEAD)
if ( $(git describe --tags HEAD 2> $null) -match "v?\d+\.\d+\.\d+(-\w+)?") {
    $Version = $Matches[0]
} else {
    $Version = "v0.0.0-dev-${Commit}"
}

$WailsJsonPath = Join-Path $PSScriptRoot "wails.json"
$WailsData = Get-JsonContent -Path $WailsJsonPath
$WailsData.Info.productVersion = $Version
Set-JsonContent -Path $WailsJsonPath -Value $WailsData

Push-Location $PSScriptRoot -StackName "project-root"
echo `wails build -clean -ldflags "-X main.Version=${Version} -X main.Date=${Date} -X main.Commit=${Commit}" -nsis -upx`
Set-Location ./build/bin
foreach ($file in Get-ChildItem *-installer.exe -File -ErrorAction SilentlyContinue) {
    $sha256Name = $file.Name + ".sha256"
    $sha1Name = $file.Name + ".sha1"
    $sha256Hash = (Get-FileHash $file -Algorithm SHA256).Hash # | ForEach-Object { $_.Hash } | Out-File -FilePath $sha256Name -Encoding utf8
    $sha1Hash = (Get-FileHash $file -Algorithm SHA1).Hash # | ForEach-Object { $_.Hash } | Out-File -FilePath $sha1Name -Encoding utf8
    "$sha256Hash  *$($file.Name)" | Out-File -FilePath $sha256Name -Encoding utf8
    "$sha1Hash  *$($file.Name)" | Out-File -FilePath $sha1Name -Encoding utf8
}
Pop-Location -StackName "project-root"