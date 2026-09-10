# Integration test of scheduling: exactly the same seeds must have identical
# state hashes with one worker and sixteen workers. Outputs remain in data/.
param([string]$OutputDirectory = '')
$ErrorActionPreference = 'Stop'
if ([string]::IsNullOrWhiteSpace($OutputDirectory)) {
    $OutputDirectory = Join-Path (Split-Path -Parent $PSScriptRoot) ('data/runner-test-' + (Get-Date -Format 'yyyyMMdd-HHmmss-fff'))
}
if (Test-Path -LiteralPath $OutputDirectory) { throw 'Test output directory already exists.' }
$parallelPath = Join-Path $OutputDirectory 'parallel'
$serialPath = Join-Path $OutputDirectory 'serial'
& (Join-Path $PSScriptRoot 'experiments.ps1') -Cases ecology -Ticks 3000 -Every 1000 -Seeds (1..16) -Workers 16 -OutputDirectory $parallelPath
& (Join-Path $PSScriptRoot 'experiments.ps1') -Cases ecology -Ticks 3000 -Every 1000 -Seeds (1..16) -Workers 1 -OutputDirectory $serialPath
$parallel = Get-Content -Raw -LiteralPath (Join-Path $parallelPath 'summary.json') | ConvertFrom-Json
$serial = Get-Content -Raw -LiteralPath (Join-Path $serialPath 'summary.json') | ConvertFrom-Json
if ($parallel.Count -ne 16 -or $serial.Count -ne 16) { throw 'Missing runs.' }
foreach ($run in $parallel) {
    $match = $serial | Where-Object Seed -eq $run.Seed
    if ($run.StateSHA256 -ne $match.StateSHA256) { throw "State differs for seed $($run.Seed)" }
}
Write-Output 'All 16 serial/parallel state hashes match.'
