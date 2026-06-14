param(
    [Parameter(Position = 0, ValueFromRemainingArguments = $true)]
    [string[]]$Prompt,

    [string]$Tool = "",
    [string]$Lang = "fr",
    [string]$TxgptBin = "",
    [string]$Python = "python"
)

$ErrorActionPreference = "Stop"
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::InputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
$env:PYTHONIOENCODING = "utf-8"

if (-not $Prompt -or $Prompt.Count -eq 0) {
    Write-Host 'Usage: .\txgpt_rich.ps1 [-Tool powershell] [-Lang fr] "Votre prompt"'
    exit 1
}

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
if (-not $TxgptBin) {
    $exePath = Join-Path $scriptDir "txgpt.exe"
    $unixPath = Join-Path $scriptDir "txgpt"
    if (Test-Path -LiteralPath $exePath) {
        $TxgptBin = $exePath
    } elseif (Test-Path -LiteralPath $unixPath) {
        $TxgptBin = $unixPath
    } else {
        $TxgptBin = "txgpt.exe"
    }
}

$promptText = ($Prompt -join " ")
$txArgs = @("--json", "--lang", $Lang)
if ($Tool) {
    $txArgs += @("--tool", $Tool)
}
$txArgs += $promptText

$renderer = Join-Path $scriptDir "rich_display.py"
$tempOutput = Join-Path ([System.IO.Path]::GetTempPath()) ("txgpt-rich-{0}.json" -f ([System.Guid]::NewGuid().ToString("N")))

try {
    & $TxgptBin @txArgs | Out-File -LiteralPath $tempOutput -Encoding utf8
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }

    & $Python $renderer --prompt $promptText --output-file $tempOutput --title "txGPT Rich"
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
} finally {
    if (Test-Path -LiteralPath $tempOutput) {
        Remove-Item -LiteralPath $tempOutput -Force
    }
}
