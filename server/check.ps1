param(
    [string]$RepoRoot = (Get-Location).Path,
    [string]$OutRoot = '',
    [int]$Port = 0,
    [switch]$AllowDirtyTree,
    [switch]$IncludeMillionDataset
)

Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Net.Http -ErrorAction SilentlyContinue | Out-Null

function Get-CheckGoCommand {
    $preferred = 'K:\go\go1.20.14\bin\go.exe'
    if (Test-Path -LiteralPath $preferred) {
        return $preferred
    }
    $go = Get-Command go -ErrorAction SilentlyContinue
    if ($go) {
        return $go.Source
    }
    throw 'go executable was not found in PATH and K:\go\go1.20.14\bin\go.exe was not found'
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
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
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

function Save-CheckJson {
    param(
        [Parameter(Mandatory=$true)][string]$Path,
        [Parameter(Mandatory=$true)]$Value
    )
    $json = $Value | ConvertTo-Json -Depth 40
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
    $runnerPath = Join-Path $Ctx.TmpDir "$safeName.ps1"
    $logPath = Join-Path $Ctx.LogsDir "$safeName.log"
    $started = Get-Date
    $runner = @"
`$ErrorActionPreference = 'Stop'
Set-Location -LiteralPath '$($WorkingDirectory.Replace("'", "''"))'
try {
    `$global:LASTEXITCODE = `$null
    `$Error.Clear()
    $Command
    `$exitCode = `$global:LASTEXITCODE
    if (`$null -eq `$exitCode) {
        if (`$? -and `$Error.Count -eq 0) { `$exitCode = 0 } else { `$exitCode = 1 }
    }
    exit `$exitCode
} catch {
    Write-Error `$_
    exit 1
}
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
    return $record
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

function Add-CommandFeatureAssessment {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [Parameter(Mandatory=$true)][string]$Id,
        [Parameter(Mandatory=$true)][ValidateSet('minimum','good','excellent','engineering')][string]$Level,
        [Parameter(Mandatory=$true)][string]$Category,
        [Parameter(Mandatory=$true)][string]$Requirement,
        [Parameter(Mandatory=$true)][string]$CommandName
    )
    $has = $Ctx.CommandResults.ContainsKey($CommandName)
    $ok = $false
    if ($has) {
        $ok = ([int]$Ctx.CommandResults[$CommandName].exit_code -eq 0)
    }
    Add-BooleanFeatureAssessment -Ctx $Ctx -Id $Id -Level $Level -Category $Category -Requirement $Requirement -Implemented $has -Conformant $ok -Evidence @("logs/$CommandName.log") -Details "command=$CommandName"
}

# Set-StrictMode 2.0 превращает обращение к отсутствующему полю ConvertFrom-Json в terminating error,
# поэтому любое поле ответа читается только через эти функции. Прямое $resp.json.candidates роняет
# скрипт на решении, которое поле не отдаёт, — вместо незачёта по признаку получается падение прогона.
function Get-JsonValue {
    param(
        $Object,
        [Parameter(Mandatory=$true)][string]$Path
    )
    $current = $Object
    foreach ($segment in ($Path -split '\.')) {
        if ($null -eq $current) { return $null }
        if (@($current.PSObject.Properties.Name) -notcontains $segment) { return $null }
        $current = $current.$segment
    }
    return $current
}

function Get-JsonArray {
    param(
        $Object,
        [Parameter(Mandatory=$true)][string]$Path
    )
    $value = Get-JsonValue -Object $Object -Path $Path
    # Запятая не даёт конвейеру развернуть массив: без неё пустой результат становится $null,
    # а одноэлементный — скаляром, и обращение к .Count падает под Set-StrictMode 2.0.
    if ($null -eq $value) { return ,@() }
    return ,@($value)
}

function Test-JsonHasProperty {
    param(
        $Object,
        [Parameter(Mandatory=$true)][string]$Name
    )
    if ($null -eq $Object) { return $false }
    return (@($Object.PSObject.Properties.Name) -contains $Name)
}

function Invoke-HttpRequestSafe {
    param(
        [Parameter(Mandatory=$true)][string]$Method,
        [Parameter(Mandatory=$true)][string]$Uri,
        [string]$Body = '',
        [string]$ContentType = 'application/json',
        [int]$TimeoutSec = 10
    )
    $clientHandler = New-Object System.Net.Http.HttpClientHandler
    $client = New-Object System.Net.Http.HttpClient($clientHandler)
    $client.Timeout = [TimeSpan]::FromSeconds($TimeoutSec)
    $httpMethod = switch ($Method.ToUpperInvariant()) {
        'GET' { [System.Net.Http.HttpMethod]::Get }
        'POST' { [System.Net.Http.HttpMethod]::Post }
        'PUT' { [System.Net.Http.HttpMethod]::Put }
        'PATCH' { [System.Net.Http.HttpMethod]::new('PATCH') }
        'DELETE' { [System.Net.Http.HttpMethod]::Delete }
        default { [System.Net.Http.HttpMethod]::new($Method) }
    }
    $request = New-Object System.Net.Http.HttpRequestMessage($httpMethod, $Uri)
    if ($Method -in @('POST', 'PUT', 'PATCH')) {
        $request.Content = New-Object System.Net.Http.StringContent($Body, [System.Text.Encoding]::UTF8, $ContentType)
    }
    try {
        $response = $client.SendAsync($request).GetAwaiter().GetResult()
        $text = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
        $headers = @{}
        foreach ($header in $response.Headers) {
            $headers[[string]$header.Key] = [string]($header.Value -join ',')
        }
        foreach ($header in $response.Content.Headers) {
            $headers[[string]$header.Key] = [string]($header.Value -join ',')
        }
        $json = $null
        try { $json = $text | ConvertFrom-Json } catch {}
        return [ordered]@{
            status_code = [int]$response.StatusCode
            headers = $headers
            body = $text
            json = $json
            transport_error = ''
        }
    } catch {
        # Транспортная ошибка (отказ соединения, таймаут, обрыв) возвращается как ответ со статусом 0,
        # а не выбрасывается: иначе один зависший endpoint обрывает прогон целиком вместо незачёта по
        # этому признаку, и ожидание готовности сервера падает на первой же неудачной попытке.
        return [ordered]@{
            status_code = 0
            headers = @{}
            body = ''
            json = $null
            transport_error = [string]$_.Exception.Message
        }
    } finally {
        if ($request) { $request.Dispose() }
        if ($client) { $client.Dispose() }
        if ($clientHandler) { $clientHandler.Dispose() }
    }
}

# Задание фиксирует единый формат ошибки {"error":{"code","message","field"}} и минимальный набор из
# шести кодов, но не фиксирует HTTP-статусы и не запрещает дополнительные коды. Поэтому проверяем
# 4xx, конверт и попадание кода в список допустимых для сценария, а фактические значения возвращаем
# в probe, чтобы отличать «код не из набора» от «код из набора, но статус другой».
function Test-ErrorEnvelope {
    param(
        [Parameter(Mandatory=$true)]$Response,
        [Parameter(Mandatory=$true)][string[]]$AcceptedCodes,
        [string]$ExpectedField = '',
        [int]$ExpectedStatus = 0
    )
    $contentType = ''
    if ($Response.headers) {
        if ($Response.headers.ContainsKey('Content-Type')) {
            $contentType = [string]$Response.headers['Content-Type']
        } elseif ($Response.headers.ContainsKey('content-type')) {
            $contentType = [string]$Response.headers['content-type']
        }
    }
    $hasJsonType = $true
    if (-not [string]::IsNullOrWhiteSpace($contentType)) {
        $hasJsonType = $contentType.ToLower().StartsWith('application/json')
    }
    $payload = $Response.json
    if (-not $payload -and -not [string]::IsNullOrWhiteSpace([string]$Response.body)) {
        try { $payload = [string]$Response.body | ConvertFrom-Json } catch {}
    }
    $code = [string](Get-JsonValue -Object $payload -Path 'error.code')
    $message = [string](Get-JsonValue -Object $payload -Path 'error.message')
    $field = [string](Get-JsonValue -Object $payload -Path 'error.field')
    $codeOk = ($AcceptedCodes -contains $code)
    $fieldOk = $true
    if (-not [string]::IsNullOrWhiteSpace($ExpectedField)) {
        # Пример в задании использует полный путь ("time.tolerance"), поэтому принимаем и короткое имя.
        $fieldOk = (-not [string]::IsNullOrWhiteSpace($field)) -and (($field -eq $ExpectedField) -or ($field -eq ($ExpectedField -split '\.')[-1]))
    }
    $statusOk = ($Response.status_code -ge 400 -and $Response.status_code -lt 500)
    if ($ExpectedStatus -gt 0) { $statusOk = $statusOk -and ($Response.status_code -eq $ExpectedStatus) }
    return [ordered]@{
        status_code = [int]$Response.status_code
        status_4xx = ($Response.status_code -ge 400 -and $Response.status_code -lt 500)
        json_content_type = $hasJsonType
        code = $code
        code_accepted = $codeOk
        message_present = (-not [string]::IsNullOrWhiteSpace($message))
        field = $field
        field_ok = $fieldOk
        ok = ($statusOk -and $hasJsonType -and $codeOk -and (-not [string]::IsNullOrWhiteSpace($message)) -and $fieldOk)
    }
}

function Get-FreeTcpPort {
    $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, 0)
    $listener.Start()
    $port = $listener.LocalEndpoint.Port
    $listener.Stop()
    return [int]$port
}

function New-GeneratedDatasetFile {
    param(
        [Parameter(Mandatory=$true)][string]$Path,
        [int]$Count = 100000,
        [string]$TargetEventId = 'large_target_054321',
        [int]$TargetIndex = 54321
    )
    $started = Get-Date
    $utf8 = New-Object System.Text.UTF8Encoding($false)
    $writer = New-Object System.IO.StreamWriter($Path, $false, $utf8)
    try {
        $base = [DateTime]::Parse('2026-06-16T00:00:00Z').ToUniversalTime()
        for ($i = 0; $i -lt $Count; $i++) {
            $eventId = ('evt_large_{0:D7}' -f $i)
            $userId = ('user_{0:D3}' -f ($i % 100))
            $fileName = ('file_{0:D7}' -f $i)
            $action = if (($i % 7) -eq 0) { 'email_send' } else { 'read' }
            $destination = if (($i % 11) -eq 0) { 'external' } else { 'internal' }
            if ($i -eq $TargetIndex) {
                $eventId = $TargetEventId
                $userId = 'target_user'
                $fileName = ('file_{0}' -f $TargetEventId)
                $action = 'email_send'
                $destination = 'external'
            }
            $timestamp = $base.AddSeconds($i).ToString('yyyy-MM-ddTHH:mm:ssZ')
            $line = ('{{"event_id":"{0}","user_id":"{1}","file_name":"{2}","action":"{3}","destination_type":"{4}","timestamp":"{5}"}}' -f $eventId, $userId, $fileName, $action, $destination, $timestamp)
            $writer.WriteLine($line)
        }
        $writer.Flush()
    } finally {
        $writer.Dispose()
    }
    $ended = Get-Date
    return [ordered]@{
        count = $Count
        bytes = [int64](Get-Item -LiteralPath $Path).Length
        duration_ms = [int](($ended - $started).TotalMilliseconds)
        target_event_id = $TargetEventId
    }
}

# Контрольная фикстура задаёт идентификаторы, на которые опираются probes: evt_exact — полное
# совпадение по всем подсказкам, evt_boundary — ровно на границе окна tolerance=30m от 10:15,
# evt_outside — вне этого окна, evt_archive и evt_mail_second — соседние события для require_nearby
# в окне before=30m/after=10m и вне окна before=5m/after=1m.
function Get-ControlFixtureEvents {
    return @(
        [ordered]@{ event_id = 'evt_archive';     timestamp = '2026-06-16T09:50:00Z'; user_id = 'ivanov';  machine_id = 'pc_003'; action = 'create_archive'; channel = 'fs';    file_name = 'client_base.zip';       file_ext = 'zip';  content_classes = @('client_data');                 destination_type = 'internal'; destination = 'local_disk';         severity = 'medium' },
        [ordered]@{ event_id = 'evt_exact';       timestamp = '2026-06-16T10:15:00Z'; user_id = 'ivanov';  machine_id = 'pc_003'; action = 'email_send';     channel = 'email'; file_name = 'client_base.xlsx';      file_ext = 'xlsx'; content_classes = @('client_data','personal_data'); destination_type = 'external'; destination = 'external_email_001'; severity = 'high' },
        [ordered]@{ event_id = 'evt_no_hints';    timestamp = '2026-06-16T10:20:00Z'; user_id = 'sidorov'; machine_id = 'pc_011'; action = 'read';           channel = 'fs';    file_name = 'notes.txt';             file_ext = 'txt';  content_classes = @();                              destination_type = 'internal'; destination = 'local_disk';         severity = 'low' },
        [ordered]@{ event_id = 'evt_mail_second'; timestamp = '2026-06-16T10:22:00Z'; user_id = 'ivanov';  machine_id = 'pc_003'; action = 'email_send';     channel = 'email'; file_name = 'quarter_report.pdf';    file_ext = 'pdf';  content_classes = @('internal_data');               destination_type = 'external'; destination = 'external_email_002'; severity = 'medium' },
        [ordered]@{ event_id = 'evt_partial';     timestamp = '2026-06-16T10:30:00Z'; user_id = 'ivanov2'; machine_id = 'pc_007'; action = 'email_send';     channel = 'email'; file_name = 'base_client.csv';       file_ext = 'csv';  content_classes = @('client_data');                 destination_type = 'internal'; destination = 'internal_email_003'; severity = 'medium' },
        [ordered]@{ event_id = 'evt_boundary';    timestamp = '2026-06-16T10:45:00Z'; user_id = 'ivanova'; machine_id = 'pc_004'; action = 'email_send';     channel = 'email'; file_name = 'client_base_copy.xlsx'; file_ext = 'xlsx'; content_classes = @('client_data');                 destination_type = 'external'; destination = 'external_email_004'; severity = 'high' },
        [ordered]@{ event_id = 'evt_outside';     timestamp = '2026-06-20T10:15:00Z'; user_id = 'petrov';  machine_id = 'pc_009'; action = 'email_send';     channel = 'email'; file_name = 'client_base_old.xlsx';  file_ext = 'xlsx'; content_classes = @('client_data');                 destination_type = 'external'; destination = 'external_email_005'; severity = 'low' }
    )
}

function Get-TestFixtureEvents {
    return @(
        [ordered]@{ event_id = 'evt_test_1'; timestamp = '2026-06-20T11:00:00Z'; user_id = 'tester'; machine_id = 'pc_100'; action = 'read';       channel = 'fs';    file_name = 'a.txt'; file_ext = 'txt'; content_classes = @();          destination_type = 'internal'; destination = 'local_disk';         severity = 'low' },
        [ordered]@{ event_id = 'evt_test_2'; timestamp = '2026-06-20T11:20:00Z'; user_id = 'tester'; machine_id = 'pc_100'; action = 'email_send'; channel = 'email'; file_name = 'b.pdf'; file_ext = 'pdf'; content_classes = @('other'); destination_type = 'external'; destination = 'external_email_900'; severity = 'medium' },
        [ordered]@{ event_id = 'evt_test_3'; timestamp = '2026-06-20T11:40:00Z'; user_id = 'tester'; machine_id = 'pc_100'; action = 'read';       channel = 'fs';    file_name = 'c.txt'; file_ext = 'txt'; content_classes = @();          destination_type = 'internal'; destination = 'local_disk';         severity = 'low' }
    )
}

function Write-JsonLinesFile {
    param(
        [Parameter(Mandatory=$true)][string]$Path,
        [Parameter(Mandatory=$true)]$Events
    )
    $utf8 = New-Object System.Text.UTF8Encoding($false)
    $writer = New-Object System.IO.StreamWriter($Path, $false, $utf8)
    try {
        foreach ($event in @($Events)) {
            $writer.WriteLine(($event | ConvertTo-Json -Depth 5 -Compress))
        }
        $writer.Flush()
    } finally {
        $writer.Dispose()
    }
    return [int64](Get-Item -LiteralPath $Path).Length
}

# Имена файлов заданы реализацией: cmd/event-memory-search-api/main.go читает из каталога --datasets
# строго events.jsonl и testEvents.jsonl (обязательные, при ошибке log.Fatal), а также опциональные
# events100k.jsonl и events1m.jsonl. Каталог с другими именами приводит к немедленному выходу сервера.
function New-CheckDatasetFiles {
    param(
        [Parameter(Mandatory=$true)][string]$Directory,
        [switch]$WithMillion
    )
    New-Item -ItemType Directory -Path $Directory -Force | Out-Null
    $control = Get-ControlFixtureEvents
    $test = Get-TestFixtureEvents
    $meta = [ordered]@{
        directory = $Directory
        control_file = 'events.jsonl'
        control_count = @($control).Count
        control_bytes = Write-JsonLinesFile -Path (Join-Path $Directory 'events.jsonl') -Events $control
        control_event_ids = @($control | ForEach-Object { $_.event_id })
        test_file = 'testEvents.jsonl'
        test_count = @($test).Count
        test_bytes = Write-JsonLinesFile -Path (Join-Path $Directory 'testEvents.jsonl') -Events $test
        large100k = New-GeneratedDatasetFile -Path (Join-Path $Directory 'events100k.jsonl') -Count 100000
        large1m = $null
    }
    if ($WithMillion) {
        $meta.large1m = New-GeneratedDatasetFile -Path (Join-Path $Directory 'events1m.jsonl') -Count 1000000 -TargetEventId 'million_target_654321' -TargetIndex 654321
    }
    return $meta
}

# Ответ /api/datasets заданием по составу полей не зафиксирован: страница требует только «Список
# доступных наборов событий». Нормализуем и массив, и объект-конверт, и оба распространённых имени
# полей, чтобы не превращать выбор имени в незачёт.
function Get-DatasetEntries {
    param([Parameter(Mandatory=$true)]$Response)
    $payload = $Response.json
    if (-not $payload) { return @() }
    $raw = @()
    if ($payload -is [System.Collections.IEnumerable] -and $payload -isnot [string]) {
        $raw = @($payload)
    } elseif (Test-JsonHasProperty -Object $payload -Name 'datasets') {
        $raw = Get-JsonArray -Object $payload -Path 'datasets'
    } elseif (Test-JsonHasProperty -Object $payload -Name 'value') {
        $raw = Get-JsonArray -Object $payload -Path 'value'
    } else {
        $raw = @($payload)
    }
    $entries = @()
    foreach ($item in $raw) {
        if (-not $item) { continue }
        $id = ''
        foreach ($key in @('id','dataset_id','name')) {
            $value = Get-JsonValue -Object $item -Path $key
            if (-not [string]::IsNullOrWhiteSpace([string]$value)) {
                $id = [string]$value
                break
            }
        }
        $size = -1
        foreach ($key in @('size','event_count','total_events','count')) {
            $value = Get-JsonValue -Object $item -Path $key
            if ($null -ne $value) {
                $size = [int]$value
                break
            }
        }
        if ($id) { $entries += [ordered]@{ id = $id; size = $size } }
    }
    return ,@($entries)
}

# Идентификатор набора реализация выбирает сама: фиксированные имена дают control/test/large100k,
# сканирование каталога — имя файла без расширения. Поэтому id определяется по имени с запасным
# вариантом по размеру набора, а не задаётся в скрипте константой.
function Resolve-DatasetId {
    param(
        $Entries,
        [string[]]$PreferredIds = @(),
        [int]$ExpectedSize = -1
    )
    foreach ($name in $PreferredIds) {
        $hit = @($Entries | Where-Object { $_.id -eq $name }) | Select-Object -First 1
        if ($hit) { return [string]$hit.id }
    }
    if ($ExpectedSize -ge 0) {
        $bySize = @($Entries | Where-Object { [int]$_.size -eq $ExpectedSize }) | Select-Object -First 1
        if ($bySize) { return [string]$bySize.id }
    }
    return ''
}

# Тела запросов собираются из структуры, а не из строковых шаблонов: идентификатор набора и порог
# min_score определяются в рантайме, а ручная подстановка в JSON-текст легко даёт невалидное тело.
function New-SearchRequestBody {
    param(
        [Parameter(Mandatory=$true)][string]$DatasetId,
        # Целое намеренно: ConvertTo-Json сериализует [double] 0 как 0.0, и Go отказывается разбирать
        # 0.0 в целочисленное поле — запрос отвергается как невалидный JSON.
        [int]$MinScore = 0,
        [int]$Limit = 20,
        [string]$Around = '',
        [string]$Tolerance = '',
        $Hints = $null,
        [string]$Before = '',
        [string]$After = '',
        [string[]]$RequireNearby = @()
    )
    $request = [ordered]@{ dataset_id = $DatasetId }
    if (-not [string]::IsNullOrWhiteSpace($Around)) {
        $time = [ordered]@{ around = $Around }
        if (-not [string]::IsNullOrWhiteSpace($Tolerance)) { $time['tolerance'] = $Tolerance }
        $request['time'] = $time
    }
    if ($Hints) { $request['hints'] = $Hints }
    if ((-not [string]::IsNullOrWhiteSpace($Before)) -or (-not [string]::IsNullOrWhiteSpace($After)) -or (@($RequireNearby).Count -gt 0)) {
        $context = [ordered]@{}
        if (-not [string]::IsNullOrWhiteSpace($Before)) { $context['before'] = $Before }
        if (-not [string]::IsNullOrWhiteSpace($After)) { $context['after'] = $After }
        if (@($RequireNearby).Count -gt 0) {
            $context['require_nearby'] = @(@($RequireNearby) | ForEach-Object { [ordered]@{ action = [string]$_ } })
        }
        $request['context'] = $context
    }
    $request['scoring'] = [ordered]@{ min_score = $MinScore; limit = $Limit }
    return ($request | ConvertTo-Json -Depth 6 -Compress)
}

function Test-HintMentioned {
    param(
        [Parameter(Mandatory=$true)]$MatchedHints,
        [Parameter(Mandatory=$true)][string]$Hint
    )
    foreach ($entry in @($MatchedHints)) {
        if ([string]$entry -match [regex]::Escape($Hint)) { return $true }
    }
    return $false
}

function Get-ContributionSum {
    param([Parameter(Mandatory=$true)]$Contributions)
    $sum = 0.0
    foreach ($part in @($Contributions)) {
        $points = Get-JsonValue -Object $part -Path 'points'
        if ($null -ne $points) { $sum += [double]$points }
    }
    return $sum
}

# Формат ответа контекста заданием не зафиксирован, кроме смысла «события вокруг». Поэтому контекст
# проверяется по составу окон: соседи из контрольной фикстуры обязаны присутствовать, а посторонних
# событий быть не должно. Именно это отличает верный набор от одноимённого события из другого набора.
function Get-ExpectedNeighbours {
    param(
        [Parameter(Mandatory=$true)]$Fixture,
        [Parameter(Mandatory=$true)][string]$AnchorEventId,
        [Parameter(Mandatory=$true)][timespan]$Before,
        [Parameter(Mandatory=$true)][timespan]$After
    )
    $anchor = @($Fixture | Where-Object { [string]$_.event_id -eq $AnchorEventId }) | Select-Object -First 1
    if (-not $anchor) { return [ordered]@{ before = @(); after = @() } }
    $anchorTime = ([DateTime]::Parse([string]$anchor.timestamp)).ToUniversalTime()
    # Имена накопителей не должны совпадать с параметрами $Before/$After: имена переменных в
    # PowerShell регистронезависимы, и присваивание затирает переданный timespan.
    $beforeIDs = @()
    $afterIDs = @()
    foreach ($event in @($Fixture)) {
        if ([string]$event.event_id -eq $AnchorEventId) { continue }
        $time = ([DateTime]::Parse([string]$event.timestamp)).ToUniversalTime()
        if ($time -lt $anchorTime -and $time -ge $anchorTime.Subtract($Before)) { $beforeIDs += [string]$event.event_id }
        if ($time -gt $anchorTime -and $time -le $anchorTime.Add($After)) { $afterIDs += [string]$event.event_id }
    }
    return [ordered]@{ before = $beforeIDs; after = $afterIDs }
}

function Get-ContextEvaluation {
    param(
        [Parameter(Mandatory=$true)]$Response,
        [Parameter(Mandatory=$true)][string[]]$FixtureIDs,
        [Parameter(Mandatory=$true)]$ExpectedNeighbours,
        $ExpectedEvent = $null
    )
    $eventValue = Get-JsonValue -Object $Response.json -Path 'event'
    $eventObject = if ($null -ne $eventValue) { $eventValue } else { $Response.json }
    $beforeIDs = @((Get-JsonArray -Object $Response.json -Path 'before') | ForEach-Object { [string](Get-JsonValue -Object $_ -Path 'event_id') })
    $afterIDs = @((Get-JsonArray -Object $Response.json -Path 'after') | ForEach-Object { [string](Get-JsonValue -Object $_ -Path 'event_id') })
    $returned = @(@($beforeIDs) + @($afterIDs) | Where-Object { -not [string]::IsNullOrWhiteSpace([string]$_) })
    $foreign = @($returned | Where-Object { $FixtureIDs -notcontains $_ })
    $expectedAll = @(@($ExpectedNeighbours.before) + @($ExpectedNeighbours.after))
    $missing = @($expectedAll | Where-Object { $returned -notcontains $_ })
    return [ordered]@{
        status_code = [int]$Response.status_code
        returned_event_id = [string](Get-JsonValue -Object $eventObject -Path 'event_id')
        before_ids = $beforeIDs
        after_ids = $afterIDs
        foreign_ids = $foreign
        missing_expected_ids = $missing
        neighbours_ok = ($foreign.Count -eq 0 -and $missing.Count -eq 0 -and $returned.Count -gt 0)
        event_field_match = if ($ExpectedEvent -and $null -ne $eventValue) { Test-SameEvent -Expected $ExpectedEvent -Actual $eventObject } else { $null }
    }
}

function Test-SameEvent {
    param(
        [Parameter(Mandatory=$true)]$Expected,
        [Parameter(Mandatory=$true)]$Actual
    )
    $result = [ordered]@{ event_id = $false; user_id = $false; action = $false; file_name = $false; timestamp = $false }
    if (-not $Actual) { return $result }
    foreach ($field in @('event_id','user_id','action','file_name')) {
        $result[$field] = ([string]$Expected.$field -eq [string](Get-JsonValue -Object $Actual -Path $field))
    }
    $expectedTime = [DateTime]::MinValue
    $actualTime = [DateTime]::MinValue
    [void][DateTime]::TryParse([string]$Expected.timestamp, [ref]$expectedTime)
    [void][DateTime]::TryParse([string](Get-JsonValue -Object $Actual -Path 'timestamp'), [ref]$actualTime)
    if ($expectedTime -ne [DateTime]::MinValue -and $actualTime -ne [DateTime]::MinValue) {
        $result.timestamp = ($expectedTime.ToUniversalTime() -eq $actualTime.ToUniversalTime())
    }
    return $result
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
        New-Item -ItemType Directory -Path $parent -Force | Out-Null
    }
    Copy-Item -LiteralPath $Source -Destination $destination -Recurse -Force
}

function Add-StandardEngineeringAssessments {
    param([Parameter(Mandatory=$true)]$Ctx)
    $testFiles = @(Get-ChildItem -LiteralPath $Ctx.RepoRoot -Recurse -File -Filter '*_test.go' -ErrorAction SilentlyContinue)
    $testFunctions = @($testFiles | Select-String -Pattern '^\s*func\s+Test[A-Za-z0-9_]+\s*\(' -ErrorAction SilentlyContinue)
    $benchmarkFunctions = @($testFiles | Select-String -Pattern '^\s*func\s+Benchmark[A-Za-z0-9_]+\s*\(' -ErrorAction SilentlyContinue)
    Add-BooleanFeatureAssessment -Ctx $Ctx -Id 'engineering.unit_tests_present' -Level 'engineering' -Category 'tests' -Requirement 'Go unit tests are present' -Implemented ($testFunctions.Count -gt 0) -Conformant ($testFunctions.Count -gt 0) -Evidence @('cmd/event-memory-search-api/main_test.go') -Details "tests=$($testFunctions.Count)"
    Add-BooleanFeatureAssessment -Ctx $Ctx -Id 'engineering.benchmarks_present' -Level 'engineering' -Category 'benchmarks' -Requirement 'Go benchmarks are present' -Implemented ($benchmarkFunctions.Count -gt 0) -Conformant ($benchmarkFunctions.Count -gt 0) -Evidence @('cmd/event-memory-search-api/main_test.go') -Details "benchmarks=$($benchmarkFunctions.Count)"

    foreach ($pair in @(
        @{ id = 'engineering.gofmt_runs'; cmd = 'go_fmt' ; req = 'gofmt command passes' },
        @{ id = 'engineering.go_test_passes'; cmd = 'go_test_all'; req = 'go test ./... passes' },
        @{ id = 'engineering.make_test_runs'; cmd = 'make_test'; req = 'make test passes' },
        @{ id = 'engineering.make_bench_runs'; cmd = 'make_bench'; req = 'make bench passes' },
        @{ id = 'engineering.make_demo_runs'; cmd = 'make_demo'; req = 'make demo passes' },
        @{ id = 'engineering.build_server'; cmd = 'build_api_server'; req = 'go build server passes' }
    )) {
        if ($Ctx.CommandResults.ContainsKey($pair.cmd)) {
            Add-CommandFeatureAssessment -Ctx $Ctx -Id $pair.id -Level 'engineering' -Category 'reproducibility' -Requirement $pair.req -CommandName $pair.cmd
        }
    }
    if ($Ctx.CommandResults.ContainsKey('go_test_race')) {
        Add-CommandFeatureAssessment -Ctx $Ctx -Id 'engineering.race_test_passes' -Level 'engineering' -Category 'tests' -Requirement 'go test -race ./... passes' -CommandName 'go_test_race'
    }

    $readmePath = Join-Path $Ctx.RepoRoot 'README.md'
    $readmeOk = (Test-Path -LiteralPath $readmePath) -and ((Get-Item -LiteralPath $readmePath).Length -gt 100)
    Add-BooleanFeatureAssessment -Ctx $Ctx -Id 'engineering.readme' -Level 'engineering' -Category 'documentation' -Requirement 'README.md exists and is not empty' -Implemented $readmeOk -Conformant $readmeOk -Evidence @('repo_snapshot/README.md')

    $makefilePath = Join-Path $Ctx.RepoRoot 'Makefile'
    $makefileText = if (Test-Path -LiteralPath $makefilePath) { Get-Content -LiteralPath $makefilePath -Raw } else { '' }
    foreach ($target in @('test','bench','demo','serve')) {
        $targetOk = $makefileText -match "(?m)^\s*${target}\s*:"
        Add-BooleanFeatureAssessment -Ctx $Ctx -Id "engineering.make_$target" -Level 'engineering' -Category 'reproducibility' -Requirement "Makefile has target $target" -Implemented $targetOk -Conformant $targetOk -Evidence @('repo_snapshot/Makefile')
    }

    $controlPath = Join-Path $Ctx.RepoRoot 'testdata\control'
    $controlFiles = @()
    if (Test-Path -LiteralPath $controlPath) {
        $controlFiles = @(Get-ChildItem -LiteralPath $controlPath -Recurse -File -ErrorAction SilentlyContinue)
    }
    Add-BooleanFeatureAssessment -Ctx $Ctx -Id 'engineering.control_data' -Level 'engineering' -Category 'reproducibility' -Requirement 'Fixed testdata/control set exists' -Implemented ($controlFiles.Count -gt 0) -Conformant ($controlFiles.Count -gt 0) -Evidence @($controlFiles | ForEach-Object { $_.FullName }) -Details "files=$($controlFiles.Count)"

    $solutionPath = Join-Path $Ctx.RepoRoot 'docs\reshenie.md'
    $solutionOk = (Test-Path -LiteralPath $solutionPath) -and ((Get-Item -LiteralPath $solutionPath).Length -gt 100)
    Add-BooleanFeatureAssessment -Ctx $Ctx -Id 'engineering.solution_doc' -Level 'engineering' -Category 'documentation' -Requirement 'Non-empty docs/reshenie.md exists' -Implemented $solutionOk -Conformant $solutionOk -Evidence @('repo_snapshot/docs/reshenie.md')
}

function Complete-Check {
    param(
        [Parameter(Mandatory=$true)]$Ctx,
        [hashtable]$Extra = @{}
    )
    Add-StandardEngineeringAssessments -Ctx $Ctx

    Invoke-CheckCommand -Ctx $Ctx -Name 'meta_git_head' -Command "git rev-parse HEAD | Set-Content -LiteralPath '$($Ctx.MetaDir)\git_head.txt' -Encoding UTF8" | Out-Null
    Invoke-CheckCommand -Ctx $Ctx -Name 'meta_git_status' -Command "`$statusPath = '$($Ctx.MetaDir)\git_status_short.txt'; `$status = git status --short; if (`$LASTEXITCODE -ne 0) { exit `$LASTEXITCODE }; if (`$null -eq `$status) { '' | Set-Content -LiteralPath `$statusPath -Encoding UTF8 } else { @(`$status) | Set-Content -LiteralPath `$statusPath -Encoding UTF8 }" | Out-Null
    Invoke-CheckCommand -Ctx $Ctx -Name 'meta_go_version' -Command "& '$($Ctx.GoCmd)' version | Set-Content -LiteralPath '$($Ctx.MetaDir)\go_version.txt' -Encoding UTF8" | Out-Null
    Invoke-CheckCommand -Ctx $Ctx -Name 'meta_go_env' -Command "& '$($Ctx.GoCmd)' env GOVERSION GOOS GOARCH CGO_ENABLED | Set-Content -LiteralPath '$($Ctx.MetaDir)\go_env.txt' -Encoding UTF8" | Out-Null

    foreach ($name in @('README.md', 'Makefile', 'go.mod', 'docs', 'testdata')) {
        Copy-CheckPath -Ctx $Ctx -Source (Join-Path $Ctx.RepoRoot $name) -RelativeDestination "repo_snapshot/$name"
    }

    $assessmentItems = @($Ctx.Assessments)
    $summary = [ordered]@{}
    foreach ($level in @('minimum','good','excellent','engineering')) {
        $items = @($assessmentItems | Where-Object { $_.level -eq $level })
        $summary[$level] = [ordered]@{
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
        summary = $summary
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

# Прогон на грязном дереве не привязан к коммиту: результаты нельзя сопоставить с проверяемым кодом.
# Поэтому состояние дерева снимается до любых команд и по умолчанию блокирует прогон.
Push-Location -LiteralPath $ctx.RepoRoot
try {
    $headSha = (git rev-parse HEAD 2>$null | Select-Object -First 1)
    # Гейт смотрит только на каталог проверки: в парном проекте правки frontend не влияют на
    # привязку результатов backend. Состояние всего репозитория остаётся в архиве справочно.
    $scopedStatus = @(git status --short -- . 2>$null | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
    $repoStatus = @(git status --short 2>$null | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
} finally {
    Pop-Location
}
$dirtyLines = @($scopedStatus | Where-Object { $_ -notmatch '(^|\s)\.check-results/?$' })
$isDirty = ($dirtyLines.Count -gt 0)
$attribution = [ordered]@{
    head = [string]$headSha
    dirty = $isDirty
    dirty_entries = $dirtyLines
    repo_wide_entry_count = $repoStatus.Count
    allow_dirty_tree = [bool]$AllowDirtyTree
    attributable = (-not $isDirty)
}
Save-CheckJson -Path (Join-Path $ctx.MetaDir 'run_attribution.json') -Value $attribution
if ($isDirty -and -not $AllowDirtyTree) {
    Write-Host 'CHECK_ABORTED=dirty_worktree'
    Write-Host "Рабочее дерево не чистое ($($dirtyLines.Count) записей в $($ctx.RepoRoot)), прогон не привязать к коммиту."
    Write-Host 'Закоммитьте и опубликуйте изменения, затем повторите прогон.'
    Write-Host 'Осознанный прогон на грязном дереве: повторите с -AllowDirtyTree (архив будет помечен как непривязанный).'
    Write-Host (@($dirtyLines | Select-Object -First 20) -join [Environment]::NewLine)
    if ($dirtyLines.Count -gt 20) { Write-Host "... и ещё $($dirtyLines.Count - 20); полный список в meta/run_attribution.json" }
    exit 2
}

Invoke-CheckCommand -Ctx $ctx -Name 'go_fmt' -Command "& '$($ctx.GoCmd)' fmt ./..." | Out-Null
Invoke-CheckCommand -Ctx $ctx -Name 'go_test_all' -Command "& '$($ctx.GoCmd)' test ./..." | Out-Null
if (Test-Path -LiteralPath (Join-Path $ctx.RepoRoot 'Makefile')) {
    Invoke-CheckCommand -Ctx $ctx -Name 'make_test' -Command 'make test' | Out-Null
    Invoke-CheckCommand -Ctx $ctx -Name 'make_bench' -Command 'make bench' | Out-Null
    Invoke-CheckCommand -Ctx $ctx -Name 'make_demo' -Command 'make demo' | Out-Null
}
$cgoEnabled = (& $ctx.GoCmd env CGO_ENABLED).Trim()
if ($cgoEnabled -eq '1') {
    Invoke-CheckCommand -Ctx $ctx -Name 'go_test_race' -Command "& '$($ctx.GoCmd)' test -race ./..." | Out-Null
}

$serverExe = Join-Path $ctx.OutputsDir 'event-memory-search-api.exe'
Invoke-CheckCommand -Ctx $ctx -Name 'build_api_server' -Command "& '$($ctx.GoCmd)' build -o '$serverExe' ./cmd/event-memory-search-api" | Out-Null

$probes = [ordered]@{}
$probes['minimum.scoring'] = [ordered]@{ implemented = $false; conformant = $false; details = ''; evidence = 'outputs/minimum_scoring.json' }
$probes['minimum.result_format'] = [ordered]@{ implemented = $false; conformant = $false; details = ''; evidence = 'outputs/minimum_result_format.json' }
$probes['engineering.cli_search'] = [ordered]@{ implemented = $false; conformant = $false; details = ''; evidence = 'outputs/cli_search.json' }
$probes['good.time_filter'] = [ordered]@{ implemented = $false; conformant = $false; details = ''; evidence = 'outputs/good_time_filter.json' }
$probes['good.nearby'] = [ordered]@{ implemented = $false; conformant = $false; details = ''; evidence = 'outputs/good_nearby.json' }
$probes['good.structured_errors'] = [ordered]@{ implemented = $false; conformant = $false; details = ''; evidence = 'outputs/good_structured_errors.json' }
$probes['good.error_fields'] = [ordered]@{ implemented = $false; conformant = $false; details = ''; evidence = 'outputs/good_error_fields.json' }
$probes['excellent.large_dataset'] = [ordered]@{ implemented = $false; conformant = $false; details = ''; evidence = 'outputs/excellent_large_dataset.json' }
$probes['excellent.memory_million'] = [ordered]@{ implemented = $false; conformant = $false; details = ''; evidence = 'outputs/excellent_memory_million.json' }

$healthImplemented = $false
$healthConformant = $false
$datasetsImplemented = $false
$datasetsConformant = $false
$searchImplemented = $false
$searchConformant = $false
$searchByIDImplemented = $false
$searchByIDConformant = $false
$contextImplemented = $false
$contextConformant = $false
$explainImplemented = $false
$explainConformant = $false
# Инициализация на верхнем уровне: при неудачной сборке рантайм-блок не выполняется, а итоговые
# Add-BooleanFeatureAssessment читают эти переменные, и Set-StrictMode 2.0 упал бы на необъявленной.
$scoringProbe = [ordered]@{}
$formatProbe = [ordered]@{}
$contextProbe = [ordered]@{}
$explainProbe = [ordered]@{}
$runtimeError = $null

$tempDatasetsDir = Join-Path $ctx.TmpDir 'datasets'
$runtimePort = if ($Port -eq 0) { Get-FreeTcpPort } else { $Port }
$serverStdout = Join-Path $ctx.LogsDir 'server_stdout.log'
$serverStderr = Join-Path $ctx.LogsDir 'server_stderr.log'
$serverProc = $null
$datasetMeta = $null
$largeMeta = $null
$portOwnership = [ordered]@{ checked = $false; free_before_start = $null; listening_pids = @(); owned_by_server = $null }
$controlFixture = Get-ControlFixtureEvents
$controlExpected = @{}
foreach ($event in $controlFixture) { $controlExpected[[string]$event.event_id] = $event }

if ([int]$ctx.CommandResults['build_api_server'].exit_code -eq 0) {
    $portOwnership.free_before_start = -not (@(Get-NetTCPConnection -LocalPort $runtimePort -State Listen -ErrorAction SilentlyContinue).Count -gt 0)
    if (-not $portOwnership.free_before_start) {
        Save-CheckJson -Path (Join-Path $ctx.MetaDir 'port_ownership.json') -Value $portOwnership
        throw "port $runtimePort is already in use before start; refusing to run so that responses stay attributable"
    }

    $datasetMeta = New-CheckDatasetFiles -Directory $tempDatasetsDir -WithMillion:$IncludeMillionDataset
    $largeMeta = $datasetMeta.large100k
    Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'large_dataset_generation.json') -Value $datasetMeta

    $serverArgs = @('serve', '--datasets', $tempDatasetsDir, '--addr', "127.0.0.1:$runtimePort")
    $serverProc = Start-Process -FilePath $serverExe -ArgumentList $serverArgs -WorkingDirectory $ctx.RepoRoot -RedirectStandardOutput $serverStdout -RedirectStandardError $serverStderr -PassThru -WindowStyle Hidden
    Save-CheckJson -Path (Join-Path $ctx.MetaDir 'server_process.json') -Value ([ordered]@{
        pid = $serverProc.Id
        port = $runtimePort
        args = $serverArgs
        started_at = (Get-Date).ToString('o')
    })

    $baseUrl = "http://127.0.0.1:$runtimePort"
    try {
        # Ожидание рассчитано на загрузку сгенерированных наборов: 100 000 событий всегда,
        # 1 000 000 — при -IncludeMillionDataset.
        $readyTimeoutSec = if ($IncludeMillionDataset) { 180 } else { 60 }
        $ready = $false
        $healthResp = $null
        $deadline = (Get-Date).AddSeconds($readyTimeoutSec)
        while (-not $ready -and (Get-Date) -lt $deadline) {
            if ($serverProc.HasExited) {
                throw "server process exited with code $($serverProc.ExitCode) before becoming ready; see logs/server_stderr.log"
            }
            $healthResp = Invoke-HttpRequestSafe -Method 'GET' -Uri "$baseUrl/api/health" -TimeoutSec 2
            if ($healthResp.status_code -eq 200 -and [string](Get-JsonValue -Object $healthResp.json -Path 'status') -eq 'ok') {
                $ready = $true
                break
            }
            Start-Sleep -Milliseconds 400
        }
        if (-not $ready) { throw "server did not become ready in $readyTimeoutSec s" }

        # На порту должен слушать именно запущенный процесс. Без этой проверки ответить может
        # посторонний экземпляр сервера, и функциональные результаты перестают относиться к сборке.
        $listeners = @(Get-NetTCPConnection -LocalPort $runtimePort -State Listen -ErrorAction SilentlyContinue)
        $portOwnership.checked = $true
        $portOwnership.listening_pids = @($listeners | ForEach-Object { [int]$_.OwningProcess } | Sort-Object -Unique)
        $portOwnership.owned_by_server = ($portOwnership.listening_pids -contains [int]$serverProc.Id)
        Save-CheckJson -Path (Join-Path $ctx.MetaDir 'port_ownership.json') -Value $portOwnership
        if (-not $portOwnership.owned_by_server) {
            throw "port $runtimePort is held by pid(s) $($portOwnership.listening_pids -join ',') instead of the started server pid $($serverProc.Id)"
        }

        Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'health.json') -Value $healthResp.json
        $healthImplemented = $true
        $healthConformant = $true

        $datasetsResp = Invoke-HttpRequestSafe -Method 'GET' -Uri "$baseUrl/api/datasets" -TimeoutSec 5
        Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'datasets.json') -Value $datasetsResp
        $datasetEntries = Get-DatasetEntries -Response $datasetsResp
        $datasetIDs = @($datasetEntries | ForEach-Object { $_.id })
        $controlDatasetId = Resolve-DatasetId -Entries $datasetEntries -PreferredIds @('control','events') -ExpectedSize @($controlFixture).Count
        $largeDatasetId = Resolve-DatasetId -Entries $datasetEntries -PreferredIds @('large100k','large','events100k') -ExpectedSize 100000
        $millionDatasetId = Resolve-DatasetId -Entries $datasetEntries -PreferredIds @('large1m','million','events1m') -ExpectedSize 1000000
        $controlEntry = @($datasetEntries | Where-Object { $_.id -eq $controlDatasetId }) | Select-Object -First 1
        $largeEntry = @($datasetEntries | Where-Object { $_.id -eq $largeDatasetId }) | Select-Object -First 1
        $datasetsProbe = [ordered]@{
            status_code = $datasetsResp.status_code
            ids = $datasetIDs
            resolved_control_id = $controlDatasetId
            resolved_large_id = $largeDatasetId
            resolved_million_id = $millionDatasetId
            control_size = if ($controlEntry) { [int]$controlEntry.size } else { -1 }
            large_size = if ($largeEntry) { [int]$largeEntry.size } else { -1 }
            control_size_matches_fixture = ($null -ne $controlEntry -and [int]$controlEntry.size -eq @($controlFixture).Count)
        }
        Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'minimum_datasets.json') -Value $datasetsProbe
        $datasetsImplemented = $true
        $datasetsConformant = ($datasetsResp.status_code -eq 200 -and -not [string]::IsNullOrWhiteSpace($controlDatasetId) -and $datasetsProbe.control_size_matches_fixture)
        if ([string]::IsNullOrWhiteSpace($controlDatasetId)) {
            throw "control dataset was not published by /api/datasets; ids seen: $($datasetIDs -join ',')"
        }

        $scoringProbe = [ordered]@{}
        $mainHints = [ordered]@{ user_id = 'ivan'; file_name = 'client base'; action = 'email_send'; destination_type = 'external' }
        # min_score здесь 0: абсолютная шкала score задана решением, поэтому основной запрос не должен
        # ничего отсекать. Семантика min_score проверяется отдельным запросом с порогом из наблюдённых
        # значений.
        $searchBody1 = New-SearchRequestBody -DatasetId $controlDatasetId -Around '2026-06-16T10:15:00Z' -Tolerance '30m' -Hints $mainHints -Before '30m' -After '10m' -RequireNearby @('create_archive','email_send') -MinScore 0 -Limit 20
        Set-Content -LiteralPath (Join-Path $ctx.InputsDir 'search_request.json') -Value $searchBody1 -Encoding UTF8
        $searchResp1 = Invoke-HttpRequestSafe -Method 'POST' -Uri "$baseUrl/api/search" -Body $searchBody1 -TimeoutSec 10
        Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'search_response.json') -Value $searchResp1
        $searchImplemented = $true
        $searchID = [string](Get-JsonValue -Object $searchResp1.json -Path 'search_id')
        $searchConformant = ($searchResp1.status_code -eq 200 -and -not [string]::IsNullOrWhiteSpace($searchID))
        $searchByIDImplemented = $searchConformant
        $contextImplemented = $searchConformant
        $explainImplemented = $searchConformant
        $first = $null
        $sorted = $true
        $minScoreOk = $true
        $matchedHintsOk = $false
        $searchByIDOk = $false
        $contextOk = $false
        $explainOk = $false
        if ($searchConformant) {
            $candidates = Get-JsonArray -Object $searchResp1.json -Path 'candidates'
            if ($candidates.Count -gt 0) {
                $first = $candidates[0]
                $matched = Get-JsonArray -Object $first -Path 'matched_hints'
                # Задание задаёт matched_hints как свободные фразы («user_id similar to ivan»),
                # поэтому проверяется упоминание имени подсказки, а не точное равенство строке.
                $matchedHintsOk = (
                    (Test-HintMentioned -MatchedHints $matched -Hint 'user_id') -and
                    (Test-HintMentioned -MatchedHints $matched -Hint 'file_name') -and
                    (Test-HintMentioned -MatchedHints $matched -Hint 'destination_type')
                )
                $scores = @($candidates | ForEach-Object { [double](Get-JsonValue -Object $_ -Path 'score') })
                for ($i = 1; $i -lt $scores.Count; $i++) {
                    if ($scores[$i-1] -lt $scores[$i]) { $sorted = $false }
                }

                # Порог берётся вторым по величине из фактических score, поэтому проверка не зависит
                # от выбранной шкалы: часть кандидатов обязана отсечься, а остаток — быть выше порога.
                $distinctScores = @($scores | Sort-Object -Descending -Unique)
                $minScoreProbe = [ordered]@{ observed_scores = $scores; applicable = ($distinctScores.Count -ge 2) }
                if ($distinctScores.Count -ge 2) {
                    # Округление вниз: min_score в запросе целое, а вторая по величине оценка может
                    # быть дробной. Ожидаемый состав считается по тому же целому порогу.
                    $threshold = [int][Math]::Floor([double]$distinctScores[1])
                    $minScoreBody = New-SearchRequestBody -DatasetId $controlDatasetId -Around '2026-06-16T10:15:00Z' -Tolerance '30m' -Hints $mainHints -Before '30m' -After '10m' -RequireNearby @('create_archive','email_send') -MinScore $threshold -Limit 20
                    $minScoreResp = Invoke-HttpRequestSafe -Method 'POST' -Uri "$baseUrl/api/search" -Body $minScoreBody -TimeoutSec 10
                    $filteredScores = @((Get-JsonArray -Object $minScoreResp.json -Path 'candidates') | ForEach-Object { [double](Get-JsonValue -Object $_ -Path 'score') })
                    $expectedCount = @($scores | Where-Object { $_ -ge $threshold }).Count
                    $minScoreProbe['threshold'] = $threshold
                    $minScoreProbe['status_code'] = $minScoreResp.status_code
                    $minScoreProbe['returned_scores'] = $filteredScores
                    $minScoreProbe['expected_count'] = $expectedCount
                    $minScoreProbe['all_at_or_above_threshold'] = (@($filteredScores | Where-Object { $_ -lt $threshold }).Count -eq 0)
                    $minScoreProbe['count_matches_expected'] = ($filteredScores.Count -eq $expectedCount)
                    $minScoreOk = ($minScoreResp.status_code -eq 200 -and $minScoreProbe['all_at_or_above_threshold'] -and $minScoreProbe['count_matches_expected'])
                } else {
                    # Все score равны: отсечь нечего, признак не проверяется и не выдаётся за пройденный.
                    $minScoreOk = $true
                    $minScoreProbe['threshold'] = -1
                }
                Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'minimum_min_score.json') -Value $minScoreProbe

                # Формат результата поиска зафиксирован страницей задания: total_events, warnings и
                # осмысленный summary входят в контракт, на который опирается frontend.
                $responseKeys = @($searchResp1.json.PSObject.Properties.Name)
                $candidateKeys = @($first.PSObject.Properties.Name)
                $formatProbe['response_keys'] = $responseKeys
                $formatProbe['has_warnings'] = ($responseKeys -contains 'warnings')
                $formatProbe['total_events'] = if ($responseKeys -contains 'total_events') { [int](Get-JsonValue -Object $searchResp1.json -Path 'total_events') } else { -1 }
                $formatProbe['total_events_matches_dataset'] = ([int]$formatProbe['total_events'] -eq @($controlFixture).Count)
                $formatProbe['total_candidates'] = if ($responseKeys -contains 'total_candidates') { [int](Get-JsonValue -Object $searchResp1.json -Path 'total_candidates') } else { -1 }
                $formatProbe['total_candidates_matches_array'] = ([int]$formatProbe['total_candidates'] -eq $candidates.Count)
                $formatProbe['candidate_keys'] = $candidateKeys
                $formatProbe['candidate_required_keys_present'] = (
                    ($candidateKeys -contains 'event_id') -and
                    ($candidateKeys -contains 'timestamp') -and
                    ($candidateKeys -contains 'score') -and
                    ($candidateKeys -contains 'summary') -and
                    ($candidateKeys -contains 'matched_hints') -and
                    ($candidateKeys -contains 'event')
                )
                $formatProbe['summary_non_empty'] = (-not [string]::IsNullOrWhiteSpace([string](Get-JsonValue -Object $first -Path 'summary')))
                $formatProbe['matched_hints_non_empty'] = (@($matched | Where-Object { -not [string]::IsNullOrWhiteSpace([string]$_) }).Count -gt 0)
                Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'minimum_result_format.json') -Value $formatProbe
                $probes['minimum.result_format'].implemented = $true
                $probes['minimum.result_format'].conformant = (
                    $formatProbe['has_warnings'] -and
                    $formatProbe['total_events_matches_dataset'] -and
                    $formatProbe['total_candidates_matches_array'] -and
                    $formatProbe['candidate_required_keys_present'] -and
                    $formatProbe['summary_non_empty'] -and
                    $formatProbe['matched_hints_non_empty']
                )
                $probes['minimum.result_format'].details = "warnings=$($formatProbe['has_warnings']); total_events=$($formatProbe['total_events']); summary=$($formatProbe['summary_non_empty'])"

                $searchByIDResp = Invoke-HttpRequestSafe -Method 'GET' -Uri "$baseUrl/api/search/$searchID" -TimeoutSec 10
                Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'search_by_id.json') -Value $searchByIDResp
                $byIDCandidates = Get-JsonArray -Object $searchByIDResp.json -Path 'candidates'
                $candidateIDs = @($candidates | ForEach-Object { [string](Get-JsonValue -Object $_ -Path 'event_id') })
                $byIDCandidateIDs = @($byIDCandidates | ForEach-Object { [string](Get-JsonValue -Object $_ -Path 'event_id') })
                $searchByIDOk = (
                    $searchByIDResp.status_code -eq 200 -and
                    [string](Get-JsonValue -Object $searchByIDResp.json -Path 'search_id') -eq $searchID -and
                    $byIDCandidates.Count -eq $candidates.Count -and
                    ($byIDCandidateIDs -join ',') -eq ($candidateIDs -join ',')
                )
                $searchByIDConformant = $searchByIDOk

                $eventID = [string](Get-JsonValue -Object $first -Path 'event_id')
                $expectedEvent = $controlExpected[$eventID]
                # Задание не определяет параметр набора у этого endpoint, поэтому контракт — вызов по
                # одному event_id. Ответ должен описывать то же событие, что и кандидат: совпадение
                # только по event_id проходит и когда набор выбран неверно, а идентификаторы совпали.
                $contextResp = Invoke-HttpRequestSafe -Method 'GET' -Uri "$baseUrl/api/events/$eventID/context?before=30m&after=10m" -TimeoutSec 10
                Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'event_context.json') -Value $contextResp
                $contextScopedResp = Invoke-HttpRequestSafe -Method 'GET' -Uri "$baseUrl/api/events/$eventID/context?dataset_id=$controlDatasetId&before=30m&after=10m" -TimeoutSec 10
                Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'event_context_scoped.json') -Value $contextScopedResp

                $expectedNeighbours = Get-ExpectedNeighbours -Fixture $controlFixture -AnchorEventId $eventID -Before ([timespan]::FromMinutes(30)) -After ([timespan]::FromMinutes(10))
                $fixtureIDs = @($controlFixture | ForEach-Object { [string]$_.event_id })

                $plainEval = Get-ContextEvaluation -Response $contextResp -FixtureIDs $fixtureIDs -ExpectedNeighbours $expectedNeighbours -ExpectedEvent $expectedEvent
                $scopedEval = Get-ContextEvaluation -Response $contextScopedResp -FixtureIDs $fixtureIDs -ExpectedNeighbours $expectedNeighbours -ExpectedEvent $expectedEvent
                # Требовать параметр dataset_id задание не даёт, но и не запрещает: реализация вправе
                # либо находить событие по одному идентификатору, либо честно просить набор. Незачёт
                # ставится за третий случай — ответ 200 с событиями постороннего набора.
                $plainDemandsDataset = ($contextResp.status_code -ge 400 -and $contextResp.status_code -lt 500 -and ([string](Get-JsonValue -Object $contextResp.json -Path 'error.field') -match 'dataset' -or [string](Get-JsonValue -Object $contextResp.json -Path 'error.message') -match 'dataset'))
                $contextProbe['requested_event_id'] = $eventID
                $contextProbe['expected_before_ids'] = $expectedNeighbours.before
                $contextProbe['expected_after_ids'] = $expectedNeighbours.after
                $contextProbe['plain_call'] = $plainEval
                $contextProbe['plain_demands_dataset_id'] = $plainDemandsDataset
                $contextProbe['scoped_call'] = $scopedEval
                $contextProbe['status_code'] = $contextResp.status_code
                $contextProbe['returned_event_id'] = $plainEval.returned_event_id
                $contextProbe['wrong_dataset_silently'] = ($contextResp.status_code -eq 200 -and $plainEval.foreign_ids.Count -gt 0)
                $contextOk = (
                    ($contextResp.status_code -eq 200 -and $plainEval.neighbours_ok) -or
                    ($plainDemandsDataset -and $contextScopedResp.status_code -eq 200 -and $scopedEval.neighbours_ok)
                )
                Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'good_context.json') -Value $contextProbe
                $contextConformant = $contextOk

                $explainResp = Invoke-HttpRequestSafe -Method 'GET' -Uri "$baseUrl/api/search/$searchID/candidates/$eventID/explain" -TimeoutSec 10
                Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'explain.json') -Value $explainResp
                # Вклады дробные (11,25; 22,5), поэтому сумма считается в double с допуском:
                # приведение каждого слагаемого к int даёт расхождение на корректных данных.
                $contributions = Get-JsonArray -Object $explainResp.json -Path 'contributions'
                $sum = Get-ContributionSum -Contributions $contributions
                $declared = [double](Get-JsonValue -Object $explainResp.json -Path 'score')
                $explainProbe['status_code'] = $explainResp.status_code
                $explainProbe['event_id_matches'] = ([string](Get-JsonValue -Object $explainResp.json -Path 'event_id') -eq $eventID)
                $explainProbe['contribution_count'] = $contributions.Count
                $explainProbe['contribution_sum'] = $sum
                $explainProbe['declared_score'] = $declared
                $explainProbe['sum_matches_score'] = ([Math]::Abs($sum - $declared) -le 0.01)
                $explainProbe['has_missed_hints'] = (Test-JsonHasProperty -Object $explainResp.json -Name 'missed_hints')
                # Задание задаёт вклад как hint/type/value/query/points. Имя признака принимается и как
                # hint, и как field — это выбор названия, а не отклонение от смысла; type требуется,
                # потому что вид совпадения (exact/substring/fuzzy/context) задан алгоритмическими
                # требованиями и нужен frontend для объяснения score.
                $explainProbe['contribution_fields_ok'] = ($contributions.Count -gt 0)
                foreach ($part in $contributions) {
                    $partKeys = @($part.PSObject.Properties.Name)
                    $hasName = (($partKeys -contains 'hint') -or ($partKeys -contains 'field'))
                    if (-not ($hasName -and ($partKeys -contains 'type') -and ($partKeys -contains 'points'))) {
                        $explainProbe['contribution_fields_ok'] = $false
                    }
                }
                Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'good_explain.json') -Value $explainProbe
                $explainOk = (
                    $explainResp.status_code -eq 200 -and
                    $explainProbe['event_id_matches'] -and
                    $explainProbe['contribution_count'] -gt 0 -and
                    $explainProbe['contribution_fields_ok'] -and
                    $explainProbe['sum_matches_score'] -and
                    $explainProbe['has_missed_hints']
                )
                $explainConformant = $explainOk
            }
        }
        $searchBody2 = New-SearchRequestBody -DatasetId $controlDatasetId -Hints ([ordered]@{ file_name = 'client base' }) -MinScore 0 -Limit 1
        $searchResp2 = Invoke-HttpRequestSafe -Method 'POST' -Uri "$baseUrl/api/search" -Body $searchBody2 -TimeoutSec 10
        $limitOk = ($searchResp2.status_code -eq 200 -and (Get-JsonArray -Object $searchResp2.json -Path 'candidates').Count -le 1)
        # Шкала и веса score — решение стажёра (настраиваемые веса вынесены в критерий «Отлично» 3),
        # поэтому проверяются свойства ранжирования, а не конкретное числовое значение.
        $scoringConformant = (
            $searchConformant -and
            $first -and
            ([string](Get-JsonValue -Object $first -Path 'event_id') -eq 'evt_exact') -and
            $matchedHintsOk -and
            $sorted -and
            $minScoreOk -and
            $limitOk -and
            $searchByIDOk
        )
        $scoringProbe['search1_status'] = $searchResp1.status_code
        $scoringProbe['search2_status'] = $searchResp2.status_code
        $allCandidates = Get-JsonArray -Object $searchResp1.json -Path 'candidates'
        $scoringProbe['candidate_count'] = $allCandidates.Count
        $scoringProbe['candidate_ids'] = @($allCandidates | ForEach-Object { [string](Get-JsonValue -Object $_ -Path 'event_id') })
        $scoringProbe['first_event_id'] = if ($first) { [string](Get-JsonValue -Object $first -Path 'event_id') } else { '' }
        $scoringProbe['first_event_id_expected'] = 'evt_exact'
        $scoringProbe['first_score'] = if ($first) { [double](Get-JsonValue -Object $first -Path 'score') } else { -1 }
        $scoringProbe['matched_hints'] = if ($first) { Get-JsonArray -Object $first -Path 'matched_hints' } else { @() }
        $scoringProbe['matched_hints_ok'] = $matchedHintsOk
        $scoringProbe['sorted_score_desc'] = $sorted
        $scoringProbe['min_score_inclusive_ok'] = $minScoreOk
        $scoringProbe['limit_ok'] = $limitOk
        $scoringProbe['search_by_id_ok'] = $searchByIDOk
        Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'minimum_scoring.json') -Value $scoringProbe
        $probes['minimum.scoring'].implemented = $true
        $probes['minimum.scoring'].conformant = $scoringConformant
        $probes['minimum.scoring'].details = "first=$($scoringProbe.first_event_id); expected=evt_exact; score=$($scoringProbe.first_score)"

        $timeProbe = [ordered]@{}
        $timeNarrowBody = New-SearchRequestBody -DatasetId $controlDatasetId -Around '2026-06-16T10:15:00Z' -Tolerance '30m' -MinScore 0 -Limit 100
        $timeNarrowResp = Invoke-HttpRequestSafe -Method 'POST' -Uri "$baseUrl/api/search" -Body $timeNarrowBody -TimeoutSec 10
        $timeOutsideBody = New-SearchRequestBody -DatasetId $controlDatasetId -Around '2026-06-20T10:15:00Z' -Tolerance '1m' -MinScore 0 -Limit 20
        $timeOutsideResp = Invoke-HttpRequestSafe -Method 'POST' -Uri "$baseUrl/api/search" -Body $timeOutsideBody -TimeoutSec 10
        $narrowIDs = @((Get-JsonArray -Object $timeNarrowResp.json -Path 'candidates') | ForEach-Object { [string](Get-JsonValue -Object $_ -Path 'event_id') })
        $outsideIDs = @((Get-JsonArray -Object $timeOutsideResp.json -Path 'candidates') | ForEach-Object { [string](Get-JsonValue -Object $_ -Path 'event_id') })
        # evt_boundary стоит ровно на +30m от around, то есть проверяется включающая граница окна.
        $boundaryIncluded = ($narrowIDs -contains 'evt_boundary')
        $outsideExcluded = -not ($narrowIDs -contains 'evt_outside')
        $outsideFound = ($outsideIDs -contains 'evt_outside')
        $timeProbe['narrow_ids'] = $narrowIDs
        $timeProbe['outside_ids'] = $outsideIDs
        $timeProbe['boundary_included'] = $boundaryIncluded
        $timeProbe['outside_excluded_in_narrow'] = $outsideExcluded
        $timeProbe['outside_found_around_outside'] = $outsideFound
        Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'good_time_filter.json') -Value $timeProbe
        $probes['good.time_filter'].implemented = $true
        $probes['good.time_filter'].conformant = ($timeNarrowResp.status_code -eq 200 -and $timeOutsideResp.status_code -eq 200 -and $boundaryIncluded -and $outsideExcluded -and $outsideFound)
        $probes['good.time_filter'].details = "narrow_count=$($narrowIDs.Count); outside_count=$($outsideIDs.Count)"

        $nearbyProbe = [ordered]@{}
        $nearbyGoodBody = New-SearchRequestBody -DatasetId $controlDatasetId -Around '2026-06-16T10:15:00Z' -Tolerance '1h' -Before '30m' -After '10m' -RequireNearby @('create_archive','email_send') -MinScore 0 -Limit 20
        $nearbyShortBody = New-SearchRequestBody -DatasetId $controlDatasetId -Around '2026-06-16T10:15:00Z' -Tolerance '1h' -Before '5m' -After '1m' -RequireNearby @('create_archive','email_send') -MinScore 0 -Limit 20
        $nearbyMissingBody = New-SearchRequestBody -DatasetId $controlDatasetId -Around '2026-06-16T10:15:00Z' -Tolerance '1h' -Before '30m' -After '10m' -RequireNearby @('action_missing') -MinScore 0 -Limit 20
        $nearbyGoodResp = Invoke-HttpRequestSafe -Method 'POST' -Uri "$baseUrl/api/search" -Body $nearbyGoodBody -TimeoutSec 10
        $nearbyShortResp = Invoke-HttpRequestSafe -Method 'POST' -Uri "$baseUrl/api/search" -Body $nearbyShortBody -TimeoutSec 10
        $nearbyMissingResp = Invoke-HttpRequestSafe -Method 'POST' -Uri "$baseUrl/api/search" -Body $nearbyMissingBody -TimeoutSec 10
        $goodIDs = @((Get-JsonArray -Object $nearbyGoodResp.json -Path 'candidates') | ForEach-Object { [string](Get-JsonValue -Object $_ -Path 'event_id') })
        $shortIDs = @((Get-JsonArray -Object $nearbyShortResp.json -Path 'candidates') | ForEach-Object { [string](Get-JsonValue -Object $_ -Path 'event_id') })
        $missingIDs = @((Get-JsonArray -Object $nearbyMissingResp.json -Path 'candidates') | ForEach-Object { [string](Get-JsonValue -Object $_ -Path 'event_id') })
        $nearbyProbe['good_ids'] = $goodIDs
        $nearbyProbe['short_ids'] = $shortIDs
        $nearbyProbe['exact_included'] = ($goodIDs -contains 'evt_exact')
        $nearbyProbe['exact_excluded_short_window'] = -not ($shortIDs -contains 'evt_exact')
        $nearbyProbe['missing_action_excludes_all'] = ($missingIDs.Count -eq 0)
        Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'good_nearby.json') -Value $nearbyProbe
        $probes['good.nearby'].implemented = $true
        $probes['good.nearby'].conformant = ($nearbyGoodResp.status_code -eq 200 -and $nearbyShortResp.status_code -eq 200 -and $nearbyMissingResp.status_code -eq 200 -and $nearbyProbe.exact_included -and $nearbyProbe.exact_excluded_short_window -and $nearbyProbe.missing_action_excludes_all)
        $probes['good.nearby'].details = "good=$($goodIDs.Count); short=$($shortIDs.Count); missing=$($missingIDs.Count)"

        # Минимальный набор кодов по заданию: dataset_not_found, invalid_query, invalid_time,
        # invalid_duration, search_not_found, internal_error. HTTP-статусы заданием не зафиксированы,
        # поэтому требуется 4xx. Коды вне этого набора (method_not_allowed, window_too_wide) заданием
        # не предусмотрены и на вердикт не влияют — они собираются как справочные.
        # Обрезанное тело — намеренно невалидный JSON, поэтому собирается вручную.
        $errMalformedBody = (New-SearchRequestBody -DatasetId $controlDatasetId -MinScore 0 -Limit 1).TrimEnd('}')
        $errTimeBody = New-SearchRequestBody -DatasetId $controlDatasetId -Around 'bad-time' -Tolerance '30m' -MinScore 0 -Limit 1
        $errToleranceBody = New-SearchRequestBody -DatasetId $controlDatasetId -Around '2026-06-16T10:15:00Z' -Tolerance 'bad' -MinScore 0 -Limit 1
        $errDatasetBody = New-SearchRequestBody -DatasetId 'dataset_that_does_not_exist' -MinScore 0 -Limit 1
        $errMalformed = Invoke-HttpRequestSafe -Method 'POST' -Uri "$baseUrl/api/search" -Body $errMalformedBody -TimeoutSec 10
        $errDataset = Invoke-HttpRequestSafe -Method 'POST' -Uri "$baseUrl/api/search" -Body $errDatasetBody -TimeoutSec 10
        $errTime = Invoke-HttpRequestSafe -Method 'POST' -Uri "$baseUrl/api/search" -Body $errTimeBody -TimeoutSec 10
        $errTolerance = Invoke-HttpRequestSafe -Method 'POST' -Uri "$baseUrl/api/search" -Body $errToleranceBody -TimeoutSec 10
        $errSearchMissing = Invoke-HttpRequestSafe -Method 'GET' -Uri "$baseUrl/api/search/srch_does_not_exist" -TimeoutSec 10

        $structuredProbe = [ordered]@{}
        $structuredProbe['invalid_query'] = Test-ErrorEnvelope -Response $errMalformed -AcceptedCodes @('invalid_query','invalid_json')
        $structuredProbe['dataset_not_found'] = Test-ErrorEnvelope -Response $errDataset -AcceptedCodes @('dataset_not_found')
        $structuredProbe['invalid_time'] = Test-ErrorEnvelope -Response $errTime -AcceptedCodes @('invalid_time')
        $structuredProbe['invalid_duration'] = Test-ErrorEnvelope -Response $errTolerance -AcceptedCodes @('invalid_duration')
        $structuredProbe['search_not_found'] = Test-ErrorEnvelope -Response $errSearchMissing -AcceptedCodes @('search_not_found')
        $structuredCodesOk = $true
        foreach ($key in @('invalid_query','dataset_not_found','invalid_time','invalid_duration','search_not_found')) {
            if (-not $structuredProbe[$key].ok) { $structuredCodesOk = $false }
        }
        $structuredProbe['observed_codes'] = @($structuredProbe.Keys | Where-Object { $_ -ne 'observed_codes' } | ForEach-Object { "$($_)=$($structuredProbe[$_].code)" })
        Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'good_structured_errors.json') -Value $structuredProbe
        $probes['good.structured_errors'].implemented = $true
        $probes['good.structured_errors'].conformant = $structuredCodesOk
        $probes['good.structured_errors'].details = ($structuredProbe['observed_codes'] -join '; ')

        # Формат ошибки в задании содержит field, и frontend по контракту показывает сообщение рядом
        # с полем. Принимается и полный путь ("time.tolerance"), и короткое имя.
        $fieldsProbe = [ordered]@{}
        $fieldsProbe['invalid_time'] = Test-ErrorEnvelope -Response $errTime -AcceptedCodes @('invalid_time') -ExpectedField 'time.around'
        $fieldsProbe['invalid_duration'] = Test-ErrorEnvelope -Response $errTolerance -AcceptedCodes @('invalid_duration') -ExpectedField 'time.tolerance'
        Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'good_error_fields.json') -Value $fieldsProbe
        $probes['good.error_fields'].implemented = $true
        $probes['good.error_fields'].conformant = ($fieldsProbe['invalid_time'].ok -and $fieldsProbe['invalid_duration'].ok)
        $probes['good.error_fields'].details = "around_field=$($fieldsProbe['invalid_time'].field); tolerance_field=$($fieldsProbe['invalid_duration'].field)"

        # Идентификаторы наборов сервер выводит из имён файлов, поэтому набор на 100 000 событий
        # называется large100k. Требование задания — ответ до 2 секунд; с согласованным допуском
        # ±10% (комментарий Степаненко Руслана от 30.07.2026) порог 2200 мс.
        $largeProbe = [ordered]@{}
        $largeCount = if ($largeEntry) { [int]$largeEntry.size } else { -1 }
        $largeBody = New-SearchRequestBody -DatasetId $largeDatasetId -Around '2026-06-16T10:15:00Z' -Tolerance '30d' -Hints ([ordered]@{ user_id = 'target_user'; file_name = 'large_target_054321'; action = 'email_send'; destination_type = 'external' }) -MinScore 0 -Limit 5
        $sw = [System.Diagnostics.Stopwatch]::StartNew()
        $largeResp = Invoke-HttpRequestSafe -Method 'POST' -Uri "$baseUrl/api/search" -Body $largeBody -TimeoutSec 30
        $sw.Stop()
        $largeSearchDuration = $sw.ElapsedMilliseconds
        $largeSearchID = [string](Get-JsonValue -Object $largeResp.json -Path 'search_id')
        $largeByIDOk = $false
        if ($largeResp.status_code -eq 200 -and -not [string]::IsNullOrWhiteSpace($largeSearchID)) {
            $largeByID = Invoke-HttpRequestSafe -Method 'GET' -Uri "$baseUrl/api/search/$largeSearchID" -TimeoutSec 30
            $largeByIDOk = ($largeByID.status_code -eq 200 -and [string](Get-JsonValue -Object $largeByID.json -Path 'search_id') -eq $largeSearchID)
        }
        $largeCandidates = Get-JsonArray -Object $largeResp.json -Path 'candidates'
        $largeTopID = if ($largeCandidates.Count -gt 0) { [string](Get-JsonValue -Object $largeCandidates[0] -Path 'event_id') } else { '' }
        $serverProc.Refresh()
        $largeProbe['event_count'] = $largeCount
        $largeProbe['event_count_expected'] = 100000
        $largeProbe['search_duration_ms'] = [int]$largeSearchDuration
        $largeProbe['search_budget_ms'] = 2200
        $largeProbe['search_within_budget'] = ($largeSearchDuration -le 2200)
        $largeProbe['top_event_id'] = $largeTopID
        $largeProbe['search_by_id_ok'] = $largeByIDOk
        $largeProbe['working_set_mb'] = [int]([Math]::Round($serverProc.WorkingSet64 / 1MB))
        $largeProbe['peak_working_set_mb'] = [int]([Math]::Round($serverProc.PeakWorkingSet64 / 1MB))
        $largeProbe['large_generation'] = $largeMeta
        Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'excellent_large_dataset.json') -Value $largeProbe
        $probes['excellent.large_dataset'].implemented = $true
        $probes['excellent.large_dataset'].conformant = ($largeCount -eq 100000 -and $largeResp.status_code -eq 200 -and $largeTopID -eq 'large_target_054321' -and $largeProbe.search_within_budget -and $largeByIDOk)
        $probes['excellent.large_dataset'].details = "count=$largeCount; duration_ms=$largeSearchDuration; budget_ms=2200; top=$largeTopID; rss_mb=$($largeProbe['working_set_mb'])"

        # Требование 2 раздела «Требования к памяти и производительности»: для набора до 1 000 000
        # событий память процесса не более 512 МБ. Проверяется только при -IncludeMillionDataset,
        # иначе остаётся not_tested, а не выдаётся за проверенное.
        if ($IncludeMillionDataset) {
            $millionProbe = [ordered]@{}
            $millionEntry = @($datasetEntries | Where-Object { $_.id -eq $millionDatasetId }) | Select-Object -First 1
            $millionBody = New-SearchRequestBody -DatasetId $millionDatasetId -Around '2026-06-16T10:15:00Z' -Tolerance '365d' -Hints ([ordered]@{ user_id = 'target_user'; file_name = 'million_target_654321'; action = 'email_send'; destination_type = 'external' }) -MinScore 0 -Limit 5
            $swMillion = [System.Diagnostics.Stopwatch]::StartNew()
            $millionResp = Invoke-HttpRequestSafe -Method 'POST' -Uri "$baseUrl/api/search" -Body $millionBody -TimeoutSec 120
            $swMillion.Stop()
            $serverProc.Refresh()
            $millionCandidates = Get-JsonArray -Object $millionResp.json -Path 'candidates'
            $millionTopID = if ($millionCandidates.Count -gt 0) { [string](Get-JsonValue -Object $millionCandidates[0] -Path 'event_id') } else { '' }
            $millionProbe['event_count'] = if ($millionEntry) { [int]$millionEntry.size } else { -1 }
            $millionProbe['search_duration_ms'] = [int]$swMillion.ElapsedMilliseconds
            $millionProbe['top_event_id'] = $millionTopID
            $millionProbe['working_set_mb'] = [int]([Math]::Round($serverProc.WorkingSet64 / 1MB))
            $millionProbe['peak_working_set_mb'] = [int]([Math]::Round($serverProc.PeakWorkingSet64 / 1MB))
            $millionProbe['memory_budget_mb'] = 563
            $millionProbe['within_memory_budget'] = ($millionProbe['peak_working_set_mb'] -le 563)
            $millionProbe['large1m_generation'] = $datasetMeta.large1m
            Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'excellent_memory_million.json') -Value $millionProbe
            $probes['excellent.memory_million'].implemented = $true
            $probes['excellent.memory_million'].conformant = ($millionResp.status_code -eq 200 -and $millionProbe['event_count'] -eq 1000000 -and $millionProbe['within_memory_budget'] -and $millionTopID -eq 'million_target_654321')
            $probes['excellent.memory_million'].details = "count=$($millionProbe['event_count']); peak_mb=$($millionProbe['peak_working_set_mb']); budget_mb=563"
        }

        # Раздел «CLI для разработки» задания требует режим локальной проверки:
        # event-memory-search-api search --events <jsonl> --query <json> --out result.json
        $cliQueryPath = Join-Path $ctx.InputsDir 'cli_query.json'
        Set-Content -LiteralPath $cliQueryPath -Value $searchBody1 -Encoding UTF8
        $cliOut = Join-Path $ctx.OutputsDir 'cli_search_result.json'
        $cliEvents = Join-Path $tempDatasetsDir 'events.jsonl'
        Invoke-CheckCommand -Ctx $ctx -Name 'cli_search' -Command "& '$serverExe' search --events '$cliEvents' --query '$cliQueryPath' --out '$cliOut'" | Out-Null
        $cliProbe = [ordered]@{
            command_exit = [int]$ctx.CommandResults['cli_search'].exit_code
            output_exists = (Test-Path -LiteralPath $cliOut)
            parsed = $false
            candidate_count = -1
            top_event_id = ''
            top_matches_http = $false
        }
        if ($cliProbe.output_exists) {
            try {
                $cliJson = Get-Content -LiteralPath $cliOut -Raw | ConvertFrom-Json
                $cliProbe.parsed = $true
                $cliCandidates = Get-JsonArray -Object $cliJson -Path 'candidates'
                $cliProbe.candidate_count = $cliCandidates.Count
                if ($cliCandidates.Count -gt 0) {
                    $cliProbe.top_event_id = [string](Get-JsonValue -Object $cliCandidates[0] -Path 'event_id')
                    $cliProbe.top_matches_http = ($cliProbe.top_event_id -eq [string]$scoringProbe['first_event_id'])
                }
            } catch {}
        }
        Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'cli_search.json') -Value $cliProbe
        $probes['engineering.cli_search'].implemented = ($cliProbe.command_exit -eq 0)
        $probes['engineering.cli_search'].conformant = ($cliProbe.command_exit -eq 0 -and $cliProbe.parsed -and $cliProbe.candidate_count -gt 0 -and $cliProbe.top_matches_http)
        $probes['engineering.cli_search'].details = "exit=$($cliProbe.command_exit); candidates=$($cliProbe.candidate_count); top=$($cliProbe.top_event_id)"
    }
    catch {
        # Без этого обработчика любой throw рантайм-фазы (сервер не поднялся, порт занят, набор не
        # опубликован) убивал скрипт до сборки архива, и разбирать прогон было нечем. Теперь причина
        # попадает в архив, а вердикты остаются not_tested.
        $runtimeError = [ordered]@{
            message = [string]$_.Exception.Message
            type = [string]$_.Exception.GetType().FullName
            at = (Get-Date).ToString('o')
        }
        Save-CheckJson -Path (Join-Path $ctx.MetaDir 'runtime_error.json') -Value $runtimeError
        Write-Host "RUNTIME_PHASE_FAILED=$($runtimeError.message)"
    }
    finally {
        # Stop-Process -Force обрывает перекачку перенаправленных потоков, и хвост журнала сервера
        # теряется. Пауза перед завершением даёт вывод дописаться, а размеры журналов фиксируются,
        # чтобы пустой stdout нельзя было принять за отсутствие вывода в коде.
        if ($serverProc -and -not $serverProc.HasExited) {
            Start-Sleep -Milliseconds 750
            Stop-Process -Id $serverProc.Id -Force
            try { Wait-Process -Id $serverProc.Id -Timeout 10 -ErrorAction SilentlyContinue } catch {}
            Start-Sleep -Milliseconds 250
        }
        $stopped = $true
        if ($serverProc) { $stopped = $serverProc.HasExited }
        Save-CheckJson -Path (Join-Path $ctx.MetaDir 'server_stop.json') -Value ([ordered]@{
            pid = if ($serverProc) { $serverProc.Id } else { $null }
            stopped = $stopped
            stopped_at = (Get-Date).ToString('o')
            exit_code = if ($serverProc -and $serverProc.HasExited) { [int]$serverProc.ExitCode } else { $null }
            stdout_bytes = if (Test-Path -LiteralPath $serverStdout) { [int64](Get-Item -LiteralPath $serverStdout).Length } else { -1 }
            stderr_bytes = if (Test-Path -LiteralPath $serverStderr) { [int64](Get-Item -LiteralPath $serverStderr).Length } else { -1 }
        })
    }
}

$datasetsRemoved = $false
if (Test-Path -LiteralPath $tempDatasetsDir) {
    Remove-Item -LiteralPath $tempDatasetsDir -Recurse -Force -ErrorAction SilentlyContinue
    $datasetsRemoved = -not (Test-Path -LiteralPath $tempDatasetsDir)
}
Save-CheckJson -Path (Join-Path $ctx.OutputsDir 'cleanup.json') -Value ([ordered]@{
    runtime_port = $runtimePort
    temp_datasets_dir = $tempDatasetsDir
    temp_datasets_removed = $datasetsRemoved
    million_dataset_generated = [bool]$IncludeMillionDataset
})

Add-BooleanFeatureAssessment -Ctx $ctx -Id 'minimum.health' -Level 'minimum' -Category 'api' -Requirement 'Backend starts and answers /api/health' -Implemented $healthImplemented -Conformant $healthConformant -Evidence @('outputs/health.json')
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'minimum.datasets' -Level 'minimum' -Category 'api' -Requirement 'GET /api/datasets returns datasets' -Implemented $datasetsImplemented -Conformant $datasetsConformant -Evidence @('outputs/datasets.json')
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'minimum.search' -Level 'minimum' -Category 'api' -Requirement 'POST /api/search returns candidates' -Implemented $searchImplemented -Conformant $searchConformant -Evidence @('outputs/search_response.json','outputs/minimum_scoring.json')
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'minimum.scoring' -Level 'minimum' -Category 'algorithm' -Requirement 'Runtime scoring ranks evt_exact first and honours order, matched_hints, limit and min_score' -Implemented $probes['minimum.scoring'].implemented -Conformant $probes['minimum.scoring'].conformant -Evidence @($probes['minimum.scoring'].evidence) -Details $probes['minimum.scoring'].details
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'minimum.result_format' -Level 'minimum' -Category 'api' -Requirement 'Search result matches the assignment format: total_events, total_candidates, warnings, candidate summary and matched_hints' -Implemented $probes['minimum.result_format'].implemented -Conformant $probes['minimum.result_format'].conformant -Evidence @($probes['minimum.result_format'].evidence) -Details $probes['minimum.result_format'].details

Add-BooleanFeatureAssessment -Ctx $ctx -Id 'good.search_by_id' -Level 'good' -Category 'api' -Requirement 'GET /api/search/{search_id} returns stored result' -Implemented $searchByIDImplemented -Conformant $searchByIDConformant -Evidence @('outputs/search_by_id.json')
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'good.context' -Level 'good' -Category 'api' -Requirement 'GET /api/events/{event_id}/context returns the context of the requested candidate, not of a same-id event from another dataset' -Implemented $contextImplemented -Conformant $contextConformant -Evidence @('outputs/event_context.json','outputs/good_context.json') -Details $(if ($contextProbe.Count -gt 0) { "requested=$($contextProbe['requested_event_id']); returned=$($contextProbe['returned_event_id'])" } else { '' })
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'good.explain' -Level 'good' -Category 'api' -Requirement 'Explain returns contributions whose points sum to the declared score within 0.01' -Implemented $explainImplemented -Conformant $explainConformant -Evidence @('outputs/explain.json','outputs/good_explain.json') -Details $(if ($explainProbe.Count -gt 0) { "sum=$($explainProbe['contribution_sum']); score=$($explainProbe['declared_score'])" } else { '' })
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'good.time_filter' -Level 'good' -Category 'algorithm' -Requirement 'Runtime time filter validates boundary, around and outside event' -Implemented $probes['good.time_filter'].implemented -Conformant $probes['good.time_filter'].conformant -Evidence @($probes['good.time_filter'].evidence) -Details $probes['good.time_filter'].details
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'good.nearby' -Level 'good' -Category 'algorithm' -Requirement 'Runtime nearby checks all required rules and windows' -Implemented $probes['good.nearby'].implemented -Conformant $probes['good.nearby'].conformant -Evidence @($probes['good.nearby'].evidence) -Details $probes['good.nearby'].details
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'good.structured_errors' -Level 'good' -Category 'api' -Requirement 'Errors use the single envelope and the six codes required by the assignment (invalid_query, dataset_not_found, invalid_time, invalid_duration, search_not_found)' -Implemented $probes['good.structured_errors'].implemented -Conformant $probes['good.structured_errors'].conformant -Evidence @($probes['good.structured_errors'].evidence) -Details $probes['good.structured_errors'].details
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'good.error_fields' -Level 'good' -Category 'api' -Requirement 'Validation errors carry the field that failed, as in the assignment error example' -Implemented $probes['good.error_fields'].implemented -Conformant $probes['good.error_fields'].conformant -Evidence @($probes['good.error_fields'].evidence) -Details $probes['good.error_fields'].details
$apiDocsPath = Join-Path $ctx.RepoRoot 'docs\api.md'
$apiDocsOk = (Test-Path -LiteralPath $apiDocsPath) -and ((Get-Item -LiteralPath $apiDocsPath).Length -gt 64)
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'good.api_docs' -Level 'good' -Category 'documentation' -Requirement 'API documentation artifact exists' -Implemented $apiDocsOk -Conformant $apiDocsOk -Evidence @('docs/api.md')

Add-BooleanFeatureAssessment -Ctx $ctx -Id 'excellent.large_dataset' -Level 'excellent' -Category 'performance' -Requirement 'Search over 100000 events answers within 2 s (+10% tolerance) and the stored result matches' -Implemented $probes['excellent.large_dataset'].implemented -Conformant $probes['excellent.large_dataset'].conformant -Evidence @($probes['excellent.large_dataset'].evidence, 'outputs/large_dataset_generation.json') -Details $probes['excellent.large_dataset'].details
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'excellent.memory_million' -Level 'excellent' -Category 'performance' -Requirement 'With a 1000000 event dataset the process stays within 512 MB (+10% tolerance)' -Implemented $probes['excellent.memory_million'].implemented -Conformant $probes['excellent.memory_million'].conformant -Evidence @($probes['excellent.memory_million'].evidence) -Details $(if ($IncludeMillionDataset) { $probes['excellent.memory_million'].details } else { 'not measured: run with -IncludeMillionDataset' })

Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.cli_search' -Level 'engineering' -Category 'cli' -Requirement 'CLI search mode from the assignment runs and agrees with the HTTP result' -Implemented $probes['engineering.cli_search'].implemented -Conformant $probes['engineering.cli_search'].conformant -Evidence @($probes['engineering.cli_search'].evidence, 'logs/cli_search.log') -Details $probes['engineering.cli_search'].details
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.clean_worktree' -Level 'engineering' -Category 'reproducibility' -Requirement 'Run is attributable: worktree clean, results tied to a commit' -Implemented $true -Conformant $attribution.attributable -Evidence @('meta/run_attribution.json') -Details "head=$($attribution.head); dirty_entries=$($attribution.dirty_entries.Count); allow_dirty_tree=$($attribution.allow_dirty_tree)"
Add-BooleanFeatureAssessment -Ctx $ctx -Id 'engineering.runtime_phase' -Level 'engineering' -Category 'reproducibility' -Requirement 'Runtime phase completed without aborting: server started, port owned, datasets published' -Implemented $true -Conformant ($null -eq $runtimeError) -Evidence @('meta/runtime_error.json','logs/server_stderr.log','meta/port_ownership.json') -Details $(if ($runtimeError) { [string]$runtimeError.message } else { 'completed' })

Complete-Check -Ctx $ctx -Extra @{
    runtime_port = $runtimePort
    run_attribution = $attribution
    port_ownership = $portOwnership
    datasets = $datasetMeta
    million_dataset_included = [bool]$IncludeMillionDataset
    expected_api = @('/api/health','/api/datasets','POST /api/search','GET /api/search/{search_id}','GET /api/events/{event_id}/context','GET /api/search/{search_id}/candidates/{event_id}/explain')
    required_error_codes = @('dataset_not_found','invalid_query','invalid_time','invalid_duration','search_not_found','internal_error')
    probe_files = @('outputs/minimum_datasets.json','outputs/minimum_scoring.json','outputs/minimum_min_score.json','outputs/minimum_result_format.json','outputs/good_context.json','outputs/good_explain.json','outputs/good_time_filter.json','outputs/good_nearby.json','outputs/good_structured_errors.json','outputs/good_error_fields.json','outputs/excellent_large_dataset.json','outputs/excellent_memory_million.json','outputs/cli_search.json')
}


