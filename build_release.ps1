param(
  [string]$AppName = "Nukastrest"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

$dist = Join-Path $root "dist"
New-Item -ItemType Directory -Force $dist | Out-Null

$go = "go"

Write-Host "==> Building Windows GUI (Wails)"
$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
& $go build -tags production -trimpath -ldflags "-s -w -H=windowsgui" -o (Join-Path $dist "$AppName-windows-amd64-wails.exe") ./cmd/nukastrest-wails

Write-Host "==> Building Windows CLI"
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
& $go build -trimpath -ldflags "-s -w" -o (Join-Path $dist "$AppName-windows-amd64-cli.exe") ./cmd/nukastrest-cli

Write-Host "==> Building Linux CLI"
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
& $go build -trimpath -ldflags "-s -w" -o (Join-Path $dist "$AppName-linux-amd64-cli") ./cmd/nukastrest-cli

Write-Host "==> Linux GUI (Wails) build is produced by GitHub Actions release workflow (native Linux runner)."
Write-Host "Done. Artifacts in: $dist"