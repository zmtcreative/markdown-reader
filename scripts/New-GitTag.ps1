<#
    .SYNOPSIS
        Creates a new Git tag with the specified name and message, or increments the
        version based on the current tag.
    .DESCRIPTION
        This script allows you to create a new Git tag with a specified name and
        message. It can also increment the version based on the current tag,
        allowing you to set alpha, beta, or release candidate tags, or increment
        the patch, minor, or major version numbers.

        It updates the NSIS project file and Wails JSON file with the new version
        information, commits and pushes the changes to the repository, then creates
        the new tag and pushes it to the remote repository.
    .PARAMETER TagName
        The name of the new Git tag to create. If not specified, it will prompt for
        the tag name interactively.
    .PARAMETER Message
        The message for the new Git tag. If not specified, it will use the tag name as the message.
    .PARAMETER IncrementPatch
        Increment the patch version number.
    .PARAMETER IncrementMinor
        Increment the minor version number.
    .PARAMETER IncrementMajor
        Increment the major version number.
    .PARAMETER IncrementPrerelease
        Increment the existing prerelease version number.
    .PARAMETER SetAlpha
        Set the tag to alpha version.
    .PARAMETER SetBeta
        Set the tag to beta version.
    .PARAMETER SetReleaseCandidate
        Set the tag to rc (release candidate) version.
    .PARAMETER ReleaseVersion
        This is a release version, which removes the prerelease suffix but does not increment the version.
#>

#Requires -Version 7.0

[CmdletBinding(DefaultParameterSetName = "None")]
param (
    [Parameter(Mandatory = $false, HelpMessage = "The new Git tag to create.")]
    [string]$TagName = "",
    [Parameter(Mandatory = $false, HelpMessage = "The message for the new Git tag.")]
    [string]$Message = "",
    [Parameter(Mandatory = $false, ParameterSetName = "SetAlpha", HelpMessage = "Set the tag to alpha.")]
    [Parameter(Mandatory = $false, ParameterSetName = "SetBeta", HelpMessage = "Set the tag to beta.")]
    [Parameter(Mandatory = $false, ParameterSetName = "SetReleaseCandidate", HelpMessage = "Set the tag to release candidate.")]
    [Parameter(Mandatory = $false, ParameterSetName = "IncPatch", HelpMessage = "Increment the patch version.")]
    [switch]$IncrementPatch,
    [Parameter(Mandatory = $false, ParameterSetName = "SetAlpha", HelpMessage = "Set the tag to alpha.")]
    [Parameter(Mandatory = $false, ParameterSetName = "SetBeta", HelpMessage = "Set the tag to beta.")]
    [Parameter(Mandatory = $false, ParameterSetName = "SetReleaseCandidate", HelpMessage = "Set the tag to release candidate.")]
    [Parameter(Mandatory = $false, ParameterSetName = "IncMinor", HelpMessage = "Increment the minor version.")]
    [switch]$IncrementMinor,
    [Parameter(Mandatory = $false, ParameterSetName = "SetAlpha", HelpMessage = "Set the tag to alpha.")]
    [Parameter(Mandatory = $false, ParameterSetName = "SetBeta", HelpMessage = "Set the tag to beta.")]
    [Parameter(Mandatory = $false, ParameterSetName = "SetReleaseCandidate", HelpMessage = "Set the tag to release candidate.")]
    [Parameter(Mandatory = $false, ParameterSetName = "IncMajor", HelpMessage = "Increment the major version.")]
    [switch]$IncrementMajor,
    [Parameter(Mandatory = $false, ParameterSetName = "IncPrerelease", HelpMessage = "Increment the existing prerelease version.")]
    [switch]$IncrementPrerelease,
    [Parameter(Mandatory = $false, ParameterSetName = "SetAlpha", HelpMessage = "Set the tag to alpha.")]
    [switch]$SetAlpha,
    [Parameter(Mandatory = $false, ParameterSetName = "SetBeta", HelpMessage = "Set the tag to beta.")]
    [switch]$SetBeta,
    [Parameter(Mandatory = $false, ParameterSetName = "SetReleaseCandidate", HelpMessage = "Set the tag to release candidate.")]
    [switch]$SetReleaseCandidate,
    [Parameter(Mandatory = $false, ParameterSetName = "ReleaseVersion", HelpMessage = "This is a release version (remove prerelease suffix but no increments).")]
    [switch]$ReleaseVersion,
    [Parameter(Mandatory = $false, ParameterSetName = "ShowCurrentVersion", HelpMessage = "Show the current version without making changes.")]
    [Alias("show", "current","version")]
    [switch]$ShowCurrentVersion
)

# Set-Location $PSScriptRoot

# Set up script and project paths
$ScriptFullName = $MyInvocation.MyCommand.Path
$ScriptRoot = Split-Path -Parent $ScriptFullName
$ScriptName = Split-Path -Leaf $ScriptFullName
if ($ScriptRoot -match '[\\/]scripts[\\/]?$') {
    $ScriptRelativePath = 'scripts/' + $ScriptName ; $ScriptRelativePath | Out-Null
    $tmpProjectRoot = $ScriptRoot -replace '[\\/]scripts[\\/]?', ''
} else {
    $tmpProjectRoot = $ScriptRoot
}
if (Test-Path -Path "$tmpProjectRoot\go.mod") {
    $ProjectRoot = $tmpProjectRoot
} else {
    Write-Host -ForegroundColor Red "Could not find go.mod in the expected project root: $tmpProjectRoot"
    exit 1
}

# List of files modified as part of this tag creation process
$FileList = @(
    "wails.json",
    "frontend/package.json",
    "build/windows/info.json",
    "build/windows/installer/project.nsi"
)

function Write-StdErr {
    <#
    .SYNOPSIS
    Writes text to stderr when running in a regular console window,
    to the host''s error stream otherwise.

    .DESCRIPTION
    Writing to true stderr allows you to write a well-behaved CLI
    as a PS script that can be invoked from a batch file, for instance.

    Note that PS by default sends ALL its streams to *stdout* when invoked from
    cmd.exe.

    This function acts similarly to Write-Host in that it simply calls
    .ToString() on its input; to get the default output format, invoke
    it via a pipeline and precede with Out-String.

    #>
    param (
        [Parameter(Mandatory = $false, Position = 0)]
        [PSObject]$InputObject
    )
    $outFunc = if ($Host.Name -eq 'ConsoleHost') {
        [Console]::Error.WriteLine
    } else {
        $host.ui.WriteErrorLine
    }
    if ($InputObject) {
        [void] $outFunc.Invoke($InputObject.ToString())
    } else {
        [string[]] $lines = @()
        $Input | ForEach-Object { $lines += $_.ToString() }
        [void] $outFunc.Invoke($lines -join "`r`n")
    }
}

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

function Update-ProjectNSI {
    <#
    .SYNOPSIS
        Updates the NSI project file with the new version information.
    .DESCRIPTION
        This function modifies the specified NSI project file to reflect the new version
        information based on the provided tag name.
    .PARAMETER ProjectNSIPath
        The path to the NSI project file to update.
    .PARAMETER TagName
        The tag name to use for the version update.
    #>
    param (
        [Parameter(Mandatory = $true)]
        [string]$ProjectNSIPath,
        [string]$TagName
    )

    if (-not (Test-Path -Path $ProjectNSIPath)) {
        Write-Host -ForegroundColor Red "Project NSI file not found at path: $ProjectNSIPath"
        return
    }

    Write-Host -ForegroundColor Cyan "Updating NSI project file: $ProjectNSIPath"

    $tmpVersionHash = Get-VersionHash -TagName $TagName
    $BuildVersionInfo = Get-BuildVersionInfo -VersionHash $tmpVersionHash
    $FileVersion = $BuildVersionInfo.FileVersion

    $NSIData = Get-Content -Path $ProjectNSIPath
    $NewNSIData = @()
    $FileChanged = $false

    foreach ($line in $NSIData) {
        if ($line -match '^(?<key>VIFileVersion\s+)"(?<value>[^"]+)"\s*$') {
            if ($Matches.value -ne $FileVersion) {
                $newval1 = $Matches.key + '"' + $FileVersion + '"'
                $line = $line -replace '^(VIFileVersion\s+)"([^"]+)"\s*$', $newval1
                Write-Host -ForegroundColor Green "  Updating VIFileVersion in NSIS project file"
                $FileChanged = $true
            }
        }
        if ($line -match '^(?<key>VIProductVersion\s+)"(?<value>[^"]+)"\s*$') {
            if ($Matches.value -ne $FileVersion) {
                Write-Host -ForegroundColor Green "  Updating VIProductVersion in NSIS project file"
                $newval2 = $Matches.key + '"' + $FileVersion + '"'
                $line = $line -replace '^(VIProductVersion\s+)"([^"]+)"\s*$', $newval2
                $FileChanged = $true
            }
        }
        if ($line -match '^(?<key>VIAddVersionKey\s+"FileVersion"\s+)"(?<value>[^"]+)"\s*$') {
            if ($Matches.value -ne $FileVersion) {
                Write-Host -ForegroundColor Green "  Updating VIAddVersionKey FileVersion in NSIS project file"
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
        Updates the Wails JSON file with the new version information.
    .DESCRIPTION
        This function modifies the specified Wails JSON file to reflect the new version
        information based on the provided tag name.
    .PARAMETER WailsJsonPath
        The path to the Wails JSON file to update.
    .PARAMETER TagName
        The tag name to use for the version update.
    #>
    param (
        [Parameter(Mandatory = $true)]
        [string]$WailsJsonPath,
        [string]$TagName
    )

    if (-not (Test-Path -Path $WailsJsonPath)) {
        Write-Host -ForegroundColor Red "wails.json file not found at path: $WailsJsonPath"
        return
    }

    $tmpVersionHash = Get-VersionHash -TagName $TagName
    $BuildVersionInfo = Get-BuildVersionInfo -VersionHash $tmpVersionHash
    $Version = $BuildVersionInfo.Version

    Write-Host -ForegroundColor Cyan "Updating wails.json with version value: $Version"
    $WailsData = Get-JsonContent -Path $WailsJsonPath

    if (-not $WailsData) {
        Write-Host -ForegroundColor Red "  Failed to read wails.json or it is empty."
        return
    }
    if ($WailsData.Info.productVersion -ne $Version) {
        $WailsData.Info.productVersion = $Version
        Set-JsonContent -Path $WailsJsonPath -Value $WailsData
        Write-Host -ForegroundColor Green "  Version changed to: $Version"
    } else {
        Write-Host -ForegroundColor Yellow "  No changes made to wails.json, version is already set to: $Version"
    }
}

function Update-PackageJSON {
    <#
    .SYNOPSIS
        Updates the package.json file with the new version information.
    .DESCRIPTION
        This function modifies the specified package.json file to reflect the new version
        information based on the provided tag name.
    .PARAMETER PackageJsonPath
        The path to the package.json file to update.
    .PARAMETER TagName
        The tag name to use for the version update.
    #>
    param (
        [Parameter(Mandatory = $true)]
        [string]$PackageJsonPath,
        [string]$TagName
    )

    if (-not (Test-Path -Path $PackageJsonPath)) {
        Write-Host -ForegroundColor Red "package.json file not found at path: $PackageJsonPath"
        return
    }

    $tmpVersionHash = Get-VersionHash -TagName $TagName
    $BuildVersionInfo = Get-BuildVersionInfo -VersionHash $tmpVersionHash
    $Version = $BuildVersionInfo.Version

    Write-Host -ForegroundColor Cyan "Updating package.json with version value: $Version"
    $PackageData = Get-JsonContent -Path $PackageJsonPath

    if (-not $PackageData) {
        Write-Host -ForegroundColor Red "  Failed to read package.json or it is empty."
        return
    }
    if ($PackageData.version -ne $Version) {
        $PackageData.version = $Version
        Set-JsonContent -Path $PackageJsonPath -Value $PackageData
        Write-Host -ForegroundColor Green "  Version changed to: $Version"
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
        [string]$TagName
    )

    if (-not (Test-Path -Path $InfoJsonPath)) {
        Write-Host -ForegroundColor Red "info.json file not found at path: $InfoJsonPath"
        return
    }

    $tmpVersionHash = Get-VersionHash -TagName $TagName
    $BuildVersionInfo = Get-BuildVersionInfo -VersionHash $tmpVersionHash
    $FileVersion = $BuildVersionInfo.FileVersion

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

function Set-NewTag {
    <#
    .SYNOPSIS
        Creates a new Git tag with the specified name and message.
    .DESCRIPTION
        This function creates a new Git tag in the local repository and pushes it to the remote repository.
    .PARAMETER TagName
        The name of the new Git tag.
    .PARAMETER Message
        The message for the new Git tag.
    #>
    param (
        [Parameter(Mandatory = $true, HelpMessage = "The name of the new Git tag.")]
        [string]$TagName,
        [Parameter(Mandatory = $true, HelpMessage = "The message for the new Git tag.")]
        [string]$Message
    )

    if (-not $TagName) {
        Write-Host -ForegroundColor Red "[Set-NewTag()] No tag name provided. Please specify a tag name."
        return
    }

    # Check if the tag already exists
    $existingTags = git tag
    if ($existingTags -contains $TagName) {
        Write-Host -ForegroundColor Yellow "   Tag '$TagName' already exists. Please choose a different tag name."
        return
    } else {
        Write-Host -NoNewLine -ForegroundColor Cyan "Creating New Tag: " ; Write-Host -NoNewLine -ForegroundColor Yellow "$TagName"
        Write-Host -NoNewLine -ForegroundColor Cyan " -- Message: " ; Write-Host -ForegroundColor Yellow "$Message"
        $gitTagResults = (git tag -a "$TagName" -m "$Message" 2>&1)
        if (! $?) {
            Write-Host -ForegroundColor Red "   Failed to create tag '$TagName'.`n   Please check the repository status."
            $gitTagResults | ForEach-Object { Write-Host -ForegroundColor Yellow "   $_" }
            return
        }
        git push origin "$TagName" 2>&1 | Out-Null
        if (! $?) {
            Write-Host -ForegroundColor Red "  Failed to push tag '$TagName' to remote repository.`n   Please check the repository status."
            return
        }
        Write-Host -ForegroundColor Green "  Tag '$TagName' created and pushed to remote repository."
        git push 2>&1 | Out-Null
    }
}

function Get-NewTagNamePrompt {
    <#
    .SYNOPSIS
        Prompts the user to enter a new tag name.
    .DESCRIPTION
        This function allows the user to specify a new tag name, with the option to use a suggested tag name.
    .PARAMETER SuggestedTagName
        The current tag name.
    #>
    param (
        [Parameter(Mandatory = $true, HelpMessage = "The current tag name.")]
        [string]$SuggestedTagName
    )

    if (-not $SuggestedTagName) {
        Write-Host -ForegroundColor Red "No suggested tag provided. Please specify a tag name."
        return $null
    }

    while ($true) {
        Write-Host -ForegroundColor Cyan "Enter a new tag name (or press ENTER to use the suggested tag)"
        Write-Host -NoNewline '  [' ; Write-Host -NoNewline -ForegroundColor Yellow 'Tag: '
        Write-Host -NoNewline -ForegroundColor Green "$SuggestedTagName"
        $response = Read-Host "]"
        if ([string]::IsNullOrWhiteSpace($response)) {
            $NewTagName = $SuggestedTagName
        } elseif ($response -imatch '^\s*Q') {
            return $null
        } elseif ($response -imatch '^\s*(?:y|yes|n|no)\s*$') {
            Write-Host -ForegroundColor Red "  Oops! This is not a yes/no prompt! Try again..."
            Start-Sleep 2
            continue
        } else {
            $NewTagName = $response
        }

        Write-Host -NoNewLine -ForegroundColor Cyan "  You entered: "
        Write-Host -ForegroundColor Yellow "$NewTagName"
        Write-Host -NoNewline -ForegroundColor Cyan "  Is this correct? "
        $verify = Read-Host "(y/N/q)"
        if ($verify -imatch '^\s*y(?:es)?\s*$') {
            return $NewTagName
        } elseif ($verify -imatch '^\s*q(?:uit)?\s*$') {
            return $null
        }
    }
}

function Get-NewMessagePrompt {
    <#
    .SYNOPSIS
        Prompts the user to confirm the current message or enter a new one.
    .DESCRIPTION
        This function allows the user to specify a new message, with the option to use a suggested message.
    .PARAMETER SuggestedMessage
        The current message.
    #>
    param (
        [Parameter(Mandatory = $true, HelpMessage = "The current message.")]
        [string]$SuggestedMessage
    )

    if (-not $SuggestedMessage) {
        Write-Host -ForegroundColor Red "No suggested message provided. Please specify a message."
        return $null
    }

    while ($true) {
        Write-Host -ForegroundColor Cyan "Enter a new message (or press ENTER to use the suggested message)"
        Write-Host -NoNewline '  [' ; Write-Host -NoNewline -ForegroundColor Yellow 'Message: '
        Write-Host -NoNewline -ForegroundColor Green "$SuggestedMessage"
        $response = Read-Host "]"
        if ([string]::IsNullOrWhiteSpace($response)) {
            $NewMessage = $SuggestedMessage
        } elseif ($response -imatch '^\s*Q') {
            return $null
        } elseif ($response -imatch '^\s*(?:y|yes|n|no)\s*$') {
            Write-Host -ForegroundColor Red "  Oops! This is not a yes/no prompt! Try again..."
            Start-Sleep 2
            continue
        } else {
            $NewMessage = $response
        }

        Write-Host -NoNewLine -ForegroundColor Green "  You entered: "
        Write-Host -ForegroundColor Yellow "$NewMessage"
        Write-Host -NoNewline -ForegroundColor Cyan "  Is this correct? "
        $verify = Read-Host "(y/N/q)"
        if ($verify -imatch '^\s*y(?:es)?\s*$') {
            return $NewMessage
        } elseif ($verify -imatch '^\s*q(?:uit)?\s*$') {
            return $null
        }
    }
}

function Get-NextTagName {
    <#
    .SYNOPSIS
        Retrieves the next tag name based on the current tag name hash.
    .DESCRIPTION
        This function calculates the next tag name by incrementing the version components
        based on the specified parameters.
    .PARAMETER VersionHash
        The current tag name hash.
    #>
    param(
        [Parameter(Mandatory = $true, HelpMessage = "The current tag name hash.")]
        [object]$VersionHash
    )

    $tmpVersionHash = [pscustomobject]@{
        Major = $VersionHash.Major
        Minor = $VersionHash.Minor
        Patch = $VersionHash.Patch
        Prerelease = $VersionHash.Prerelease
        Ahead = $VersionHash.Ahead
        Hash = $VersionHash.Hash
        IsValid = $VersionHash.IsValid
    }
    $nextTagName = ""
    $prstatus = ""
    $prnumber = $null
    $paramsUsed = $false

    if ($IncrementPatch -or $IncrementMinor -or $IncrementMajor -or $IncrementPrerelease) {
        $paramsUsed = $true
    }
    elseif ($SetAlpha -or $SetBeta -or $SetReleaseCandidate) {
        $paramsUsed = $true
    }
    elseif ($ReleaseVersion) {
        $paramsUsed = $true
    }

    if ($SetAlpha) {
        $prstatus = "alpha"
        $prnumber = 1
    } elseif ($SetBeta) {
        $prstatus = "beta"
        $prnumber = 1
    } elseif ($SetReleaseCandidate) {
        $prstatus = "rc"
        $prnumber = 1
    }

    if ($ReleaseVersion) {
        $tmpVersionHash.Prerelease = ""
        $tmpVersionHash.Ahead = ""
        $tmpVersionHash.Hash = ""
        $prstatus = ""
        $prnumber = $null
        return "v$($tmpVersionHash.Major).$($tmpVersionHash.Minor).$($tmpVersionHash.Patch)"
    }

    if ($IncrementPatch -or $IncrementMinor -or $IncrementMajor) {
        if (-not [string]::IsNullOrWhiteSpace($tmpVersionHash.Prerelease)) {
            $tmpVersionHash.Prerelease = ""
            $tmpVersionHash.Ahead = ""
            $tmpVersionHash.Hash = ""
            if ( -not ($SetAlpha -or $SetBeta -or $SetReleaseCandidate) ) {
                $tmpVersionHash.Prerelease = ""
            }
        }
    }
    if ($IncrementPatch) {
        $tmpVersionHash.Patch = [int]$tmpVersionHash.Patch + 1
    }
    elseif ($IncrementMinor) {
        $tmpVersionHash.Minor = [int]$tmpVersionHash.Minor + 1
        $tmpVersionHash.Patch = 0
    }
    elseif ($IncrementMajor) {
        $tmpVersionHash.Major = [int]$tmpVersionHash.Major + 1
        $tmpVersionHash.Minor = 0
        $tmpVersionHash.Patch = 0
    }
    elseif ($IncrementPrerelease) {
        if (-not [string]::IsNullOrWhiteSpace($tmpVersionHash.Prerelease)) {
            if ($tmpVersionHash.Prerelease -match '^(?<status>[a-zA-Z]+)?(?<number>\d+)?$') {
                $prstatus = $Matches.status
                $prnumber = [int]$Matches.number + 1
            }
        } else {
            $prstatus = "alpha"
            $prnumber = 1
        }
        $tmpVersionHash.Prerelease = "${prstatus}${prnumber}"
    }
    if (-not [string]::IsNullOrWhiteSpace($prstatus)) {
        $tmpVersionHash.Prerelease = "${prstatus}${prnumber}"
        $nextTagName = "v$($tmpVersionHash.Major).$($tmpVersionHash.Minor).$($tmpVersionHash.Patch)-$($tmpVersionHash.Prerelease)"
    }
    elseif ($paramsUsed) {
        $nextTagName = "v$($tmpVersionHash.Major).$($tmpVersionHash.Minor).$($tmpVersionHash.Patch)"
    }
    else {
        if (-not [string]::IsNullOrWhiteSpace($tmpVersionHash.Prerelease)) {
            if ($tmpVersionHash.Prerelease -match '^(?<status>[a-zA-z]+)?(?<number>\d+)?$') {
                $prstatus = $Matches.status
                $prnumber = [int]$Matches.number + 1
                # Write-Host  "Incrementing prerelease version: ${prstatus}  ${prnumber}"
                $tmpVersionHash.Prerelease = "${prstatus}${prnumber}"
                $nextTagName = "v$($tmpVersionHash.Major).$($tmpVersionHash.Minor).$($tmpVersionHash.Patch)-$($tmpVersionHash.Prerelease)"
            }
        }
        else {
            $tmpVersionHash.Patch = [int]$tmpVersionHash.Patch + 1
            $nextTagName = "v$($tmpVersionHash.Major).$($tmpVersionHash.Minor).$($tmpVersionHash.Patch)"
        }
    }

    return $nextTagName
}

function Confirm-RepositoryIsClean {
    <#
    .SYNOPSIS
        Checks if the Git repository is clean (no uncommitted changes).
    .DESCRIPTION
        This function checks the status of the Git repository and returns $true if there are no uncommitted changes,
        otherwise it returns $false and displays a warning message.
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
            Write-Host "  - Commit your changes: git commit -a -m 'Your commit message'"
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

function Push-RepositoryCommit {
    <#
    .SYNOPSIS
        Commits and pushes the modified files in $FileList array.
    .DESCRIPTION
        This function commits the changes made to the files listed in the $FileList array
        in the Git repository. It checks if there are any changes to commit, and if so,
        it commits them with a message that includes the tag name.
    .PARAMETER TagName
        The name of the new Git tag.
    #>
    param(
        [Parameter(Mandatory = $true, HelpMessage = "The tag name to commit changes for.")]
        [string]$TagName,
        [Parameter(Mandatory = $false, HelpMessage = "The commit message for the changes.")]
        [string]$Message
    )
    Push-Location $ProjectRoot -StackName "commitproject"
    # Write-Host -ForegroundColor Yellow "Restoring repository to a clean state..."

    if ([string]::IsNullOrWhiteSpace($Message)) {
        $Message = "Commit project with tag $TagName"
    }
    $status = git status --porcelain=v1
    if ([string]::IsNullOrWhiteSpace($status)) {
        Write-Host -ForegroundColor Green "Repository is clean. No changes to commit."
        Pop-Location -StackName "commitproject"
        return $true
    }
    $statusList = $status -split "`n"
    $newStatusList = @()

    foreach ($line in $statusList) {
        $found = $false
        foreach ($file in $FileList) {
            $fileRE = [regex]::Escape($file)
            # Write-StdErr "Checking line: $line ($fileRE)"
            if ($line -match "^\s*[A-Z]\s+$fileRE") {
                $found = $true
                git add "$file" 2>&1 | Out-Null
                break
            }
        }
        if (-not $found) {
            $newStatusList += $line
        }
    }
    if ( ($newStatusList.Count -ne 0) -and (-not (Confirm-RepositoryIsClean -IgnoreFileList)) ) {
        # Confirm-RepositoryIsClean -IgnoreFileList
        Pop-Location -StackName "commitproject"
        return $false
    }
    Write-Host -ForegroundColor Cyan "Committing changes to repository with message: $Message"
    git commit -m "$Message" 2>&1 | Out-Null
    if (! $?) {
        Write-Host -ForegroundColor Red "  Failed to commit changes. Please check the repository status."
        Pop-Location -StackName "commitproject"
        return $false
    }
    else {
        Write-Host -ForegroundColor Green "  Changes committed successfully."
    }
    git push 2>&1 | Out-Null
    if (! $?) {
        Write-Host -ForegroundColor Red "  Failed to push changes. Please check the repository status."
        Pop-Location -StackName "commitproject"
        return $false
    }
    else {
        Write-Host -ForegroundColor Green "  Changes pushed successfully."
    }
    Pop-Location -StackName "commitproject"
    return $true
}

function Invoke-NewGitTag {
    <#
    .SYNOPSIS
        Creates a new Git tag.
    .DESCRIPTION
        This function creates a new Git tag in the local repository and pushes it to the remote repository.
    #>
    $currentTag = Get-MostRecentTag
    $newTagName = ""
    $ProjectNSI = Join-Path $ProjectRoot "build" "windows" "installer" "project.nsi"
    $WailsJsonPath = Join-Path $ProjectRoot "wails.json"
    $PackageJsonPath = Join-Path $ProjectRoot "frontend" "package.json"
    $InfoJsonPath = Join-Path $ProjectRoot "build" "windows" "info.json"

    if ($ShowCurrentVersion) {
        Write-Host ""
        if ($currentTag) {
            Write-Host -NoNewLine -ForegroundColor Cyan "Current (most recent) tag: "; Write-Host -ForegroundColor Green "$currentTag"
        } else {
            Write-Host -ForegroundColor Yellow "No current tag/version found."
        }
        Write-Host ""
        return
    }

    if (-not (Confirm-RepositoryIsClean -IgnoreFileList)) {
        return
    }

    if (-not $currentTag) {
        Write-Host -ForegroundColor Yellow "No current tag found."
        if (-not $TagName) {
            # Write-Host "No tag name provided."
            $tmpTagName = "v0.0.1-alpha1"
            $newTagName = Get-NewTagNamePrompt -SuggestedTagName $tmpTagName
        } else {
            $newTagName = Get-NewTagNamePrompt -SuggestedTagName $TagName
        }
    } else {
        $versionHash = Get-VersionHash -TagName $currentTag
        Write-Host -ForegroundColor Green "Current (most recent) tag: ${currentTag}"
        if (-not $TagName) {
            $tmpTagName = Get-NextTagName -VersionHash $versionHash
            $newTagName = Get-NewTagNamePrompt -SuggestedTagName $tmpTagName
        } else {
            $newTagName = Get-NewTagNamePrompt -SuggestedTagName $TagName
        }
    }

    if ($newTagName) {
        $verMessage = $newTagName -ireplace '^v', 'Version '
        if ([string]::IsNullOrWhiteSpace($Message)) {
            $Message = $verMessage
        } else {
            $Message = $Message + " (${newTagName})"
        }
        $tmpMessage = Get-NewMessagePrompt -SuggestedMessage $Message
        if ($null -eq $tmpMessage) {
            Write-Host -ForegroundColor Red "Quit requested. Exiting."
            return
        } elseif ($tmpMessage -ne $Message) {
            $re = [regex]::Escape($newTagName)
            if (-not ($tmpMessage -match $re)) {
                $Message = $tmpMessage + " (${newTagName})"
            } else {
                $Message = $tmpMessage
            }
        } else {
            $Message = $tmpMessage
        }
        Write-Host -ForegroundColor Cyan "New Tag and Message:"
        Write-Host -NoNewLine -ForegroundColor Green "          New Tag: " ; Write-Host -ForegroundColor Yellow "$newTagName"
        Write-Host -NoNewLine -ForegroundColor Green "  Current Message: " ; Write-Host -ForegroundColor Yellow "$Message"
        Write-Host -NoNewline -ForegroundColor Cyan "Do you want to proceed with this tag and message? "
        # Write-Host -NoNewline -ForegroundColor Yellow "(y/N)"
        $verify = Read-Host -Prompt "(y/N)"
        if ($verify -inotmatch '^\s*Y') {
            Write-Host -ForegroundColor Red "Operation cancelled by user."
            return
        }

        Update-ProjectNSI -ProjectNSIPath $ProjectNSI -TagName $newTagName
        Update-WailsJSON -WailsJsonPath $WailsJsonPath -TagName $newTagName
        Update-PackageJSON -PackageJsonPath $PackageJsonPath -TagName $newTagName
        Update-InfoJSON -InfoJsonPath $InfoJsonPath -TagName $newTagName

        if (-not (Push-RepositoryCommit -TagName $newTagName -Message $Message)) {
            Write-Host -ForegroundColor Red "Failed to commit changes before tagging. Exiting."
            return
        } else {
            Write-Host -NoNewLine -ForegroundColor Cyan "  New Tag Name: "
            Write-Host -ForegroundColor Yellow "$newTagName"
            Write-Host -NoNewLine -ForegroundColor Cyan "       Message: "
            Write-Host -ForegroundColor Yellow "$Message"
            Set-NewTag -TagName $newTagName -Message $Message
        }
    }
    elseif ($null -eq $newTagName) {
        Write-Host -ForegroundColor Red "Quit requested. Exiting."
        return
    }
    else {
        Write-Host -ForegroundColor Red "No new tag name provided. Exiting."
        return
    }
}

Push-Location $ProjectRoot -StackName "ProjectRoot"
Invoke-NewGitTag
Pop-Location -StackName "ProjectRoot"
