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
    [Alias("n")]
    [switch]$NoBuild,
    [Parameter(Mandatory = $false, HelpMessage = "Skip the NSIS installer generation and just build the application.")]
    [Alias("e","exe","app")]
    [switch]$ExeOnly,
    [Parameter(Mandatory = $false, HelpMessage = "(Implies -NoBuild) Skip the version update in wails.json and just show the current version. Useful for checking the current version without making changes.")]
    [Alias("s","v","show")]
    [switch]$ShowVersionOnly,
    [Parameter(Mandatory = $false, HelpMessage = "Use UPX to compress the executable files.")]
    [Alias("u")]
    [switch]$UPX
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

function Get-DateStamp {
    $DateNow = Get-Date -AsUTC
    $TicksPerDay = 24 * 6   # number of 10 minute intervals in a day
    $DayOfYear = $DateNow.DayOfYear - 1
    $HourTicksToday = $DateNow.Hour * 6
    $TicksThisHour = int($DateNow.Minute / 10)
    $result = ($DayOfYear * $TicksPerDay) + $HourTicksToday + $TicksThisHour
    return int($result)
}

function Update-ProjectNSI {
    param (
        [Parameter(Mandatory = $true)]
        [string]$ProjectNSIPath,
        [string]$FileVersion
    )

    if (-not (Test-Path -Path $ProjectNSIPath)) {
        Write-Error "Project NSI file not found at path: $ProjectNSIPath"
        return
    }

    Write-Host -ForegroundColor Green "Updating NSI project file: $ProjectNSIPath"

    $NSIData = Get-Content -Path $ProjectNSIPath
    $NewNSIData = @()

    foreach ($line in $NSIData) {
        if ($line -match '^(VIFileVersion\s+)"([^"]+)"\s*$') {
            Write-Host -ForegroundColor Yellow "  Updating VIFileVersion in NSIS project file"
            $newval1 = $Matches[1] + '"' + $FileVersion + '"'
            $line = $line -replace '^(VIFileVersion\s+)"([^"]+)"\s*$', $newval1
        }
        if ($line -match '^(VIProductVersion\s+)"([^"]+)"\s*$') {
            Write-Host -ForegroundColor Yellow "  Updating VIProductVersion in NSIS project file"
            $newval2 = $Matches[1] + '"' + $FileVersion + '"'
            $line = $line -replace '^(VIProductVersion\s+)"([^"]+)"\s*$', $newval2
        }
        if ($line -match '^(VIAddVersionKey\s+"FileVersion"\s+)"([^"]+)"\s*$') {
            Write-Host -ForegroundColor Yellow "  Updating VIAddVersionKey FileVersion in NSIS project file"
            $newval3 = $Matches[1] + '"' + $FileVersion + '"'
            $line = $line -replace '^(VIAddVersionKey\s+"FileVersion"\s+)"([^"]+)"\s*$', $newval3
        }
        $NewNSIData += $line
    }
    Set-Content -Path $ProjectNSIPath -Encoding utf8 -Value $NewNSIData
}

function Invoke-WailsBuild {
    $NSIS = "-nsis"
    $UPX = ""
    $ProjectNSI = Join-Path $PSScriptRoot "build" "windows" "installer" "project.nsi"
    $Date = $(Get-Date -AsUTC -Format "yyyy-MM-ddTHH:mm:ssZ")
    $Version = ""
    $FileVersion = ""
    $RC = @("alpha", "beta", "rc", "")
    $Commit = $(git rev-parse --short HEAD)
    if ( $(git describe --tags HEAD 2> $null) -match "v?(?<major>\d+)\.(?<minor>\d+)\.(?<patch>\d+)(?:-(?<prerelease>(?:0|[1-9]\d*|\w+\d*)))(?:-(?<ahead>\d+)(?:-g?(?<hash>[0-9a-fA-F]+))?)?$") {
        $Major = $Matches.major
        $Minor = $Matches.minor
        $Patch = $Matches.patch
        $Prerelease = $Matches.prerelease
        $Ahead = $Matches.ahead
        $Hash = $Matches.hash
        $Version = $Major + "." + $Minor + "." + $Patch
        $FileVersion = $Version
        if ($Prerelease) {
            if ($Prerelease -match "^(?<rc>alpha|beta|rc)?(?<rcnum>\d+)?$") {
                $rcString = $Matches.rc
                $rcNumber = [int]$Matches.rcnum || 0
                $rcIDX = $RC.IndexOf($rcString) + 1
            }

            $PreReleaseNumber = ($rcIDX * 10000) + ($rcNumber * 100)
            $Version += "-" + $Prerelease
        }
        if ($Ahead) {
            $Version += "+" + $Ahead
            if ($PreReleaseNumber) {
                $PreReleaseNumber += $Ahead
            }
        }
        $FileVersion = $FileVersion + '.' + $PreReleaseNumber
    } else {
        $Version = "0.0.0-dev+${Commit}"
        $ds = Get-DateStamp
        $FileVersion = "0.0.0.${ds}"
    }

    if ($ExeOnly) { $NSIS = "" } else {
        Update-ProjectNSI -ProjectNSIPath $ProjectNSI -FileVersion $FileVersion
    }
    if ($UPX) { $UPX = "-upx" }

    if ($ShowVersionOnly) {
        $NoBuild = $true
        Write-Host "   Current App version: $Version"
        Write-Host "     NSIS File version: $FileVersion"
        return
    } else {
        Write-Host -ForegroundColor Green "Updating wails.json with version value: $FileVersion"
        $WailsJsonPath = Join-Path $PSScriptRoot "wails.json"
        $WailsData = Get-JsonContent -Path $WailsJsonPath
        $WailsData.Info.productVersion = $Version
        Set-JsonContent -Path $WailsJsonPath -Value $WailsData
    }

    if (! $NoBuild) {
        Push-Location $PSScriptRoot -StackName "project-root"
        Write-Host -ForegroundColor Green "Building Wails application with version value: $Version"
        wails build -clean -ldflags "-X main.Version=${Version} -X main.Date=${Date} -X main.Commit=${Commit}" ${NSIS} ${UPX}
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