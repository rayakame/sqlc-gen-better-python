@echo off
setlocal EnableDelayedExpansion

REM ------------------------------
REM 1) CONFIGURATION - add folders here
REM    (paths are relative to repo root)
REM ------------------------------
set "TARGET_DIRS=test\driver_asyncpg test\driver_psycopg_async test\driver_psycopg_sync test\driver_aiosqlite test\driver_sqlite3 test\driver_turso_sync test\driver_turso_async"
set "SQLC_CONFIG_NAMES=sqlc.yaml"

REM ------------------------------
REM 2) BUILD THE WASM PLUGIN
REM ------------------------------
echo === Building the Go WASM plugin =================================
set GOOS=wasip1
set GOARCH=wasm
go build -o sqlc-gen-better-python.wasm plugin/main.go

REM ------------------------------
REM 3) CALCULATE SHA-256
REM ------------------------------
REM Capture the digest line whole (delims=), then strip the spaces older
REM CertUtil versions print between the hex pairs - first-token capture
REM would truncate such a digest to its first byte.
for /f "delims=" %%i in ('certutil -hashfile sqlc-gen-better-python.wasm SHA256 ^| findstr /v "hash"') do set "SHA256_HASH=%%i"
set "SHA256_HASH=%SHA256_HASH: =%"
echo SHA-256: %SHA256_HASH%

REM ------------------------------
REM 4) UPDATE ROOT yaml
REM ------------------------------
powershell -Command "(Get-Content sqlc.yaml) -replace '(?<=sha256: )\S+', '%SHA256_HASH%' | Set-Content sqlc.yaml"


REM ------------------------------
REM 5) PROPAGATE TO TARGET FOLDERS
REM ------------------------------
for %%D in (%TARGET_DIRS%) do (
    echo --------------------------------------------------------------
    echo   Processing %%D
    if not exist "%%D" (
        echo   Creating folder %%D
        mkdir "%%D"
    )

    REM Copy the plugin
    xcopy /Y /Q "sqlc-gen-better-python.wasm" "%%D\"

    REM Patch that folder's yaml in place
    for %%F in (%SQLC_CONFIG_NAMES%) do (
        powershell -Command "(Get-Content '%%D\\%%F') -replace '(?<=sha256: )\S+', '%SHA256_HASH%' | Set-Content '%%D\\%%F'"
    )
)

echo === All done - every sqlc.yaml now has SHA-256 %SHA256_HASH% ======
endlocal
