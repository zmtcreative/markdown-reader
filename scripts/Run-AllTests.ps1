#Requires -Version 7.0

[CmdletBinding()]
param(
    [Parameter(Mandatory = $false, HelpMessage = "Suppress command output and summary. Exit with 0 when all test suites pass, otherwise 1.")]
    [Alias("q")]
    [switch]$Silent,

    [Parameter(Mandatory = $false, HelpMessage = "Run the full frontend Playwright suite instead of the default fast headless slice.")]
    [switch]$RunAllTests,

    [Parameter(Mandatory = $false, HelpMessage = "Open the Playwright HTML report after frontend tests complete.")]
    [switch]$ShowFrontendReport
)

$ScriptFullName = $MyInvocation.MyCommand.Path
$ScriptRoot = Split-Path -Parent $ScriptFullName
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

$FrontendScript = Join-Path $ScriptRoot "Run-FrontendTests.ps1"
$PowerShellExe = Join-Path $PSHOME "pwsh.exe"

function Test-Command {
    param([Parameter(Mandatory = $true)][string]$Command)

    try {
        Get-Command $Command -ErrorAction Stop | Out-Null
        return $true
    } catch {
        return $false
    }
}

function Write-Status {
    param(
        [Parameter(Mandatory = $true)][string]$Message,
        [ConsoleColor]$ForegroundColor = [ConsoleColor]::Gray
    )

    if (-not $Silent) {
        Write-Host $Message -ForegroundColor $ForegroundColor
    }
}

function Invoke-TestCommand {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][scriptblock]$Command
    )

    $startedAt = Get-Date
    $exitCode = 1
    $errorMessage = $null

    Write-Status "Running $Name..." Cyan

    try {
        & $Command
        $exitCode = if ($null -ne $LASTEXITCODE) { [int]$LASTEXITCODE } else { 0 }
    } catch {
        $exitCode = 1
        $errorMessage = $_.Exception.Message
        Write-Status "$Name failed: $errorMessage" Red
    }

    $duration = (Get-Date) - $startedAt

    if ($exitCode -eq 0) {
        Write-Status "$Name passed in $($duration.ToString('hh\:mm\:ss'))" Green
    } else {
        Write-Status "$Name failed with exit code $exitCode after $($duration.ToString('hh\:mm\:ss'))" Yellow
    }

    [pscustomobject]@{
        Name = $Name
        ExitCode = $exitCode
        Passed = ($exitCode -eq 0)
        Duration = $duration
        ErrorMessage = $errorMessage
    }
}

if (-not (Test-Command "go")) {
    Write-Host -ForegroundColor Red "The 'go' command is not available in PATH."
    exit 1
}

if (-not (Test-Path -Path $PowerShellExe -PathType Leaf)) {
    Write-Host -ForegroundColor Red "Could not find pwsh.exe under PSHOME: $PowerShellExe"
    exit 1
}

if (-not (Test-Path -Path $FrontendScript -PathType Leaf)) {
    Write-Host -ForegroundColor Red "Could not find frontend test script: $FrontendScript"
    exit 1
}

$FrontendTestSuite = "fast"
$FrontendRuntimeMode = "headless"

if ($RunAllTests) {
    $FrontendTestSuite = "all"
    $FrontendRuntimeMode = "auto"
}

$frontendArgs = @(
    "-NoProfile"
    "-File"
    $FrontendScript
    "-TestSuite"
    $FrontendTestSuite
    "-RuntimeMode"
    $FrontendRuntimeMode
)

if ($ShowFrontendReport) {
    $frontendArgs += "-ShowReport"
}

$overallStartedAt = Get-Date
Push-Location $ProjectRoot

try {
    $results = @(
        (Invoke-TestCommand -Name "Go tests" -Command {
            if ($Silent) {
                & go test ./... *> $null
            } else {
                & go test ./... | Out-Host
            }
        })
        (Invoke-TestCommand -Name "Frontend tests" -Command {
            if ($Silent) {
                & $PowerShellExe @frontendArgs *> $null
            } else {
                & $PowerShellExe @frontendArgs | Out-Host
            }
        })
    )
} finally {
    Pop-Location
}

$allPassed = ($results | Where-Object { -not $_.Passed }).Count -eq 0
$overallDuration = (Get-Date) - $overallStartedAt

if (-not $Silent) {
    Write-Host ""
    Write-Host "Test summary" -ForegroundColor Cyan
    Write-Host "============" -ForegroundColor Cyan

    foreach ($result in $results) {
        $statusText = if ($result.Passed) { "PASS" } else { "FAIL" }
        $statusColor = if ($result.Passed) { "Green" } else { "Red" }
        $summaryLine = "[{0}] {1} (exit {2}, duration {3})" -f $statusText, $result.Name, $result.ExitCode, $result.Duration.ToString('hh\:mm\:ss')
        Write-Host $summaryLine -ForegroundColor $statusColor

        if ($result.ErrorMessage) {
            Write-Host "       $($result.ErrorMessage)" -ForegroundColor DarkYellow
        }
    }

    $overallColor = if ($allPassed) { "Green" } else { "Red" }
    $overallText = if ($allPassed) { "All tests passed" } else { "One or more test suites failed" }
    Write-Host ""
    Write-Host "$overallText (total duration $($overallDuration.ToString('hh\:mm\:ss')))" -ForegroundColor $overallColor
}

if ($allPassed) {
    exit 0
}

exit 1