param(
    [string]$OutputDirectory = '',
    [ValidateRange(1, 2147483647)][int]$Ticks = 100000,
    [ValidateRange(1, 2147483647)][int]$Every = 10000,
    [ValidateRange(1, 256)][int]$Workers = 16,
    [int[]]$Seeds = @(1, 7, 42),
    [ValidateSet('control', 'transport', 'transport-no-mutation', 'ecology', 'ecology-no-mutation', 'ecology-no-chemical-diffusion', 'evolvability', 'evolvability-fixed', 'evolvability-no-mutation', 'environment', 'environment-inert', 'environment-no-mutation')]
    [string[]]$Cases = @('control', 'transport', 'transport-no-mutation', 'ecology')
)

$ErrorActionPreference = 'Stop'
if ($Seeds.Count -eq 0 -or ($Seeds | Where-Object { $_ -lt 0 }).Count -gt 0) { throw 'Seeds must be nonempty and nonnegative.' }
if (@($Seeds | Select-Object -Unique).Count -ne $Seeds.Count) { throw 'Duplicate seeds would overwrite results.' }
if ($Cases.Count -eq 0 -or @($Cases | Select-Object -Unique).Count -ne $Cases.Count) { throw 'Cases must be nonempty and unique.' }
$projectRoot = Split-Path -Parent $PSScriptRoot
if ([string]::IsNullOrWhiteSpace($OutputDirectory)) {
    $OutputDirectory = Join-Path $projectRoot ('data/experiments-' + (Get-Date -Format 'yyyyMMdd-HHmmss-fff'))
}
$experimentRoot = [System.IO.Path]::GetFullPath($OutputDirectory)
if (Test-Path -LiteralPath $experimentRoot) { throw "Output directory already exists: $experimentRoot" }
New-Item -ItemType Directory -Path $experimentRoot | Out-Null

$definitions = @{
    'environment' = @{ Size = 32; Extra = @('-ecology', '-matter-diffusion', '4', '-chemical-diffusion', '4', '-copy-model', 'evolving', '-environment', 'coupled') }
    'environment-inert' = @{ Size = 32; Extra = @('-ecology', '-matter-diffusion', '4', '-chemical-diffusion', '4', '-copy-model', 'evolving', '-environment', 'inert') }
    'environment-no-mutation' = @{ Size = 32; Extra = @('-ecology', '-matter-diffusion', '4', '-chemical-diffusion', '4', '-copy-model', 'evolving', '-environment', 'coupled', '-mutation-ppm', '0') }
    'evolvability' = @{ Size = 32; Extra = @('-ecology', '-matter-diffusion', '4', '-chemical-diffusion', '4', '-copy-model', 'evolving') }
    'evolvability-fixed' = @{ Size = 32; Extra = @('-ecology', '-matter-diffusion', '4', '-chemical-diffusion', '4', '-copy-model', 'fixed') }
    'evolvability-no-mutation' = @{ Size = 32; Extra = @('-ecology', '-matter-diffusion', '4', '-chemical-diffusion', '4', '-copy-model', 'evolving', '-mutation-ppm', '0') }
    'control' = @{ Size = 64; Extra = @() }
    'transport' = @{ Size = 64; Extra = @('-matter-diffusion', '4') }
    'transport-no-mutation' = @{ Size = 64; Extra = @('-matter-diffusion', '4', '-mutation-ppm', '0') }
    'ecology' = @{ Size = 32; Extra = @('-ecology', '-matter-diffusion', '4', '-chemical-diffusion', '4') }
    'ecology-no-mutation' = @{ Size = 32; Extra = @('-ecology', '-matter-diffusion', '4', '-chemical-diffusion', '4', '-mutation-ppm', '0') }
    'ecology-no-chemical-diffusion' = @{ Size = 32; Extra = @('-ecology', '-matter-diffusion', '4', '-chemical-diffusion', '0') }
}
$running = [System.Collections.Generic.List[object]]::new()
$summaries = [System.Collections.Generic.List[object]]::new()
$manifest = [ordered]@{ Status = 'starting'; Workers = $Workers; GOMAXPROCS = 1; Seeds = $Seeds; Cases = $Cases; Ticks = $Ticks; Every = $Every; StartedUTC = [DateTime]::UtcNow.ToString('o') }
$manifestPath = Join-Path $experimentRoot 'manifest.json'
$batchClock = [System.Diagnostics.Stopwatch]::StartNew()

function Save-Progress {
    $manifest.Completed = $summaries.Count
    $manifest.ElapsedSeconds = $batchClock.Elapsed.TotalSeconds
    $manifest | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $manifestPath -Encoding utf8
    ConvertTo-Json -InputObject @($summaries | Sort-Object Case, Seed) -Depth 10 |
        Set-Content -LiteralPath (Join-Path $experimentRoot 'summary.json') -Encoding utf8
}

Push-Location $projectRoot
try {
    $revision = & git rev-parse HEAD 2>$null
    if ($LASTEXITCODE -eq 0) { $manifest.SourceRevision = $revision; $manifest.SourceDirty = [bool](& git status --porcelain) }
    $binary = Join-Path $experimentRoot 'sim.exe'
    & go build -o $binary ./cmd/sim
    if ($LASTEXITCODE -ne 0) { throw 'Build failed.' }
    $manifest.BinarySHA256 = (Get-FileHash -LiteralPath $binary -Algorithm SHA256).Hash.ToLowerInvariant()
    $queue = [System.Collections.Generic.Queue[object]]::new()
    foreach ($case in $Cases) {
        foreach ($seed in $Seeds) { $queue.Enqueue(@{ Case = $case; Seed = $seed; Definition = $definitions[$case] }) }
    }
    $manifest.Total = $queue.Count
    $manifest.Status = 'running'
    Save-Progress
    $lastProgress = 0.0
    while ($queue.Count -gt 0 -or $running.Count -gt 0) {
        while ($queue.Count -gt 0 -and $running.Count -lt $Workers) {
            $job = $queue.Dequeue()
            $job.Stem = Join-Path $experimentRoot ($job.Case + '-seed' + $job.Seed)
            $arguments = @('-width', $job.Definition.Size, '-height', $job.Definition.Size, '-seed', $job.Seed,
                '-ticks', $Ticks, '-every', $Every, '-save', ($job.Stem + '.json'), '-metrics', ($job.Stem + '.jsonl')) + $job.Definition.Extra
            $info = [System.Diagnostics.ProcessStartInfo]::new()
            $info.FileName = $binary
            $info.WorkingDirectory = $projectRoot
            $info.UseShellExecute = $false
            $info.CreateNoWindow = $true
            $info.WindowStyle = [System.Diagnostics.ProcessWindowStyle]::Hidden
            $info.RedirectStandardOutput = $true
            $info.RedirectStandardError = $true
            $info.Environment['GOMAXPROCS'] = '1'
            foreach ($argument in $arguments) { $info.ArgumentList.Add([string]$argument) }
            $job.Clock = [System.Diagnostics.Stopwatch]::StartNew()
            $job.Process = [System.Diagnostics.Process]::Start($info)
            $job.Output = $job.Process.StandardOutput.ReadToEndAsync()
            $job.Error = $job.Process.StandardError.ReadToEndAsync()
            $running.Add($job)
        }
        for ($index = $running.Count - 1; $index -ge 0; $index--) {
            $job = $running[$index]
            if (-not $job.Process.HasExited) { continue }
            $job.Process.WaitForExit()
            $job.Clock.Stop()
            $output = $job.Output.GetAwaiter().GetResult()
            $errorText = $job.Error.GetAwaiter().GetResult()
            Set-Content -LiteralPath ($job.Stem + '.stdout.txt') -Value $output -Encoding utf8
            if ($job.Process.ExitCode -ne 0) { throw "Simulation failed: $($job.Case), seed $($job.Seed): $errorText" }
            if ($output -notmatch 'state_sha256=([0-9a-f]{64})') { throw 'Simulation did not return its final state hash.' }
            $hash = $Matches[1]
            $last = Get-Content -LiteralPath ($job.Stem + '.jsonl') -Tail 1 | ConvertFrom-Json
            $state = Get-Content -Raw -LiteralPath ($job.Stem + '.json') | ConvertFrom-Json
            $summaries.Add([pscustomobject]@{
                Case = $job.Case; Seed = $job.Seed; Width = $job.Definition.Size
                Tick = $last.tick; Entities = $last.entities; Genomes = $last.genomes
                Copies = $last.copies; RecentCopies = $last.interval.copies; RecentTicks = $last.interval.ticks
                LitMatter = $last.lit_matter; DarkMatter = $last.dark_matter; Reactions = $last.converted
                Kernel = $state.kernel_version; Rules = $state.rule_version
                StateSHA256 = $hash; Seconds = $job.Clock.Elapsed.TotalSeconds
                Snapshot = [System.IO.Path]::GetFileName($job.Stem + '.json')
                Metrics = [System.IO.Path]::GetFileName($job.Stem + '.jsonl')
            })
            $job.Process.Dispose()
            $running.RemoveAt($index)
            Save-Progress
            Write-Output ("completed {0}/{1}: {2} seed={3}, entities={4}, recent_copies={5}, seconds={6:F1}" -f
                $summaries.Count, $manifest.Total, $job.Case, $job.Seed, $last.entities, $last.interval.copies, $job.Clock.Elapsed.TotalSeconds)
        }
        if ($batchClock.Elapsed.TotalSeconds - $lastProgress -ge 15) {
            Write-Output "running=$($running.Count) queued=$($queue.Count) completed=$($summaries.Count)/$($manifest.Total)"
            $lastProgress = $batchClock.Elapsed.TotalSeconds
        }
        if ($running.Count -gt 0) { Start-Sleep -Milliseconds 50 }
    }
    $manifest.Status = 'complete'
    Save-Progress
}
catch {
    $manifest.Status = 'failed'
    $manifest.Error = $_.Exception.Message
    Save-Progress
    throw
}
finally {
    # Only processes started by this invocation are stopped on failure/cancel.
    foreach ($job in $running) {
        if (-not $job.Process.HasExited) { $job.Process.Kill($true); $job.Process.WaitForExit() }
        $job.Process.Dispose()
    }
    if ($manifest.Status -eq 'running') { $manifest.Status = 'interrupted'; Save-Progress }
    Pop-Location
}
