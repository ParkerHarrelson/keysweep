<# Build & smoke-test KeySweep scanner on Windows.
   Usage:  .\scripts\buildPhase1.ps1             # builds keysweep-scan:dev
           .\scripts\buildPhase1.ps1 -Tag mytag  # builds mytag           #>
[CmdletBinding()]
param(
  [string]$Tag = "keysweep-scan:dev"
)

$ErrorActionPreference = "Stop"

# 1. Resolve repo root (git if available, else CWD)
try {
    $RepoRoot = git rev-parse --show-toplevel 2>$null
    if (-not $RepoRoot) { $RepoRoot = (Get-Location).Path }
} catch {
    $RepoRoot = (Get-Location).Path
}
Write-Host "📁 Repo root = $RepoRoot"

# 2. Ensure Docker is up
if (-not (docker info >$null 2>&1)) {
    Write-Error "Docker Desktop is not running.  Start it and retry."
    exit 1
}

# 3. Build Linux binary
$Env:GOOS  = "linux"
$Env:GOARCH = "amd64"
$ScannerExe = Join-Path $RepoRoot "action\keysweep-scanner"

Push-Location (Join-Path $RepoRoot "scanner-cli")
go build -o $ScannerExe .
Pop-Location
Write-Host "✔ Built scanner binary => $ScannerExe"

# build docker image  (context = repo root)
docker build --no-cache -f (Join-Path $RepoRoot "action\Dockerfile") `
             -t $Tag $RepoRoot
Write-Host "🐳 Docker image $Tag built"

# 5. Smoke-test with test.txt
$TestFile = Join-Path $RepoRoot "scanner-cli\test.txt"
# 5. Smoke-test with test.txt (piped into STDIN)
Write-Host "▶ Running smoke test (via STDIN) with $TestFile ..."
Get-Content -Raw $TestFile |
  docker run --rm -i `
    --entrypoint /bin/keysweep-scanner `
    $Tag
$Exit = $LASTEXITCODE
if ($Exit -eq 0) {
    Write-Host "✅  no leaks detected (unexpected for test.txt)" -ForegroundColor Green
} else {
    Write-Host "Gitleaks detected - exit code $Exit (expected)" -ForegroundColor Red
}
exit $Exit
