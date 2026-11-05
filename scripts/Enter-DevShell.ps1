# Dot-source this script to expose the project's installed development tools.
# Usage: . ./scripts/Enter-DevShell.ps1
$ErrorActionPreference = 'Stop'

$projectRoot = Split-Path -Parent $PSScriptRoot
$toolDirectories = @((Join-Path $projectRoot 'node_modules/.bin'))

if (-not (Get-Command bd -ErrorAction SilentlyContinue)) {
    $beadsDirectory = Join-Path $env:LOCALAPPDATA 'Programs/bd'
    if (-not (Test-Path -LiteralPath (Join-Path $beadsDirectory 'bd.exe'))) {
        throw 'Beads is not installed on PATH or in LOCALAPPDATA/Programs/bd.'
    }
    $toolDirectories += $beadsDirectory
}

foreach ($directory in $toolDirectories) {
    if (($env:PATH -split [IO.Path]::PathSeparator) -notcontains $directory) {
        $env:PATH = $directory + [IO.Path]::PathSeparator + $env:PATH
    }
}

Get-Command bd, node, npm, go, hereby | Select-Object Name, Source
