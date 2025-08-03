<#
    .SYNOPSIS
        Builds the Wails application and updates the version in wails.json.
    .DESCRIPTION
        This script builds the Wails application, updates the version in
        wails.json based on the current Git commit, and generates SHA256
        and SHA1 hashes for the installer files.

        To preserve the current state of the repository, it reverts changes to
        the wails.json and project.nsi files after building. It also checks
        if the repository is clean before proceeding with the build. If there are
        uncommitted changes, the script will not proceed and will display an error message.

        By default, it shows the current version and file version (based on the
        latest Git commit) that would be used for the build, but without actually
        performing the build.

        You can use the -Build, -NSIS, and -UPX switches to
        perform the build, create an NSIS installer, and compress the executable.
    .PARAMETER Build
        Run the build process without generating the NSIS installer.
    .PARAMETER NSIS
        Implies -Build. Create the NSIS installer after building.
    .PARAMETER UPX
        Implies -Build. Use UPX to compress the executable file.
#>

#Requires -Version 7.0

[CmdletBinding()]
param (
    [Parameter(Mandatory = $false, HelpMessage = "Run the build process without generating the NSIS installer.")]
    [Alias("b","exe")]
    [switch]$Build,
    [Parameter(Mandatory = $false, HelpMessage = "(Implies -Build) Create the NSIS installer after building.")]
    [Alias("i","installer")]
    [switch]$NSIS,
    [Parameter(Mandatory = $false, HelpMessage = "(Implies -Build) Use UPX to compress the executable file.")]
    [Alias("u","c","compress","compact")]
    [switch]$UPX
)

# Set the default behavior to show the current version and file version
# without performing the build.
$ShowVersionOnly = $true
if ($Build -or $NSIS -or $UPX) {
    $ShowVersionOnly = $false
}

# Change to the script's directory (Assumes the script is in the root of the project)
Set-Location $PSScriptRoot

# Get the script name
$ScriptName = $MyInvocation.MyCommand.Name

function Get-JsonContent {
    <#
    .SYNOPSIS
        Reads a JSON file and converts it into a PowerShell object.
    .DESCRIPTION
        This function reads the content of a JSON file and converts it into a PowerShell object.
    .PARAMETER Path
        The path to the JSON file to read.
    #>
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
    <#
    .SYNOPSIS
        Writes a PowerShell object to a JSON file.
    .DESCRIPTION
        This function takes a PowerShell object and writes it to a JSON file.
    .PARAMETER Path
        The path to the JSON file to write.
    #>
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
    <#
    .SYNOPSIS
        Generates a timestamp based on the current date and time.
    .DESCRIPTION
        This function creates a timestamp by calculating the number of 10-minute intervals
        that have elapsed since the beginning of the year.
    .OUTPUTS
        Returns an integer representing the number of 10-minute intervals since the start of the year.
        Integer value will be in the range of 1 to approximately 52560 (for a full year).
    #>
    $DateNow = Get-Date -AsUTC
    $TicksPerDay = 24 * 6   # number of 10 minute intervals in a day
    $DayOfYear = $DateNow.DayOfYear - 1
    $HourTicksToday = $DateNow.Hour * 6
    $TicksThisHour = int($DateNow.Minute / 10)
    $result = ($DayOfYear * $TicksPerDay) + $HourTicksToday + $TicksThisHour
    return int($result)
}

function Confirm-RepositoryIsClean {
    <#
    .SYNOPSIS
        Checks if the Git repository is clean (no uncommitted changes).
    .DESCRIPTION
        This function checks the status of the Git repository and determines if there are any
        uncommitted changes. It returns $true if the repository is clean, and $false otherwise.
    #>
    $status = git status --porcelain
    if ($status) {
        $isClean = $true
        $statusList = $status -split "`n"
        $newStatusList = @()
        foreach ($line in $statusList) {
            if ( -not ($line -match $ScriptName) ) {
                $isClean = $false
                $newStatusList += $line
            }
        }

        if ($isClean) { return $true }

        Write-Host ""
        Write-Host -ForegroundColor Red "WARNING: Repository is not clean. Please commit or stash your changes before building."
        Write-Host ""
        Write-Host "Uncommitted changes:"
        Write-Host ""
        $newStatusList | ForEach-Object { Write-Host -ForegroundColor Yellow "  $_" }
        Write-Host ""
        Write-Host "Suggestions:"
        Write-Host "  - Commit your changes: git commit -m 'Your commit message'"
        Write-Host "  - Create a new branch: git checkout -b new-branch-name "
        Write-Host "       and commit your changes to the branch"
        Write-Host "  - Stash your changes: git stash --all"
        Write-Host "  - Discard your changes: git reset --hard HEAD"
        Write-Host ""
        Write-Host -ForegroundColor Cyan "NOTE: Script ignores changes to the script itself ($ScriptName)"
        Write-Host ""
        return $false
    }
    return $true
}

function Restore-RepositoryToCleanState {
    <#
    .SYNOPSIS
        Restores the Git repository to a clean state by discarding uncommitted changes.
    .DESCRIPTION
        This function checks for uncommitted changes in the specified files and restores them
        to their last committed state.

        As written, it will only restore changes to the wails.json and project.nsi files.
        If you want to restore other files, add them to the $FileList array.
    #>
    Push-Location $PSScriptRoot -StackName "restoreproject"
    Write-Host -ForegroundColor Yellow "Restoring repository to a clean state..."
    $FileList = @(
        "wails.json",
        "build/windows/installer/project.nsi"
    )
    foreach ($file in $FileList) {
        if ( (git status --porcelain | Select-String "$file") ) {
            Write-Host -ForegroundColor Yellow "  Restoring: $file"
            git restore "$file"
        }
    }
    Pop-Location -StackName "restoreproject"
}

function Update-ProjectNSI {
    <#
    .SYNOPSIS
        Updates the NSIS project file with the specified file version.
    .DESCRIPTION
        This function modifies the NSIS project file to set the correct file version.
    .PARAMETER ProjectNSIPath
        The path to the NSIS project file to update.
    .PARAMETER FileVersion
        The new file version to set in the project file.
    #>
    param (
        [Parameter(Mandatory = $true)]
        [string]$ProjectNSIPath,
        [string]$FileVersion
    )

    if (-not (Test-Path -Path $ProjectNSIPath)) {
        Write-Host -ForegroundColor Red "Project NSI file not found at path: $ProjectNSIPath"
        return
    }

    Write-Host -ForegroundColor Cyan "Updating NSI project file: $ProjectNSIPath"

    $NSIData = Get-Content -Path $ProjectNSIPath
    $NewNSIData = @()
    $FileChanged = $false

    foreach ($line in $NSIData) {
        if ($line -match '^(?<key>VIFileVersion\s+)"(?<value>[^"]+)"\s*$') {
            if ($Matches.value -ne $FileVersion) {
                $newval1 = $Matches.key + '"' + $FileVersion + '"'
                $line = $line -replace '^(VIFileVersion\s+)"([^"]+)"\s*$', $newval1
                Write-Host -ForegroundColor Green "  Updating VIFileVersion to: $FileVersion"
                $FileChanged = $true
            }
        }
        if ($line -match '^(?<key>VIProductVersion\s+)"(?<value>[^"]+)"\s*$') {
            if ($Matches.value -ne $FileVersion) {
                Write-Host -ForegroundColor Green "  Updating VIProductVersion to: $FileVersion"
                $newval2 = $Matches.key + '"' + $FileVersion + '"'
                $line = $line -replace '^(VIProductVersion\s+)"([^"]+)"\s*$', $newval2
                $FileChanged = $true
            }
        }
        if ($line -match '^(?<key>VIAddVersionKey\s+"FileVersion"\s+)"(?<value>[^"]+)"\s*$') {
            if ($Matches.value -ne $FileVersion) {
                Write-Host -ForegroundColor Green "  Updating VIAddVersionKey FileVersion to: $FileVersion"
                $newval3 = $Matches.key + '"' + $FileVersion + '"'
                $line = $line -replace '^(VIAddVersionKey\s+"FileVersion"\s+)"([^"]+)"\s*$', $newval3
                $FileChanged = $true
            }
        }
        $NewNSIData += $line
    }
    if (-not $FileChanged) {
        Write-Host -ForegroundColor Yellow "  No changes made to NSIS project file"
        return
    }
    Set-Content -Path $ProjectNSIPath -Encoding utf8 -Value $NewNSIData
}

function Update-WailsJSON {
    <#
    .SYNOPSIS
        Updates the wails.json file with the specified version.
    .DESCRIPTION
        This function modifies the wails.json file to set the correct version.
    .PARAMETER WailsJsonPath
        The path to the wails.json file to update.
    .PARAMETER Version
        The new version to set in the wails.json file.
    #>
    param (
        [Parameter(Mandatory = $true)]
        [string]$WailsJsonPath,
        [string]$Version
    )

    if (-not (Test-Path -Path $WailsJsonPath)) {
        Write-Host -ForegroundColor Red "wails.json file not found at path: $WailsJsonPath"
        return
    }

    Write-Host -ForegroundColor Cyan "Updating wails.json with version value: $Version"
    $WailsData = Get-JsonContent -Path $WailsJsonPath

    if (-not $WailsData) {
        Write-Host -ForegroundColor Red "  Failed to read wails.json or it is empty."
        return
    }
    if ($WailsData.Info.productVersion -ne $Version) {
        $WailsData.Info.productVersion = $Version
        Set-JsonContent -Path $WailsJsonPath -Value $WailsData
        Write-Host -ForegroundColor Green "  Updating productVersion to: $Version"
    } else {
        Write-Host -ForegroundColor Yellow "  No changes made to wails.json, version is already set to: $Version"
    }
}

function Invoke-WailsBuild {
    <#
    .SYNOPSIS
        Invokes the Wails build process with the specified options.
    .DESCRIPTION
        This function prepares and executes the Wails build command with the appropriate
        flags and environment variables.
    #>
    $NSISOption = ""
    $UPXOption  = ""
    $ProjectNSI = Join-Path $PSScriptRoot "build" "windows" "installer" "project.nsi"
    $WailsJsonPath = Join-Path $PSScriptRoot "wails.json"

    if ($NSIS) {
        $NSISOption = "-nsis"
        $Build = $true
    }
    if ($UPX) {
        $UPXOption = "-upx"
        $Build = $true
    }

    $Date = $(Get-Date -AsUTC -Format "yyyy-MM-ddTHH:mm:ssZ")
    $Version = ""
    $FileVersion = ""
    $RC = @("alpha", "beta", "rc", "")
    $Commit = $(git rev-parse --short HEAD)
    $CurrentCommitTag = $(git describe --tags HEAD 2> $null)
    if ( $CurrentCommitTag -match "v?(?<major>\d+)\.(?<minor>\d+)\.(?<patch>\d+)(?:-(?<prerelease>(?:0|[1-9]\d*|\w+\d*)))(?:-(?<ahead>\d+)(?:-g?(?<hash>[0-9a-fA-F]+))?)?$") {
        $Major = $Matches.major
        $Minor = $Matches.minor
        $Patch = $Matches.patch
        $Prerelease = $Matches.prerelease
        $Ahead = $Matches.ahead
        $Hash = $Matches.hash
        if ($Hash) { $Hash | Out-Null }  # suppress unused variable warning
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


    if ($ShowVersionOnly) {
        $Build = $false
        Write-Host "   Current App version: $Version"
        Write-Host "     NSIS File version: $FileVersion"
        return
    }

    if (-not (Confirm-RepositoryIsClean)) {
        return
    }

    if ($Build) {
        if ($NSIS) {
            Update-ProjectNSI -ProjectNSIPath $ProjectNSI -FileVersion $FileVersion
        }
        Update-WailsJSON -WailsJsonPath $WailsJsonPath -Version $Version

        Push-Location $PSScriptRoot -StackName "project-root"
        Write-Host -NoNewLine -ForegroundColor Cyan "Building Wails application with version value: "
        Write-Host -NoNewLine -ForegroundColor Black -BackgroundColor White "$Version"
        Write-Host "`n`n$('=' * 78)"
        wails build -clean -ldflags "-X main.Version=${Version} -X main.Date=${Date} -X main.Commit=${Commit}" ${NSISOption} ${UPXOption}
        Write-Host "$('=' * 78)`n"

        if ($NSIS) {
            Write-Host -ForegroundColor Cyan "Writing sha256 and sha1 hashes for installer files..."
            Set-Location ./build/bin
            foreach ($file in Get-ChildItem *-installer.exe -File -ErrorAction SilentlyContinue) {
                $sha256Name = $file.Name + ".sha256"
                $sha1Name = $file.Name + ".sha1"
                $sha256Hash = (Get-FileHash $file -Algorithm SHA256).Hash # | ForEach-Object { $_.Hash } | Out-File -FilePath $sha256Name -Encoding utf8
                $sha1Hash = (Get-FileHash $file -Algorithm SHA1).Hash # | ForEach-Object { $_.Hash } | Out-File -FilePath $sha1Name -Encoding utf8
                "$sha256Hash  *$($file.Name)" | Out-File -FilePath $sha256Name -Encoding utf8
                "$sha1Hash  *$($file.Name)" | Out-File -FilePath $sha1Name -Encoding utf8
            }
        }

        Restore-RepositoryToCleanState
        Pop-Location -StackName "project-root"
    }
}

Invoke-WailsBuild