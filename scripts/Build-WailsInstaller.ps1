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
    .PARAMETER RunAllTests
        Run the full frontend test suite when invoking Run-AllTests.ps1 before building.
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
    [switch]$UPX,
    [Parameter(Mandatory = $false, HelpMessage = "Run the full frontend test suite before building.")]
    [switch]$RunAllTests
)

# Set up script and project paths
$ScriptFullName = $MyInvocation.MyCommand.Path
$ScriptRoot = Split-Path -Parent $ScriptFullName
$ScriptName = Split-Path -Leaf $ScriptFullName
if ($ScriptRoot -match '[\\/]scripts[\\/]?$') {
    $tmpProjectRoot = $ScriptRoot -replace '[\\/]scripts[\\/]?', ''
} else {
    $tmpProjectRoot = $ScriptRoot
}
if (Test-Path -Path "$tmpProjectRoot\wails.json") {
    $ProjectRoot = $tmpProjectRoot
} else {
    Write-Host -ForegroundColor Red "Could not find wails.json in the expected project root: $tmpProjectRoot"
    exit 1
}

# Set the default behavior to show the current version and file version
# without performing the build.
$ShowVersionOnly = $true
if ($Build -or $NSIS -or $UPX) {
    $ShowVersionOnly = $false
}

$FileList = @(
    "wails.json",
    "frontend/package.json",
    "build/windows/info.json",
    "build/windows/installer/project.nsi"
)

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
    $TicksThisHour = [int]($DateNow.Minute / 10)
    $result = ($DayOfYear * $TicksPerDay) + $HourTicksToday + $TicksThisHour
    return [int]$result
}

function Get-VersionHash {
    <#
    .SYNOPSIS
        Retrieves the version hash from a Git tag name structured as a semantic version.
    .DESCRIPTION
        This function extracts the version information from the specified Git tag name.
        It assumes the tag follows the semantic versioning format:
        vMAJOR.MINOR.PATCH[-PRERELEASE][+AHEAD-HASH].
    .PARAMETER TagName
        The name of the Git tag to parse for version information.
    #>
    param (
        [Parameter(Mandatory = $true, HelpMessage = "The tag name to parse for the version hash.")]
        [string]$TagName
    )

    $thisVersionHash = [ordered]@{
        Major = $null
        Minor = $null
        Patch = $null
        Prerelease = $null
        Ahead = $null
        Hash = $null
        IsValid = $false
    }

    if (-not $TagName) {
        Write-Host "No tag name provided. Please specify a tag name."
        return [pscustomobject]$thisVersionHash
    }

    $baseTag = $TagName
    $ahead = $null
    $hash = $null

    if ($TagName -match '^(?<base>.+?)(?:-(?<ahead>\d+)-g(?<hash>[0-9a-fA-F]+))?$') {
        $baseTag = $Matches.base
        $ahead = $Matches.ahead
        $hash = $Matches.hash
    }

    if ($baseTag -match '^v?(?<major>\d+)\.(?<minor>\d+)\.(?<patch>\d+)(?:-(?<prerelease>[0-9A-Za-z]+(?:[.-][0-9A-Za-z]+)*))?$') {
        $thisVersionHash["Major"] = $Matches.major
        $thisVersionHash["Minor"] = $Matches.minor
        $thisVersionHash["Patch"] = $Matches.patch
        $thisVersionHash["Prerelease"] = $Matches.prerelease
        $thisVersionHash["Ahead"] = $ahead
        $thisVersionHash["Hash"] = $hash
        $thisVersionHash["IsValid"] = $true
    }

    return [pscustomobject]$thisVersionHash
}

function Get-BuildVersionInfo {
    <#
    .SYNOPSIS
        Converts a parsed Git tag into semantic and numeric build versions.
    .DESCRIPTION
        This function maps release and prerelease tags to the semantic version used by the app
        and to the four-part numeric file version required by NSIS.
    .PARAMETER VersionHash
        The parsed version object returned by Get-VersionHash.
    #>
    param (
        [Parameter(Mandatory = $true)]
        [psobject]$VersionHash
    )

    $version = "$($VersionHash.Major).$($VersionHash.Minor).$($VersionHash.Patch)"
    $fileVersionSuffix = 0
    $releaseClasses = @("alpha", "beta", "rc", "patch")

    if (-not [string]::IsNullOrWhiteSpace($VersionHash.Prerelease)) {
        $version += "-$($VersionHash.Prerelease)"
        if ($VersionHash.Prerelease -match '^(?<stage>alpha|beta|rc|patch)(?<number>\d+)?$') {
            $stageName = $Matches.stage
            $stageNumber = if ([string]::IsNullOrWhiteSpace($Matches.number)) { 0 } else { [int]$Matches.number }
            $stageIndex = $releaseClasses.IndexOf($stageName) + 1
            $fileVersionSuffix = ($stageIndex * 10000) + ($stageNumber * 100)
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($VersionHash.Ahead)) {
        $version += "+$($VersionHash.Ahead)"
        $fileVersionSuffix += [int]$VersionHash.Ahead
    }

    return [pscustomobject]@{
        Version = $version
        FileVersion = "$($VersionHash.Major).$($VersionHash.Minor).$($VersionHash.Patch).$fileVersionSuffix"
    }
}

function Get-MostRecentTag {
    <#
    .SYNOPSIS
        Retrieves the most recent Git tag from the repository.
    .DESCRIPTION
        This function uses Git commands to find and return the most recent tag
        in the current repository.
    #>
    $currentTag = git describe --tags --abbrev=0 2>$null
    if ($currentTag) {
        # Write-Host "Most recent tag: $currentTag"
        return $currentTag
    } else {
        Write-Host -ForegroundColor Yellow "No tags found in the repository."
        return $null
    }
}

function Confirm-RepositoryIsClean {
    <#
    .SYNOPSIS
        Checks if the Git repository is clean (no uncommitted changes).
    .DESCRIPTION
        This function checks the status of the Git repository and determines if there are any
        uncommitted changes. It returns $true if the repository is clean, and $false otherwise.
    #>
    param(
        [Parameter(Mandatory = $false)]
        [Alias("i","ignore")]
        [switch]$IgnoreFileList,
        [Parameter(Mandatory = $false)]
        [Alias("q","s","silent")]
        [switch]$Quiet
    )
    $status = git status --porcelain=v1
    if ($status) {
        # $isClean = $true
        $statusList = $status -split "`n"
        $tmpStatusList = @()
        foreach ($line in $statusList) {
            if ( -not ($line -match $ScriptName) ) {
                # $isClean = $false
                $tmpStatusList += $line
            }
        }

        $newStatusList = @()
        if ($IgnoreFileList) {
            foreach ($line in $tmpStatusList) {
                $found = $false
                foreach ($file in $FileList) {
                    $fileRE = [regex]::Escape($file)
                    # Write-StdErr "Checking line: $line ($fileRE)"
                    if ($line -match "^\s*[A-Z]\s+$fileRE") {
                        $found = $true
                        # git add "$file" 2>&1 | Out-Null
                        break
                    }
                }
                if (-not $found) {
                    $newStatusList += $line
                }
            }
        } else {
            $newStatusList += $tmpStatusList
        }

        if ($newStatusList.Count -eq 0) { return $true }

        if (-not $Quiet) {
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
        }
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
    Push-Location $ProjectRoot -StackName "restoreproject"
    Write-Host -ForegroundColor Yellow "Restoring repository to a clean state..."
    foreach ($file in $FileList) {
        if ( (git status --porcelain=v1 | Select-String "$file") ) {
            Write-Host -NoNewLine -ForegroundColor Green "  Restoring: " ; Write-Host -ForegroundColor Yellow "$file"
            git restore "$file" 2>&1 | Out-Null
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
        [Parameter(Mandatory = $true)]
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
        [Parameter(Mandatory = $true)]
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

function Update-PackageJSON {
    <#
    .SYNOPSIS
        Updates the package.json file with the specified version.
    .DESCRIPTION
        This function modifies the package.json file to set the correct version.
    .PARAMETER PackageJsonPath
        The path to the package.json file to update.
    .PARAMETER Version
        The new version to set in the package.json file.
    #>
    param (
        [Parameter(Mandatory = $true)]
        [string]$PackageJsonPath,
        [Parameter(Mandatory = $true)]
        [string]$Version
    )

    if (-not (Test-Path -Path $PackageJsonPath)) {
        Write-Host -ForegroundColor Red "package.json file not found at path: $PackageJsonPath"
        return
    }

    Write-Host -ForegroundColor Cyan "Updating package.json with version value: $Version"
    $PackageData = Get-JsonContent -Path $PackageJsonPath

    if (-not $PackageData) {
        Write-Host -ForegroundColor Red "  Failed to read package.json or it is empty."
        return
    }
    if ($PackageData.version -ne $Version) {
        $PackageData.version = $Version
        Set-JsonContent -Path $PackageJsonPath -Value $PackageData
        Write-Host -ForegroundColor Green "  Updating version to: $Version"
    } else {
        Write-Host -ForegroundColor Yellow "  No changes made to package.json, version is already set to: $Version"
    }
}

function Update-InfoJSON {
    <#
    .SYNOPSIS
        Updates the info.json file with the specified version.
    .DESCRIPTION
        This function modifies the info.json file to set the correct version.
    .PARAMETER InfoJsonPath
        The path to the info.json file to update.
    .PARAMETER Version
        The new version to set in the info.json file.
    #>
    param (
        [Parameter(Mandatory = $true)]
        [string]$InfoJsonPath,
        [Parameter(Mandatory = $true)]
        [string]$FileVersion
    )

    if (-not (Test-Path -Path $InfoJsonPath)) {
        Write-Host -ForegroundColor Red "info.json file not found at path: $InfoJsonPath"
        return
    }

    Write-Host -ForegroundColor Cyan "Updating info.json:"
    $VIData = Get-JsonContent -Path $InfoJsonPath

    if (-not $VIData) {
        Write-Host -ForegroundColor Red "  Failed to read info.json or it is empty."
        return
    }
    $VIData.fixed.file_version = $FileVersion
    if (Set-JsonContent -Path $InfoJsonPath -Value $VIData) {
        Write-Host -ForegroundColor Green "  Updated info.json"
    }
}

function New-FileHashes {
    param(
        [Parameter(Mandatory = $false, HelpMessage = "The directory path to create file hashes in.")]
        [string]$DirectoryPath = "./build/bin",
        [Parameter(Mandatory = $false, HelpMessage = "The file pattern to match.")]
        [string]$FilePattern = "*.exe"
    )
    Push-Location $DirectoryPath -StackName "NewFileHashes"
    Write-Host -ForegroundColor Cyan "Writing sha256 and sha1 hashes for executable files..."
    foreach ($file in Get-ChildItem -Path $DirectoryPath -Filter $FilePattern -File -ErrorAction SilentlyContinue) {
        $sha256Name = $file.Name + ".sha256"
        $sha1Name = $file.Name + ".sha1"
        $sha256Hash = (Get-FileHash $file -Algorithm SHA256).Hash # | ForEach-Object { $_.Hash } | Out-File -FilePath $sha256Name -Encoding utf8
        $sha1Hash = (Get-FileHash $file -Algorithm SHA1).Hash # | ForEach-Object { $_.Hash } | Out-File -FilePath $sha1Name -Encoding utf8
        "$sha256Hash  $($file.Name)" | Out-File -FilePath $sha256Name -Encoding utf8
        "$sha1Hash  $($file.Name)" | Out-File -FilePath $sha1Name -Encoding utf8
    }
    Pop-Location -StackName "NewFileHashes"
}

function Invoke-AllTests {
    $allTestsScript = Join-Path $ScriptRoot "Run-AllTests.ps1"

    if (-not (Test-Path -Path $allTestsScript -PathType Leaf)) {
        Write-Host -ForegroundColor Red "Could not find test runner script: $allTestsScript"
        return $false
    }

    $allTestArgs = @("-Silent")

    if ($RunAllTests) {
        Write-Host -ForegroundColor Yellow "Running ALL tests (please wait)..."
        $allTestArgs += "-RunAllTests"
    } else {
        Write-Host -ForegroundColor Yellow "Running tests (please wait)..."
    }

    & $allTestsScript @allTestArgs
    $testExitCode = $LASTEXITCODE

    if ($testExitCode -eq 0) {
        Write-Host -ForegroundColor Green "  --[ALL TESTS PASSED]--"
        return $true
    }

    Write-Host -ForegroundColor Red "  **[FAILED]**"
    return $false
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
    $ProjectNSI = Join-Path $ProjectRoot "build" "windows" "installer" "project.nsi"
    $WailsJsonPath = Join-Path $ProjectRoot "wails.json"
    $PackageJsonPath = Join-Path $ProjectRoot "frontend" "package.json"
    $InfoJsonPath = Join-Path $ProjectRoot "build" "windows" "info.json"

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
    $Commit = $(git rev-parse --short HEAD)
    $CurrentCommitTag = $(git describe --tags HEAD 2> $null)
    $CurrentRepoTag = Get-MostRecentTag
    $CurrentCommitVersion = Get-VersionHash -TagName $CurrentCommitTag
    $CurrentRepoVersion = Get-VersionHash -TagName $CurrentRepoTag

    if ($CurrentCommitVersion.IsValid) {
        $BuildVersionInfo = Get-BuildVersionInfo -VersionHash $CurrentCommitVersion
        $Version = $BuildVersionInfo.Version
        $FileVersion = $BuildVersionInfo.FileVersion
    } elseif ($CurrentRepoVersion.IsValid) {
        $BuildVersionInfo = Get-BuildVersionInfo -VersionHash $CurrentRepoVersion
        $Version = $BuildVersionInfo.Version
        $FileVersion = $BuildVersionInfo.FileVersion
    } else {
        $Version = "0.0.0-dev+${Commit}"
        $ds = Get-DateStamp
        $FileVersion = "0.0.0.${ds}"
    }

    $isRepoClean = Confirm-RepositoryIsClean -Quiet -IgnoreFileList

    if ($ShowVersionOnly) {
        $Build = $false
        Write-Host -ForegroundColor Cyan "`nCurrent Repository Information:"
        Write-Host -NoNewline -ForegroundColor Yellow "  Most Recent Repo Tag: " ; Write-Host -ForegroundColor Green "$CurrentRepoTag"
        Write-Host -NoNewline -ForegroundColor Yellow "   Current Repo Status: "
        if ($isRepoClean) {
            Write-Host -ForegroundColor Green "Clean (Safe to build)"
        } else {
            Write-Host -ForegroundColor Red "Dirty (Uncommitted changes found)"
        }
        Write-Host -ForegroundColor Cyan "`nVersion Values For Next Build:"
        Write-Host -NoNewLine -ForegroundColor Yellow "      Semantic Version: " ; Write-Host -ForegroundColor Green "$Version"
        Write-Host -NoNewLine -ForegroundColor Yellow "       Numeric Version: " ; Write-Host -ForegroundColor Green "$FileVersion"
        return
    }

    if (-not (Confirm-RepositoryIsClean -IgnoreFileList)) {
        return
    }

    if ($Build) {
        if (-not (Invoke-AllTests)) {
            return
        }

        if ($NSIS) {
            Update-ProjectNSI -ProjectNSIPath $ProjectNSI -FileVersion $FileVersion
        }
        Update-WailsJSON -WailsJsonPath $WailsJsonPath -Version $Version
        Update-PackageJSON -PackageJsonPath $PackageJsonPath -Version $Version
        Update-InfoJSON -InfoJsonPath $InfoJsonPath -FileVersion $FileVersion

        Write-Host -NoNewLine -ForegroundColor Cyan "Building Wails application with version value: "
        Write-Host -NoNewLine -ForegroundColor Black -BackgroundColor White "$Version"
        Write-Host "`n`n$('=' * 78)"
        wails build -clean -ldflags "-X main.Version=${Version} -X main.Date=${Date} -X main.Commit=${Commit} -s -w" ${NSISOption} ${UPXOption}
        Write-Host "$('=' * 78)`n"

        # Create SHA1 and SHA256 hashes for the executable files
        New-FileHashes -DirectoryPath (Join-Path $ProjectRoot "build" "bin") -FilePattern "*.exe"

        Restore-RepositoryToCleanState
    }
}

Push-Location $ProjectRoot -StackName "ProjectRoot"
Invoke-WailsBuild
Pop-Location -StackName "ProjectRoot"