# Build, verify and package the local Windows compiler SDK without publishing it.
$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
Push-Location $projectRoot
try {
    $targetOS = (& go env GOOS).Trim()
    if ($LASTEXITCODE -ne 0 -or $targetOS -ne 'windows') {
        throw 'This script builds and executes a Windows compiler; GOOS must be windows.'
    }
    $targetArch = (& go env GOARCH).Trim()
    if ($LASTEXITCODE -ne 0) { throw 'Cannot determine the Go target architecture.' }
    $archiveArch = switch ($targetArch) {
        'amd64' { 'x64' }
        'arm64' { 'arm64' }
        default { throw "Unsupported Windows SDK architecture: $targetArch" }
    }
    & npm run build
    if ($LASTEXITCODE -ne 0) { throw 'Compiler build failed.' }
    & node ./scripts/Verify-PipeDelivery.mjs
    if ($LASTEXITCODE -ne 0) { throw 'Compiler delivery verification failed.' }
    $versionOutput = (& ./built/local/tsgo.exe --version).Trim()
    if ($LASTEXITCODE -ne 0 -or $versionOutput -ne 'Version 7.0.2-pipe.0') {
        throw "Unexpected compiler version: $versionOutput"
    }
    # Keep declarations next to the executable, as required by the native SDK.
    $archive = Join-Path $projectRoot "built/pipe-compiler-7.0.2-pipe.0-win32-$archiveArch.zip"
    Compress-Archive -Path ./built/local/* -DestinationPath $archive -CompressionLevel Optimal -Force
    if (-not (Test-Path -LiteralPath $archive -PathType Leaf)) { throw 'SDK archive was not created.' }
    Get-FileHash -LiteralPath $archive -Algorithm SHA256
    Write-Output "Local SDK archive: $archive"
}
finally {
    Pop-Location
}
