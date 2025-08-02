#!/usr/bin/env pwsh
<#
    .SYNOPSIS
        Builds the Wails application and updates the version in wails.json.
    .DESCRIPTION
        This script builds the Wails application, updates the version in
        wails.json based on the current Git commit, and generates SHA256
        and SHA1 hashes for the installer files.
#>

#Requires -Version 7.0

[CmdletBinding()]
param (
    [Parameter(Mandatory = $false, HelpMessage = "Skip the build process and just update the wails.json file.")]
    [Alias("nb")]
    [switch]$NoBuild,
    [Parameter(Mandatory = $false, HelpMessage = "(Implies -NoBuild) Skip the version update in wails.json and just show the current version. Useful for checking the current version without making changes.")]
    [Alias("nv")]
    [switch]$NoVersionUpdate
)

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

function Invoke-WailsBuild {
    $Date = $(Get-Date -AsUTC -Format "yyyy-MM-ddTHH:mm:ssZ")
    $Commit = $(git rev-parse HEAD)
    if ( $(git describe --tags HEAD 2> $null) -match "v?\d+\.\d+\.\d+(?:-\w+(?:-\d+)*)?") {
        $Version = $Matches[0]
    } else {
        $Version = "v0.0.0-dev-${Commit}"
    }

    if ($NoVersionUpdate) {
        $NoBuild = $false
        Write-Host "Current version: $Version"
        return
    } else {
        $WailsJsonPath = Join-Path $PSScriptRoot "wails.json"
        $WailsData = Get-JsonContent -Path $WailsJsonPath
        $WailsData.Info.productVersion = $Version
        Set-JsonContent -Path $WailsJsonPath -Value $WailsData
    }

    if (-not $NoBuild) {
        Push-Location $PSScriptRoot -StackName "project-root"
        wails build -clean -ldflags "-X main.Version=${Version} -X main.Date=${Date} -X main.Commit=${Commit}" -nsis -upx -v 2
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
    }
}

Invoke-WailsBuild