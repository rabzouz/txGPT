param(
    [Parameter(Position = 0, ValueFromRemainingArguments = $true)]
    [string[]]$Prompt,

    [string]$Tool = "",
    [string]$Lang = "fr",
    [string]$TxgptBin = "",
    [string]$Python = "python"
)

$ErrorActionPreference = "Stop"

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

$txgptOutput = & $TxgptBin @txArgs
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

$renderer = Join-Path $scriptDir "rich_display.py"
& $Python $renderer --prompt $promptText --output $txgptOutput --title "txGPT Rich"
