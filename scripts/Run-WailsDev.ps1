[CmdletBinding()]
param(
    [Parameter(Mandatory = $false, Position = 0, HelpMessage = "Path to the markdown file")]
    [Alias("file","f")]
    [string]$FilePath,
    [Parameter(Mandatory = $false, HelpMessage = "Do a 'wails build --clean' before running 'wails dev'")]
    [Alias("c")]
    [switch]$Clean
)

# Set up script and project paths
$ScriptFullName = $MyInvocation.MyCommand.Path
$ScriptRoot = Split-Path -Parent $ScriptFullName
$ScriptName = Split-Path -Leaf $ScriptFullName ; $ScriptName | Out-Null # Suppress variable not used warning
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

function Invoke-WailsDev {
    Push-Location $ProjectRoot -StackName 'ProjectRoot'

    $sample_file = "${ProjectRoot}\docs\sample.md"

    # if ($args.Length -gt 0) {
        if (Test-Path $FilePath -ErrorAction SilentlyContinue) {
            $sample_file = $FilePath
        }
    # }

    if ($Clean) {
        Write-Host -ForegroundColor Cyan "Running: wails build --clean"
        wails build --clean
    }

    $sample_file = $sample_file.Replace('\', '\\')

    Write-Host -ForegroundColor Cyan 'Running: wails dev -appargs "--file=""' + ${sample_file} + '""'
    wails dev -appargs "--file=""${sample_file}"""
    # wails dev -loglevel Trace -appargs "--nohtml"

    Pop-Location -StackName 'ProjectRoot'
}

Invoke-WailsDev