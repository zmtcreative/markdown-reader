#!/usr/bin/env pwsh
Set-Location $PSScriptRoot

$Date = $(Get-Date -AsUTC -Format "yyyy-MM-ddTHH:mm:ssZ")
$Commit = $(git rev-parse HEAD)
if ( $(git describe --tags HEAD 2> $null) -match "v?\d+\.\d+\.\d+(-\w+)?") {
    $Version = $Matches[0]
} else {
    $Version = "v0.0.0-dev-${Commit}"
}

# echo $Version
# echo $Date
# echo $Commit

Push-Location $PSScriptRoot
wails build -ldflags "-X main.Version=${Version} -X main.Date=${Date} -X main.Commit=${Commit}"
Pop-Location