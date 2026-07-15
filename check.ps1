param(
    [string]$RepoRoot = (Get-Location).Path,
    [string]$OutRoot = '',
    [int]$Port = 18080
)


# Embedded common helpers. This file is standalone and can be run from the repository root.

Set-StrictMode -Version 2.0

function Get-CheckGoCommand {
    $go = Get-Command go -ErrorAction SilentlyContinue
    if ($go) {
        return $go.Source
    }

    throw 'go executable was not found in PATH. Install Go and make sure go is available in PATH.'
}

function New-CheckContext {
    param(
        [Parameter(Mandatory=$true)][string]$Student,
        [Parameter(Mandatory=$true)][string]$RepoRoot,
        [string]$OutRoot = ''
    )

    $repo = (Resolve-Path -LiteralPath $RepoRoot).Path
    if ($OutRoot -eq '') {
        $OutRoot = Join-Path $repo '.check-results'
    }

    $timestamp = Get-Date -Format 'yyyyMMdd_HHmmss'
    $safeStudent = $Student -replace '[^A-Za-z0-9_.-]', '_'
    $resultDir = Join-Path $OutRoot "${safeStudent}_${timestamp}"
    $logsDir = Join-Path $resultDir 'logs'
    $inputsDir = Join-Path $resultDir 'inputs'
    $outputsDir = Join-Path $resultDir 'outputs'
    $metaDir = Join-Path $resultDir 'meta'
    $tmpDir = Join-Path $resultDir 'tmp'

    foreach ($dir in @($resultDir, $logsDir, $inputsDir, $outputsDir, $metaDir, $tmpDir)) {
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
    }

    $ctx = [ordered]@{
        Student = $Student
        RepoRoot = $repo
        ResultDir = $resultDir
        LogsDir = $logsDir
        InputsDir = $inputsDir
        OutputsDir = $outputsDir
        MetaDir = $metaDir
        TmpDir = $tmpDir
        CommandsPath = Join-Path $resultDir 'commands.jsonl'
        GoCmd = Get-CheckGoCommand
        StartedAt = (Get-Date).ToString('o')
        CommandResults = @{}
        Assessments = New-Object System.Collections.ArrayList
    }

    '' | Set-Content -LiteralPath $ctx.CommandsPath -Encoding UTF8
    return $ctx
}

function Write-CheckText {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)][string]$RelativePath,
        [Parameter(Mandatory=$true)][string]$Content
    )

    $path = Join-Path $Ctx.ResultDir $RelativePath
    $parent = Split-Path -Parent $path
    if ($parent) {
        New-Item -ItemType Directory -Force -Path $parent | Out-Null
    }
    Set-Content -LiteralPath $path -Value $Content -Encoding UTF8
    return $path
}

function Save-CheckJson {
    param(
        [Parameter(Mandatory=$true)][string]$Path,
        [Parameter(Mandatory=$true)]$Value
    )

    $json = $Value | ConvertTo-Json -Depth 30
    Set-Content -LiteralPath $Path -Value $json -Encoding UTF8
}

function Invoke-CheckCommand {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)][string]$Name,
        [Parameter(Mandatory=$true)][string]$Command,
        [string]$WorkingDirectory = ''
    )

    if ($WorkingDirectory -eq '') {
        $WorkingDirectory = $Ctx.RepoRoot
    }

    $safeName = $Name -replace '[^A-Za-z0-9_.-]', '_'
    $logPath = Join-Path $Ctx.LogsDir "$safeName.log"
    $runnerPath = Join-Path $Ctx.TmpDir "$safeName.ps1"
    $started = Get-Date

    $runner = @"
`$ErrorActionPreference = 'Continue'
Set-Location -LiteralPath '$($WorkingDirectory.Replace("'", "''"))'
$Command
`$exitCode = `$global:LASTEXITCODE
if (`$null -eq `$exitCode) { `$exitCode = 0 }
exit `$exitCode
"@

    Set-Content -LiteralPath $runnerPath -Value $runner -Encoding UTF8

    $output = & powershell.exe -NoProfile -ExecutionPolicy Bypass -File $runnerPath 2>&1
    $exitCode = $LASTEXITCODE
    $ended = Get-Date

    @(
        "name: $Name"
        "working_directory: $WorkingDirectory"
        "command:"
        $Command
        "exit_code: $exitCode"
        "started_at: $($started.ToString('o'))"
        "ended_at: $($ended.ToString('o'))"
        ""
        "output:"
        ($output | Out-String)
    ) | Set-Content -LiteralPath $logPath -Encoding UTF8

    $record = [ordered]@{
        name = $Name
        command = $Command
        working_directory = $WorkingDirectory
        exit_code = $exitCode
        started_at = $started.ToString('o')
        ended_at = $ended.ToString('o')
        duration_ms = [int](($ended - $started).TotalMilliseconds)
        log = "logs/$safeName.log"
    }
    ($record | ConvertTo-Json -Compress) | Add-Content -LiteralPath $Ctx.CommandsPath -Encoding UTF8
    $Ctx.CommandResults[$Name] = $record
    $script:LAST_CHECK_EXIT_CODE = $exitCode
}

function Add-FeatureAssessment {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)][string]$Id,
        [Parameter(Mandatory=$true)][ValidateSet('minimum','good','excellent','engineering')][string]$Level,
        [Parameter(Mandatory=$true)][string]$Category,
        [Parameter(Mandatory=$true)][string]$Requirement,
        [Parameter(Mandatory=$true)][ValidateSet('not_implemented','partial','full')][string]$Implementation,
        [Parameter(Mandatory=$true)][ValidateSet('not_tested','nonconformant','conformant')][string]$Conformance,
        [string[]]$Evidence = @(),
        [string]$Details = ''
    )

    $item = [ordered]@{
        id = $Id
        level = $Level
        category = $Category
        requirement = $Requirement
        implementation = $Implementation
        conformance = $Conformance
        evidence = @($Evidence)
        details = $Details
    }
    $Ctx.Assessments.Add($item) | Out-Null
}

function Add-CommandFeatureAssessment {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)][string]$Id,
        [Parameter(Mandatory=$true)][ValidateSet('minimum','good','excellent','engineering')][string]$Level,
        [Parameter(Mandatory=$true)][string]$Category,
        [Parameter(Mandatory=$true)][string]$Requirement,
        [Parameter(Mandatory=$true)][string]$CommandName,
        [string[]]$RequiredArtifacts = @(),
        [string]$Details = ''
    )

    $hasCommand = $Ctx.CommandResults.ContainsKey($CommandName)
    $exitCode = if ($hasCommand) { [int]$Ctx.CommandResults[$CommandName].exit_code } else { -999 }
    $missingArtifacts = @($RequiredArtifacts | Where-Object { -not (Test-Path -LiteralPath $_) })
    $implementation = 'not_implemented'
    $conformance = 'not_tested'

    if ($hasCommand) {
        $implementation = if ($exitCode -eq 0 -and $missingArtifacts.Count -eq 0) { 'full' } else { 'partial' }
        $conformance = if ($exitCode -eq 0 -and $missingArtifacts.Count -eq 0) { 'conformant' } else { 'nonconformant' }
    }

    $evidence = @()
    if ($hasCommand) {
        $evidence += [string]$Ctx.CommandResults[$CommandName].log
    }
    foreach ($artifact in $RequiredArtifacts) {
        if (Test-Path -LiteralPath $artifact) {
            $evidence += $artifact.Replace($Ctx.ResultDir, '').TrimStart('\')
        }
    }

    $detailParts = @()
    if ($Details) {
        $detailParts += $Details
    }
    if ($hasCommand) {
        $detailParts += "exit_code=$exitCode"
    } else {
        $detailParts += 'command was not executed'
    }
    if ($missingArtifacts.Count -gt 0) {
        $detailParts += "missing artifacts: $($missingArtifacts -join ', ')"
    }

    Add-FeatureAssessment -Ctx $Ctx -Id $Id -Level $Level -Category $Category -Requirement $Requirement -Implementation $implementation -Conformance $conformance -Evidence $evidence -Details ($detailParts -join '; ')
}

function Add-BooleanFeatureAssessment {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)][string]$Id,
        [Parameter(Mandatory=$true)][ValidateSet('minimum','good','excellent','engineering')][string]$Level,
        [Parameter(Mandatory=$true)][string]$Category,
        [Parameter(Mandatory=$true)][string]$Requirement,
        [Parameter(Mandatory=$true)][bool]$Implemented,
        [Parameter(Mandatory=$true)][bool]$Conformant,
        [string[]]$Evidence = @(),
        [string]$Details = ''
    )

    $implementation = if ($Implemented) { 'full' } else { 'not_implemented' }
    $conformance = if (-not $Implemented) { 'not_tested' } elseif ($Conformant) { 'conformant' } else { 'nonconformant' }
    Add-FeatureAssessment -Ctx $Ctx -Id $Id -Level $Level -Category $Category -Requirement $Requirement -Implementation $implementation -Conformance $conformance -Evidence $Evidence -Details $Details
}

function Add-SourceFeatureAssessment {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)][string]$Id,
        [Parameter(Mandatory=$true)][ValidateSet('minimum','good','excellent','engineering')][string]$Level,
        [Parameter(Mandatory=$true)][string]$Category,
        [Parameter(Mandatory=$true)][string]$Requirement,
        [Parameter(Mandatory=$true)][string[]]$Patterns,
        [ValidateSet('any','all')][string]$Match = 'all',
        [string]$Details = ''
    )

    $files = @(Get-ChildItem -LiteralPath $Ctx.RepoRoot -Recurse -File -ErrorAction SilentlyContinue | Where-Object {
        $_.FullName -notlike '*\.check-results\*' -and ($_.Extension -in @('.go', '.md') -or $_.Name -eq 'Makefile')
    })
    $matchedPatterns = @()
    $evidence = @()
    foreach ($pattern in $Patterns) {
        $hits = @($files | Select-String -Pattern $pattern -ErrorAction SilentlyContinue)
        if ($hits.Count -gt 0) {
            $matchedPatterns += $pattern
            $evidence += @($hits | Select-Object -First 5 | ForEach-Object {
                "$($_.Path):$($_.LineNumber)"
            })
        }
    }

    $implemented = if ($Match -eq 'all') {
        $matchedPatterns.Count -eq $Patterns.Count
    } else {
        $matchedPatterns.Count -gt 0
    }
    $implementation = if ($implemented) { 'partial' } else { 'not_implemented' }
    $detailText = "source-only check; matched=$($matchedPatterns.Count)/$($Patterns.Count)"
    if ($Details) {
        $detailText = "$Details; $detailText"
    }
    Add-FeatureAssessment -Ctx $Ctx -Id $Id -Level $Level -Category $Category -Requirement $Requirement -Implementation $implementation -Conformance 'not_tested' -Evidence ($evidence | Select-Object -Unique) -Details $detailText
}

function Add-StandardEngineeringAssessments {
    param(
        [Parameter(Mandatory=$true)]$Ctx
    )

    $testFiles = @(Get-ChildItem -LiteralPath $Ctx.RepoRoot -Recurse -File -Filter '*_test.go' -ErrorAction SilentlyContinue | Where-Object { $_.FullName -notlike '*\.check-results\*' })
    $testFunctions = @($testFiles | Select-String -Pattern '^\s*func\s+Test[A-Za-z0-9_]+\s*\(' -ErrorAction SilentlyContinue)
    $benchmarkFunctions = @($testFiles | Select-String -Pattern '^\s*func\s+Benchmark[A-Za-z0-9_]+\s*\(' -ErrorAction SilentlyContinue)

    $testFileEvidence = @($testFiles | ForEach-Object { $_.FullName })
    Add-BooleanFeatureAssessment -Ctx $Ctx -Id 'engineering.unit_tests_present' -Level 'engineering' -Category 'tests' -Requirement 'Go unit tests are present' -Implemented ($testFunctions.Count -gt 0) -Conformant ($testFunctions.Count -gt 0) -Evidence $testFileEvidence -Details "test_files=$($testFiles.Count); test_functions=$($testFunctions.Count)"
    Add-BooleanFeatureAssessment -Ctx $Ctx -Id 'engineering.benchmarks_present' -Level 'engineering' -Category 'benchmarks' -Requirement 'Go benchmark tests are present' -Implemented ($benchmarkFunctions.Count -gt 0) -Conformant ($benchmarkFunctions.Count -gt 0) -Evidence $testFileEvidence -Details "benchmark_functions=$($benchmarkFunctions.Count)"

    if ($Ctx.CommandResults.ContainsKey('go_test_all')) {
        Add-CommandFeatureAssessment -Ctx $Ctx -Id 'engineering.go_test_passes' -Level 'engineering' -Category 'tests' -Requirement 'go test ./... passes' -CommandName 'go_test_all'
    }
    if ($Ctx.CommandResults.ContainsKey('go_test_bench')) {
        Add-CommandFeatureAssessment -Ctx $Ctx -Id 'engineering.benchmarks_run' -Level 'engineering' -Category 'benchmarks' -Requirement 'Benchmark tests run' -CommandName 'go_test_bench'
    }
    if ($Ctx.CommandResults.ContainsKey('go_test_race')) {
        Add-CommandFeatureAssessment -Ctx $Ctx -Id 'engineering.race_test_passes' -Level 'engineering' -Category 'tests' -Requirement 'go test -race ./... passes' -CommandName 'go_test_race'
    }

    $readmePath = Join-Path $Ctx.RepoRoot 'README.md'
    $readmeOk = (Test-Path -LiteralPath $readmePath) -and ((Get-Item -LiteralPath $readmePath).Length -gt 100)
    Add-BooleanFeatureAssessment -Ctx $Ctx -Id 'engineering.readme' -Level 'engineering' -Category 'documentation' -Requirement 'README.md exists and is not empty' -Implemented $readmeOk -Conformant $readmeOk -Evidence @('repo_snapshot/README.md')

    $makefilePath = Join-Path $Ctx.RepoRoot 'Makefile'
    $makefileText = if (Test-Path -LiteralPath $makefilePath) { Get-Content -LiteralPath $makefilePath -Raw } else { '' }
    foreach ($target in @('test','bench','demo')) {
        $targetOk = $makefileText -match "(?m)^\s*${target}\s*:"
        Add-BooleanFeatureAssessment -Ctx $Ctx -Id "engineering.make_$target" -Level 'engineering' -Category 'reproducibility' -Requirement "Makefile has target $target" -Implemented $targetOk -Conformant $targetOk -Evidence @('repo_snapshot/Makefile')
    }

    $controlPath = Join-Path $Ctx.RepoRoot 'testdata\control'
    $controlFiles = @()
    if (Test-Path -LiteralPath $controlPath) {
        $controlFiles = @(Get-ChildItem -LiteralPath $controlPath -Recurse -File -ErrorAction SilentlyContinue)
    }
    $controlEvidence = @($controlFiles | ForEach-Object { $_.FullName })
    Add-BooleanFeatureAssessment -Ctx $Ctx -Id 'engineering.control_data' -Level 'engineering' -Category 'reproducibility' -Requirement 'Fixed testdata/control set exists' -Implemented ($controlFiles.Count -gt 0) -Conformant ($controlFiles.Count -gt 0) -Evidence $controlEvidence -Details "files=$($controlFiles.Count)"

    $solutionPath = Join-Path $Ctx.RepoRoot 'docs\reshenie.md'
    $solutionOk = (Test-Path -LiteralPath $solutionPath) -and ((Get-Item -LiteralPath $solutionPath).Length -gt 100)
    Add-BooleanFeatureAssessment -Ctx $Ctx -Id 'engineering.solution_doc' -Level 'engineering' -Category 'documentation' -Requirement 'Non-empty docs/reshenie.md exists' -Implemented $solutionOk -Conformant $solutionOk -Evidence @('repo_snapshot/docs/reshenie.md')
}

function Copy-CheckPath {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)][string]$Source,
        [Parameter(Mandatory=$true)][string]$RelativeDestination
    )

    if (-not (Test-Path -LiteralPath $Source)) {
        return
    }

    $destination = Join-Path $Ctx.ResultDir $RelativeDestination
    $parent = Split-Path -Parent $destination
    if ($parent) {
        New-Item -ItemType Directory -Force -Path $parent | Out-Null
    }
    Copy-Item -LiteralPath $Source -Destination $destination -Recurse -Force
}

function Complete-Check {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [hashtable]$Extra = @{}
    )

    Add-StandardEngineeringAssessments -Ctx $Ctx

    Invoke-CheckCommand -Ctx $Ctx -Name 'meta_git_head' -Command "git rev-parse HEAD | Set-Content -LiteralPath '$($Ctx.MetaDir)\git_head.txt' -Encoding UTF8" | Out-Null
    Invoke-CheckCommand -Ctx $Ctx -Name 'meta_git_status' -Command "git status --short | Set-Content -LiteralPath '$($Ctx.MetaDir)\git_status_short.txt' -Encoding UTF8" | Out-Null
    Invoke-CheckCommand -Ctx $Ctx -Name 'meta_go_version' -Command "& '$($Ctx.GoCmd)' version | Set-Content -LiteralPath '$($Ctx.MetaDir)\go_version.txt' -Encoding UTF8" | Out-Null
    Invoke-CheckCommand -Ctx $Ctx -Name 'meta_go_env' -Command "& '$($Ctx.GoCmd)' env GOVERSION GOOS GOARCH | Set-Content -LiteralPath '$($Ctx.MetaDir)\go_env.txt' -Encoding UTF8" | Out-Null

    foreach ($name in @('README.md', 'Makefile', 'go.mod', 'docs')) {
        $path = Join-Path $Ctx.RepoRoot $name
        Copy-CheckPath -Ctx $Ctx -Source $path -RelativeDestination "repo_snapshot/$name"
    }

    $assessmentItems = @($Ctx.Assessments)
    $assessmentSummary = [ordered]@{}
    foreach ($level in @('minimum','good','excellent','engineering')) {
        $items = @($assessmentItems | Where-Object { $_.level -eq $level })
        $assessmentSummary[$level] = [ordered]@{
            total = $items.Count
            full = @($items | Where-Object { $_.implementation -eq 'full' }).Count
            partial = @($items | Where-Object { $_.implementation -eq 'partial' }).Count
            not_implemented = @($items | Where-Object { $_.implementation -eq 'not_implemented' }).Count
            conformant = @($items | Where-Object { $_.conformance -eq 'conformant' }).Count
            nonconformant = @($items | Where-Object { $_.conformance -eq 'nonconformant' }).Count
            not_tested = @($items | Where-Object { $_.conformance -eq 'not_tested' }).Count
        }
    }
    Save-CheckJson -Path (Join-Path $Ctx.ResultDir 'assessment.json') -Value ([ordered]@{
        schema_version = 1
        statuses = [ordered]@{
            implementation = @('not_implemented','partial','full')
            conformance = @('not_tested','nonconformant','conformant')
        }
        summary = $assessmentSummary
        features = $assessmentItems
    })

    $manifest = [ordered]@{
        student = $Ctx.Student
        repo_root = $Ctx.RepoRoot
        started_at = $Ctx.StartedAt
        completed_at = (Get-Date).ToString('o')
        machine = [ordered]@{
            computer_name = $env:COMPUTERNAME
            user_name = $env:USERNAME
            os = (Get-CimInstance Win32_OperatingSystem).Caption
            powershell = $PSVersionTable.PSVersion.ToString()
        }
        result_dir = $Ctx.ResultDir
        commands_file = 'commands.jsonl'
        assessment_file = 'assessment.json'
        notes = $Extra
    }
    Save-CheckJson -Path (Join-Path $Ctx.ResultDir 'manifest.json') -Value $manifest

    $zipPath = "$($Ctx.ResultDir).zip"
    if (Test-Path -LiteralPath $zipPath) {
        Remove-Item -LiteralPath $zipPath -Force
    }
    Compress-Archive -Path (Join-Path $Ctx.ResultDir '*') -DestinationPath $zipPath -Force

    Write-Host "CHECK_RESULT_DIR=$($Ctx.ResultDir)"
    Write-Host "CHECK_RESULT_ZIP=$zipPath"
    return $zipPath
}


$ctx = New-CheckContext -Student 'memory_api_check' -RepoRoot $RepoRoot -OutRoot $OutRoot

$requestPath = Write-CheckText -Ctx $ctx -RelativePath 'inputs/search_request.json' -Content @'
{
  "dataset_id": "control",
  "time": {
    "around": "2026-06-16T10:15:00Z",
    "tolerance": "3d"
  },
  "hints": {
    "user_id": "ivan",
    "file_name": "client base",
    "action": "email",
    "destination_type": "external"
  },
  "context": {
    "before": "30m",
    "after": "10m",
    "require_nearby": [
      {"action": "create_archive"},
      {"action": "email_send"}
    ]
  },
  "limit": 20
}
'@

Invoke-CheckCommand -Ctx $ctx -Name 'go_test_all' -Command "& '$($ctx.GoCmd)' test ./..."

if (Test-Path -LiteralPath (Join-Path $ctx.RepoRoot 'Makefile')) {
    Invoke-CheckCommand -Ctx $ctx -Name 'make_test' -Command 'make test'
    Invoke-CheckCommand -Ctx $ctx -Name 'make_demo' -Command 'make demo'
}

$serverStdout = Join-Path $ctx.LogsDir 'server_stdout.log'
$serverStderr = Join-Path $ctx.LogsDir 'server_stderr.log'
$serverExe = Join-Path $ctx.OutputsDir 'event-memory-search-api.exe'
Invoke-CheckCommand -Ctx $ctx -Name 'build_api_server' -Command "& '$($ctx.GoCmd)' build -o '$serverExe' ./cmd/event-memory-search-api"

$oldPort = $env:PORT
$env:PORT = [string]$Port
$proc = Start-Process -FilePath $serverExe -WorkingDirectory $ctx.RepoRoot -RedirectStandardOutput $serverStdout -RedirectStandardError $serverStderr -PassThru -WindowStyle Hidden
$env:PORT = $oldPort
Save-CheckJson -Path (Join-Path $ctx.MetaDir 'server_process.json') -Value ([ordered]@{
    pid = $proc.Id
    port = $Port
    started_at = (Get-Date).ToString('o')
})

try {
    $baseUrl = "http://127.0.0.1:$Port"
    $waitCommand = @"
`$deadline = (Get-Date).AddSeconds(15)
do {
  try {
    Invoke-RestMethod -Uri '$baseUrl/api/health' -Method Get -TimeoutSec 2 | ConvertTo-Json -Depth 20 | Set-Content -LiteralPath '$($ctx.OutputsDir)\health.json' -Encoding UTF8
    exit 0
  } catch {
    Start-Sleep -Milliseconds 500
  }
} while ((Get-Date) -lt `$deadline)
Write-Error 'server did not become ready'
exit 1
"@
    Invoke-CheckCommand -Ctx $ctx -Name 'api_wait_health' -Command $waitCommand

    Invoke-CheckCommand -Ctx $ctx -Name 'api_datasets' -Command "Invoke-RestMethod -Uri '$baseUrl/api/datasets' -Method Get -TimeoutSec 5 | ConvertTo-Json -Depth 20 | Set-Content -LiteralPath '$($ctx.OutputsDir)\datasets.json' -Encoding UTF8"

    $searchCommand = @"
`$body = Get-Content -LiteralPath '$requestPath' -Raw
`$r = Invoke-RestMethod -Uri '$baseUrl/api/search' -Method Post -ContentType 'application/json' -Body `$body -TimeoutSec 10
`$r | ConvertTo-Json -Depth 30 | Set-Content -LiteralPath '$($ctx.OutputsDir)\search_response.json' -Encoding UTF8
if (`$r.search_id) {
  Invoke-RestMethod -Uri ('$baseUrl/api/search/' + `$r.search_id) -Method Get -TimeoutSec 10 | ConvertTo-Json -Depth 30 | Set-Content -LiteralPath '$($ctx.OutputsDir)\search_by_id.json' -Encoding UTF8
}
"@
    Invoke-CheckCommand -Ctx $ctx -Name 'api_search_flow' -Command $searchCommand

    $candidateCommand = @"
`$search = Get-Content -LiteralPath '$($ctx.OutputsDir)\search_response.json' -Raw | ConvertFrom-Json
`$eventId = if (`$search.candidates -and `$search.candidates.Count -gt 0) { `$search.candidates[0].event_id } elseif (`$search.results -and `$search.results.Count -gt 0) { `$search.results[0].event_id } else { 'evt_003' }
Invoke-RestMethod -Uri ('$baseUrl/api/events/' + `$eventId + '/context') -Method Get -TimeoutSec 10 | ConvertTo-Json -Depth 30 | Set-Content -LiteralPath '$($ctx.OutputsDir)\event_context.json' -Encoding UTF8
"@
    Invoke-CheckCommand -Ctx $ctx -Name 'api_event_context' -Command $candidateCommand

    $explainCommand = @"
`$search = Get-Content -LiteralPath '$($ctx.OutputsDir)\search_response.json' -Raw | ConvertFrom-Json
`$searchId = if (`$search.search_id) { `$search.search_id } else { 'search_stub' }
`$eventId = if (`$search.candidates -and `$search.candidates.Count -gt 0) { `$search.candidates[0].event_id } elseif (`$search.results -and `$search.results.Count -gt 0) { `$search.results[0].event_id } else { 'evt_003' }
Invoke-RestMethod -Uri ('$baseUrl/api/search/' + `$searchId + '/candidates/' + `$eventId + '/explain') -Method Get -TimeoutSec 10 | ConvertTo-Json -Depth 30 | Set-Content -LiteralPath '$($ctx.OutputsDir)\explain.json' -Encoding UTF8
"@
    Invoke-CheckCommand -Ctx $ctx -Name 'api_explain' -Command $explainCommand
}
finally {
    if ($proc -and -not $proc.HasExited) {
        Stop-Process -Id $proc.Id -Force
        try {
            Wait-Process -Id $proc.Id -Timeout 5 -ErrorAction SilentlyContinue
        } catch {
        }
        Start-Sleep -Milliseconds 500
    }
}

$expectedArtifacts = [ordered]@{
    health = Test-Path -LiteralPath (Join-Path $ctx.OutputsDir 'health.json')
    datasets = Test-Path -LiteralPath (Join-Path $ctx.OutputsDir 'datasets.json')
    search_response = Test-Path -LiteralPath (Join-Path $ctx.OutputsDir 'search_response.json')
    search_by_id = Test-Path -LiteralPath (Join-Path $ctx.OutputsDir 'search_by_id.json')
    event_context = Test-Path -LiteralPath (Join-Path $ctx.OutputsDir 'event_context.json')
    explain = Test-Path -LiteralPath (Join-Path $ctx.OutputsDir 'explain.json')
}
Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'artifact_presence.json') -Value $expectedArtifacts

Add-CommandFeatureAssessment -Ctx $ctx -Id 'minimum.health' -Level 'minimum' -Category 'api' -Requirement 'Backend starts and answers /api/health' -CommandName 'api_wait_health' -RequiredArtifacts @((Join-Path $ctx.OutputsDir 'health.json'))
Add-CommandFeatureAssessment -Ctx $ctx -Id 'minimum.datasets' -Level 'minimum' -Category 'api' -Requirement 'GET /api/datasets returns datasets' -CommandName 'api_datasets' -RequiredArtifacts @((Join-Path $ctx.OutputsDir 'datasets.json'))
Add-CommandFeatureAssessment -Ctx $ctx -Id 'minimum.search' -Level 'minimum' -Category 'api' -Requirement 'POST /api/search returns candidates' -CommandName 'api_search_flow' -RequiredArtifacts @((Join-Path $ctx.OutputsDir 'search_response.json'))
Add-CommandFeatureAssessment -Ctx $ctx -Id 'minimum.search_by_id' -Level 'minimum' -Category 'api' -Requirement 'GET /api/search/{search_id} returns result' -CommandName 'api_search_flow' -RequiredArtifacts @((Join-Path $ctx.OutputsDir 'search_by_id.json'))
Add-SourceFeatureAssessment -Ctx $ctx -Id 'minimum.scoring' -Level 'minimum' -Category 'algorithm' -Requirement 'Candidates are ranked by score and matched_hints' -Patterns @('score|Score','matched_hints|MatchedHints') -Match 'all'

Add-CommandFeatureAssessment -Ctx $ctx -Id 'good.context' -Level 'good' -Category 'api' -Requirement 'GET /api/events/{event_id}/context returns context' -CommandName 'api_event_context' -RequiredArtifacts @((Join-Path $ctx.OutputsDir 'event_context.json'))
Add-CommandFeatureAssessment -Ctx $ctx -Id 'good.explain' -Level 'good' -Category 'api' -Requirement 'Explain endpoint returns score breakdown' -CommandName 'api_explain' -RequiredArtifacts @((Join-Path $ctx.OutputsDir 'explain.json'))
Add-SourceFeatureAssessment -Ctx $ctx -Id 'good.time_filter' -Level 'good' -Category 'algorithm' -Requirement 'around and tolerance are used' -Patterns @('around|Around','tolerance|Tolerance','time\.Parse|ParseDuration') -Match 'all'
Add-SourceFeatureAssessment -Ctx $ctx -Id 'good.nearby' -Level 'good' -Category 'algorithm' -Requirement 'require_nearby and before after windows are used' -Patterns @('require_nearby|RequireNearby','before|Before','after|After') -Match 'all'
Add-SourceFeatureAssessment -Ctx $ctx -Id 'good.structured_errors' -Level 'good' -Category 'api' -Requirement 'API returns structured errors' -Patterns @('error.*json|ErrorResponse|WriteHeader') -Match 'any'

Add-SourceFeatureAssessment -Ctx $ctx -Id 'excellent.api_docs' -Level 'excellent' -Category 'documentation' -Requirement 'OpenAPI or docs/api.md exists' -Patterns @('openapi|swagger|docs/api\.md') -Match 'any'
Add-SourceFeatureAssessment -Ctx $ctx -Id 'excellent.large_dataset' -Level 'excellent' -Category 'performance' -Requirement 'Large dataset search benchmark exists' -Patterns @('Benchmark','100000') -Match 'all'
Add-SourceFeatureAssessment -Ctx $ctx -Id 'excellent.frontend_integration' -Level 'excellent' -Category 'integration' -Requirement 'Reproducible frontend backend demo exists' -Patterns @('frontend|vite|npm','demo') -Match 'all'
Add-SourceFeatureAssessment -Ctx $ctx -Id 'excellent.error_cases' -Level 'excellent' -Category 'tests' -Requirement 'Invalid time wide window and dataset errors are tested' -Patterns @('invalid.*time|bad.*time','dataset','tolerance|window') -Match 'all'

Complete-Check -Ctx $ctx -Extra @{
    expected_api = @('/api/health', '/api/datasets', 'POST /api/search', 'GET /api/search/{id}', 'GET /api/events/{id}/context', 'GET /api/search/{id}/candidates/{event_id}/explain')
    port = $Port
}


