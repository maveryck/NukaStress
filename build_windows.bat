@echo off
setlocal
set GOOS=windows
set GOARCH=amd64
echo Building NukaStress portable executable...
go build -trimpath -ldflags "-s -w -H=windowsgui" -o NukaStress.exe .
if errorlevel 1 (
  echo Build failed.
  exit /b 1
)
echo Done. Output: NukaStress.exe
endlocal
