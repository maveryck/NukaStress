$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

$dist = Join-Path $root "dist"
New-Item -ItemType Directory -Force $dist | Out-Null

$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

go build -tags production -trimpath -ldflags "-s -w -H=windowsgui" -o (Join-Path $dist "Nukastrest-windows-amd64-wails.exe") ./cmd/nukastrest-wails
Write-Host "Built: dist/Nukastrest-windows-amd64-wails.exe"