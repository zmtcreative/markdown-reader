# Optimized Playwright Testing Script for Wails Application
# This script runs the optimized Playwright test suite using Wails dev environment
# Tests automatically manage their own Wails dev server - no build required
# Note: Tests always run in headed mode (visible browser) for Wails debugging

param(
    [string]$TestSuite = "main", # Options: "main", "fast", "all"
    [switch]$ShowReport = $false
)

Write-Host "🚀 Markdown Reader - Optimized Playwright Testing Script" -ForegroundColor Cyan
Write-Host "======================================================" -ForegroundColor Cyan

# Function to check if a command exists
function Test-Command {
    param([string]$Command)
    try {
        Get-Command $Command -ErrorAction Stop | Out-Null
        return $true
    } catch {
        return $false
    }
}

# Check prerequisites
Write-Host "🔍 Checking prerequisites..." -ForegroundColor Yellow

if (-not (Test-Command "wails")) {
    Write-Error "❌ Wails CLI not found. Please install Wails v2 first."
    exit 1
}

if (-not (Test-Command "npm")) {
    Write-Error "❌ npm not found. Please install Node.js first."
    exit 1
}

# Set working directory to project root
$projectRoot = Split-Path $PSScriptRoot -Parent
Set-Location $projectRoot

Write-Host "📂 Project root: $projectRoot" -ForegroundColor Green
Write-Host "ℹ️ Tests will manage Wails dev server automatically (no build required)" -ForegroundColor Blue

# Install Playwright dependencies
Write-Host "📦 Installing Playwright dependencies..." -ForegroundColor Yellow
Set-Location "frontend"

try {
    npm install
    if ($LASTEXITCODE -ne 0) {
        throw "npm install failed"
    }
    Write-Host "✅ Dependencies installed!" -ForegroundColor Green
} catch {
    Write-Error "❌ Failed to install dependencies: $_"
    exit 1
}

# Install Playwright browsers if needed
Write-Host "🌐 Ensuring Playwright browsers are installed..." -ForegroundColor Yellow
try {
    npx playwright install chromium
    if ($LASTEXITCODE -ne 0) {
        throw "Playwright browser installation failed"
    }
    Write-Host "✅ Playwright browsers ready!" -ForegroundColor Green
} catch {
    Write-Error "❌ Failed to install Playwright browsers: $_"
    exit 1
}

# Prepare test results directory
$testResultsDir = Join-Path $PWD "test-results"
if (Test-Path $testResultsDir) {
    Write-Host "🧹 Cleaning previous test results..." -ForegroundColor Gray
    Remove-Item $testResultsDir -Recurse -Force
}
New-Item -ItemType Directory -Path $testResultsDir -Force | Out-Null
Write-Host "📁 Test results directory prepared: $testResultsDir" -ForegroundColor Green

# Determine which test suite to run
Write-Host "🧪 Running Optimized Playwright Test Suite (with Wails dev server)..." -ForegroundColor Yellow

$playwrightArgs = @("test")

# Add test suite selection
switch ($TestSuite.ToLower()) {
    "main" {
        $playwrightArgs += "main-test-suite.spec.ts"
        Write-Host "🎯 Running Main Test Suite (comprehensive, optimized)" -ForegroundColor Blue
    }
    "fast" {
        $playwrightArgs += "fast-sequential-tests.spec.ts"
        Write-Host "⚡ Running Fast Sequential Test Suite (performance demo)" -ForegroundColor Blue
    }
    "all" {
        Write-Host "🔄 Running All Available Test Suites" -ForegroundColor Blue
    }
    default {
        $playwrightArgs += "main-test-suite.spec.ts"
        Write-Host "🎯 Running Main Test Suite (default)" -ForegroundColor Blue
    }
}

# Essential Playwright configuration
$playwrightArgs += "--workers=1"  # Essential for Wails dev server stability

Write-Host "👁️ Running tests with visible browser (headed mode - hardcoded for Wails debugging)" -ForegroundColor Blue

Write-Host "ℹ️ Tests will automatically:" -ForegroundColor Cyan
Write-Host "   • Start Wails dev server (wails dev)" -ForegroundColor Gray
Write-Host "   • Connect to visible browser for Wails debugging" -ForegroundColor Gray
Write-Host "   • Run optimized test suite with shared app instances" -ForegroundColor Gray
Write-Host "   • Clean up Wails dev server when complete" -ForegroundColor Gray

try {
    # Build the command string
    $commandString = "npx playwright $($playwrightArgs -join ' ')"
    Write-Host "🎬 Executing: $commandString" -ForegroundColor Gray

    # Execute using Invoke-Expression to handle the command properly
    Invoke-Expression $commandString
    $testExitCode = $LASTEXITCODE

    if ($testExitCode -eq 0) {
        Write-Host "✅ All tests passed!" -ForegroundColor Green
    } else {
        Write-Host "⚠️ Some tests failed (exit code: $testExitCode)" -ForegroundColor Yellow
    }
} catch {
    Write-Error "❌ Failed to run tests: $_"
    exit 1
}

# Show test report if requested
if ($ShowReport) {
    Write-Host "📊 Opening test report..." -ForegroundColor Yellow
    try {
        npx playwright show-report --host 127.0.0.1 --port 9323
    } catch {
        Write-Host "ℹ️ Could not open test report automatically" -ForegroundColor Blue
        Write-Host "💡 You can view the report manually with: npx playwright show-report" -ForegroundColor Blue
    }
}

# Return to project root
Set-Location $projectRoot

Write-Host "🎉 Testing completed!" -ForegroundColor Cyan
Write-Host "📸 Screenshots and videos are available in: frontend/test-results/" -ForegroundColor Cyan
Write-Host "📋 HTML report is available in: frontend/playwright-report/" -ForegroundColor Cyan

exit $testExitCode
