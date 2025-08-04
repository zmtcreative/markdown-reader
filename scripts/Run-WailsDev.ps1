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

    if ($args.Length -gt 0) {
        if (Test-Path $args[0] -ErrorAction SilentlyContinue) {
            $sample_file = $args[0]
        }
    }

    $sample_file = $sample_file.Replace('\', '\\')

    wails dev -appargs "--file=""${sample_file}"""
    # wails dev -loglevel Trace -appargs "--nohtml"

    Pop-Location -StackName 'ProjectRoot'
}

Invoke-WailsDev