#!/usr/bin/env pwsh
Set-Location $PSScriptRoot

$sample_file = "${PSScriptRoot}\docs\sample.md"

if ($args.Length -gt 0) {
    if (Test-Path $args[0] -ErrorAction SilentlyContinue) {
        $sample_file = $args[0]
    }
}

$sample_file = $sample_file.Replace('\', '\\')

wails dev -appargs "-file ""${sample_file}"""