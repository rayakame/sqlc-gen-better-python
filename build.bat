@echo off
setlocal

:: Set environment variables
set GOOS=wasip1
set GOARCH=wasm

:: Build the Go plugin
go build -o sqlc-gen-better-python.wasm plugin/main.go

:: Generate SHA-256 hash
for /f %%i in ('certutil -hashfile sqlc-gen-better-python.wasm SHA256 ^| findstr /v "hash"') do set SHA256_HASH=%%i

:: Replace SHA-256 in sqlc.yaml
powershell -Command "(Get-Content sqlc.yaml) -replace '(?<=sha256: )\S+', '%SHA256_HASH%' | Set-Content sqlc.yaml"

echo Updated sqlc.yaml with SHA-256: %SHA256_HASH%
endlocal