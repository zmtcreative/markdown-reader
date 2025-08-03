<#
#>

#Requires -Version 7.0

[CmdletBinding(DefaultParameterSetName = "None")]
param (
    [Parameter(Mandatory = $false, ParameterSetName = "Default", Position = 0, HelpMessage = "The new Git tag to create.")]
    [string]$TagName = "",
    [Parameter(Mandatory = $false, ParameterSetName = "Default", Position = 1, HelpMessage = "The message for the new Git tag.")]
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
    [Parameter(Mandatory = $false, ParameterSetName = "isRelease", HelpMessage = "This is a release version (remove prerelease suffix but no increments).")]
    [switch]$ReleaseVersion
)

Set-Location $PSScriptRoot

$ScriptName = $MyInvocation.MyCommand.Name

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

function Confirm-RepositoryIsClean {
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

function Update-ProjectNSI {
    param (
        [Parameter(Mandatory = $true)]
        [string]$ProjectNSIPath,
        [string]$TagName
    )

    if (-not (Test-Path -Path $ProjectNSIPath)) {
        Write-Host -ForegroundColor Red "Project NSI file not found at path: $ProjectNSIPath"
        return
    }

    Write-Host -ForegroundColor Green "Updating NSI project file: $ProjectNSIPath"

    $RC = @("alpha", "beta", "rc", "patch", "")
    $tmpVersionHash = Get-VersionHash -TagName $TagName
    $FileVersion = "$($tmpVersionHash.Major).$($tmpVersionHash.Minor).$($tmpVersionHash.Patch)"
    if (-not [string]::IsNullOrWhiteSpace($tmpVersionHash.Prerelease)) {
        if ($tmpVersionHash.Prerelease -match "^(?<rc>alpha|beta|rc)?(?<rcnum>\d+)?$") {
            $rcString = $Matches.rc
            $rcNumber = [int]$Matches.rcnum || 0
            $rcIDX = $RC.IndexOf($rcString) + 1
        }

        $PreReleaseNumber = ($rcIDX * 10000) + ($rcNumber * 100)
        # $tmpVersionHash.Prerelease = $PreReleaseNumber.ToString("D5")
        $FileVersion += ".${PreReleaseNumber}"
    }

    $NSIData = Get-Content -Path $ProjectNSIPath
    $NewNSIData = @()
    $FileChanged = $false

    foreach ($line in $NSIData) {
        if ($line -match '^(?<key>VIFileVersion\s+)"(?<value>[^"]+)"\s*$') {
            if (string($Matches.key) -ne $FileVersion) {
                $newval1 = $Matches.key + '"' + $FileVersion + '"'
                $line = $line -replace '^(VIFileVersion\s+)"([^"]+)"\s*$', $newval1
                Write-Host -ForegroundColor Yellow "  Updating VIFileVersion in NSIS project file"
                $FileChanged = $true
            }
        }
        if ($line -match '^(?<key>VIProductVersion\s+)"(?<value>[^"]+)"\s*$') {
            if (string($Matches.key) -ne $FileVersion) {
                Write-Host -ForegroundColor Yellow "  Updating VIProductVersion in NSIS project file"
                $newval2 = $Matches.key + '"' + $FileVersion + '"'
                $line = $line -replace '^(VIProductVersion\s+)"([^"]+)"\s*$', $newval2
                $FileChanged = $true
            }
        }
        if ($line -match '^(?<key>VIAddVersionKey\s+"FileVersion"\s+)"(?<value>[^"]+)"\s*$') {
            if (string($Matches.key) -ne $FileVersion) {
                Write-Host -ForegroundColor Yellow "  Updating VIAddVersionKey FileVersion in NSIS project file"
                $newval3 = $Matches.key + '"' + $FileVersion + '"'
                $line = $line -replace '^(VIAddVersionKey\s+"FileVersion"\s+)"([^"]+)"\s*$', $newval3
                $FileChanged = $true
            }
        }
        $NewNSIData += $line
    }
    if (-not $FileChanged) {
        Write-Host -ForegroundColor Green "  No changes made to NSIS project file"
        return
    }
    Set-Content -Path $ProjectNSIPath -Encoding utf8 -Value $NewNSIData
}

function Update-WailsJSON {
    param (
        [Parameter(Mandatory = $true)]
        [string]$WailsJsonPath,
        [string]$Version
    )

    if (-not (Test-Path -Path $WailsJsonPath)) {
        Write-Host -ForegroundColor Red "wails.json file not found at path: $WailsJsonPath"
        return
    }

    Write-Host -ForegroundColor Green "Updating wails.json with version value: $Version"
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

function Get-MostRecentTag {
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
    param (
        [Parameter(Mandatory = $true, HelpMessage = "The tag name to parse for the version hash.")]
        [string]$TagName
    )

    $thisVersionHash = @{}

    if (-not $TagName) {
        Write-Host "No tag name provided. Please specify a tag name."
        return
    }

    if ( $TagName -match "v?(?<major>\d+)\.(?<minor>\d+)\.(?<patch>\d+)(?:-(?<prerelease>(?:0|[1-9]\d*|\w+\d*)))(?:[.+-](?<ahead>\d+)(?:-g?(?<hash>[0-9a-fA-F]+))?)?$") {
        $thisVersionHash["Major"] = $Matches.major
        $thisVersionHash["Minor"] = $Matches.minor
        $thisVersionHash["Patch"] = $Matches.patch
        $thisVersionHash["Prerelease"] = $Matches.prerelease
        $thisVersionHash["Ahead"] = $Matches.ahead
        $thisVersionHash["Hash"] = $Matches.hash
    }

    return $thisVersionHash
}

function Set-NewTag {
    param (
        [Parameter(Mandatory = $true, HelpMessage = "The name of the new Git tag.")]
        [string]$TagName,
        [Parameter(Mandatory = $true, HelpMessage = "The message for the new Git tag.")]
        [string]$Message
    )

    if (-not $TagName) {
        Write-Host "No tag name provided. Please specify a tag name."
        return
    }

    # Check if the tag already exists
    $existingTags = git tag
    if ($existingTags -contains $TagName) {
        Write-Host "Tag '$TagName' already exists. Please choose a different tag name."
        return
    } else {
        Write-Host "Creating new tag: $TagName"
        try {
            git tag -a "$TagName" -m "$Message"
        } catch {
            Write-Host "Failed to create tag '$TagName'. Error: $_"
            return
        }
        try {
            git push origin "$TagName"
        } catch {
            Write-Host "Failed to push tag '$TagName' to remote repository. Error: $_"
            return
        }
        Write-Host "Tag '$TagName' created and pushed to remote repository."
    }
}

function Get-NewTagName {
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
        $response = Read-Host " [Tag: $SuggestedTagName]"
        if ([string]::IsNullOrWhiteSpace($response)) {
            $NewTagName = $SuggestedTagName
        } else {
            $NewTagName = $response
        }

        $verify = Read-Host "  You entered '$NewTagName'. Is this correct? (N/y/q)"
        if ($verify -eq 'y' -or $verify -eq 'Y') {
            return $NewTagName
        } elseif ($verify -eq 'q' -or $verify -eq 'Q') {
            return $null
        }
    }
}

function Get-NextTagName {
    param(
        [Parameter(Mandatory = $true, HelpMessage = "The current tag name hash.")]
        [object]$VersionHash
    )

    $tmpVersionHash = $VersionHash.Clone()
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

function Invoke-NewGitTag {
    $currentTag = Get-MostRecentTag
    $newTagName = ""
    $ProjectNSI = Join-Path $PSScriptRoot "build" "windows" "installer" "project.nsi"
    $WailsJsonPath = Join-Path $PSScriptRoot "wails.json"

    if (-not (Confirm-RepositoryIsClean)) {
        return
    }

    if (-not $currentTag) {
        Write-Host -ForegroundColor Yellow "No current tag found."
        if (-not $TagName) {
            # Write-Host "No tag name provided."
            $tmpTagName = "v0.0.1-alpha1"
            $newTagName = Get-NewTagName -SuggestedTagName $tmpTagName
        } else {
            $newTagName = Get-NewTagName -SuggestedTagName $TagName
        }
    } else {
        $versionHash = Get-VersionHash -TagName $currentTag
        Write-Host -ForegroundColor Green "Current (most recent) tag: ${currentTag}"
        if (-not $TagName) {
            $tmpTagName = Get-NextTagName -VersionHash $versionHash
            $newTagName = Get-NewTagName -SuggestedTagName $tmpTagName
        } else {
            $newTagName = Get-NewTagName -SuggestedTagName $TagName
        }
    }

    Write-Host -NoNewLine -ForegroundColor Cyan " Current Tag: " ; Write-Host $currentTag
    # echo "Version Hash: $($versionHash | Out-String)"

    if ($newTagName) {
        if ([string]::IsNullOrWhiteSpace($Message)) {
            $Message = $newTagName -replace '^v', 'Version '
        }
        Update-ProjectNSI -ProjectNSIPath $ProjectNSI -TagName $newTagName
        Update-WailsJSON -WailsJsonPath $WailsJsonPath -Version $newTagName
        Write-Host -ForegroundColor Green "New Tag Name: $newTagName - Message: $Message"
    }
    else {
        Write-Host -ForegroundColor Red "No new tag name provided. Exiting."
        return
    }
}

Invoke-NewGitTag
