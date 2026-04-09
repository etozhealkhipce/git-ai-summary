# Install git-ai-summary from GitHub Releases (Windows).
# Usage: irm https://raw.githubusercontent.com/etozhealkhipce/git-ai-summary/main/install.ps1 | iex
# Pin version: $env:GIT_AI_SUMMARY_VERSION = "0.1.0"; irm ... | iex

$ErrorActionPreference = "Stop"

$Repo = "etozhealkhipce/git-ai-summary"
$BaseUrl = "https://github.com/$Repo"
$InstallDir = if ($env:GIT_AI_SUMMARY_INSTALL_DIR) { $env:GIT_AI_SUMMARY_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "git-ai-summary" }

function Get-Target {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "x86_64-pc-windows-msvc" }
        "ARM64" { return "aarch64-pc-windows-msvc" }
        default {
            Write-Error "Unsupported architecture: $arch. x86_64 (AMD64) and ARM64 are supported."
        }
    }
}

function Get-Version {
    if ($env:GIT_AI_SUMMARY_VERSION) {
        return $env:GIT_AI_SUMMARY_VERSION
    }
    try {
        $null = Invoke-WebRequest -Uri "$BaseUrl/releases/latest" -MaximumRedirection 0 -ErrorAction SilentlyContinue
    }
    catch {
        if ($_.Exception.Response.StatusCode -eq 302) {
            $location = $_.Exception.Response.Headers["Location"]
            if ($location -match "/tag/v(.+)$") {
                return $Matches[1].TrimEnd('/')
            }
        }
    }
    $apiUrl = "https://api.github.com/repos/$Repo/releases/latest"
    $release = Invoke-RestMethod -Uri $apiUrl
    $tag = $release.tag_name
    if ($tag -match "^v(.+)$") {
        return $Matches[1]
    }
    return $tag
}

$Target = Get-Target
$Version = Get-Version
$ZipName = "git-ai-summary-$Version-$Target.zip"
$Url = "$BaseUrl/releases/download/v$Version/$ZipName"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$TempZip = Join-Path ([System.IO.Path]::GetTempPath()) $ZipName

Write-Host "Downloading git-ai-summary $Version for $Target..."
Invoke-WebRequest -Uri $Url -OutFile $TempZip -UseBasicParsing

Write-Host "Extracting to $InstallDir..."
Expand-Archive -Path $TempZip -DestinationPath $InstallDir -Force
Remove-Item -Path $TempZip -Force -ErrorAction SilentlyContinue

$ExePath = Join-Path $InstallDir "git-ai-summary.exe"
if (-not (Test-Path $ExePath)) {
    Write-Error "Extraction did not produce git-ai-summary.exe in $InstallDir"
}

Write-Host "Installed git-ai-summary $Version to $ExePath"

$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host ""
    Write-Host "Add git-ai-summary to your PATH:"
    Write-Host "  [Environment]::SetEnvironmentVariable('Path', \"`$env:Path;$InstallDir\", 'User')"
    Write-Host "Then restart your terminal, or run: & '$ExePath' --help"
}

Write-Host ""
Write-Host "Configure API: & '$ExePath' setup"
