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

Push-Location $PSScriptRoot -StackName "project-root"
wails build -clean -ldflags "-X main.Version=${Version} -X main.Date=${Date} -X main.Commit=${Commit}" -nsis -upx
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