param(
    [string]$OutputDirectory = '',
    [int]$Ticks = 100000,
    [int]$Every = 10000,
    [int[]]$Seeds = @(1, 7, 42)
)

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
if ($Ticks -lt 1 -or $Every -lt 1 -or $Seeds.Count -eq 0) {
    throw 'Ticks, Every and Seeds must be positive/nonempty.'
}
if ([string]::IsNullOrWhiteSpace($OutputDirectory)) {
    $OutputDirectory = Join-Path $projectRoot ('data/experiments-' + (Get-Date -Format 'yyyyMMdd-HHmmss'))
}
$experimentRoot = [System.IO.Path]::GetFullPath($OutputDirectory)
if (Test-Path -LiteralPath $experimentRoot) { throw "Output directory already exists: $experimentRoot" }
New-Item -ItemType Directory -Path $experimentRoot | Out-Null

Push-Location $projectRoot
try {
    $binary = Join-Path $experimentRoot 'sim.exe'
    & go build -o $binary ./cmd/sim
    if ($LASTEXITCODE -ne 0) { throw 'Build failed.' }
    $cases = @(
        @{ Name = 'control'; Size = 64; Extra = @() },
        @{ Name = 'transport'; Size = 64; Extra = @('-matter-diffusion', '4') },
        @{ Name = 'transport-no-mutation'; Size = 64; Extra = @('-matter-diffusion', '4', '-mutation-ppm', '0') },
        @{ Name = 'ecology'; Size = 32; Extra = @('-ecology', '-matter-diffusion', '4', '-chemical-diffusion', '4') }
    )
    $summaries = @()
    foreach ($case in $cases) {
        foreach ($seed in $Seeds) {
            $stem = Join-Path $experimentRoot ($case.Name + '-seed' + $seed)
            $arguments = @('-width', $case.Size, '-height', $case.Size, '-seed', $seed,
                '-ticks', $Ticks, '-every', $Every, '-save', ($stem + '.json'), '-metrics', ($stem + '.jsonl')) + $case.Extra
            & $binary @arguments
            if ($LASTEXITCODE -ne 0) { throw "Simulation failed: $stem" }
            $last = Get-Content -LiteralPath ($stem + '.jsonl') -Tail 1 | ConvertFrom-Json
            $state = Get-Content -Raw -LiteralPath ($stem + '.json') | ConvertFrom-Json
            $summaries += [pscustomobject]@{
                Case = $case.Name
                Seed = $seed
                Width = $case.Size
                Tick = $last.tick
                Entities = $last.entities
                Genomes = $last.genomes
                Copies = $last.copies
                RecentCopies = $last.interval.copies
                RecentTicks = $last.interval.ticks
                LitMatter = $last.lit_matter
                DarkMatter = $last.dark_matter
                Reactions = $last.converted
                Kernel = $state.kernel_version
                Rules = $state.rule_version
                Snapshot = [System.IO.Path]::GetFileName($stem + '.json')
                Metrics = [System.IO.Path]::GetFileName($stem + '.jsonl')
            }
            $summaries | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Join-Path $experimentRoot 'summary.json') -Encoding utf8
        }
    }
    $summaries | Format-Table Case, Seed, Entities, Genomes, RecentCopies, DarkMatter
}
finally { Pop-Location }
