Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
$ProgressPreference = 'SilentlyContinue'  # Speeds up Invoke-WebRequest significantly

# Allow overriding the version
$Version = if ($env:CODECRAFTERS_CLI_VERSION) { $env:CODECRAFTERS_CLI_VERSION } else { "v46" }

# Detect architecture
$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default {
        Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
        exit 1
    }
}

$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\codecrafters" }
$InstallPath = if ($env:INSTALL_PATH) { $env:INSTALL_PATH } else { "$InstallDir\codecrafters.exe" }

$DownloadUrl = "https://github.com/codecrafters-io/cli/releases/download/$Version/${Version}_windows_$Arch.tar.gz"

Write-Host "This script will automatically install codecrafters ($Version) for you."
Write-Host "Installation path: $InstallPath"

$TempDir = Join-Path $env:TEMP "codecrafters-install-$([System.Guid]::NewGuid().ToString('N'))"
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

try {
    $TarGzPath = Join-Path $TempDir "codecrafters.tar.gz"
    Write-Host "Downloading CodeCrafters CLI..."

    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $TarGzPath -UseBasicParsing
    } catch {
        Write-Error "Failed to download. Your platform and architecture (windows-$Arch) may be unsupported."
        exit 1
    }

    tar -xzf $TarGzPath -C $TempDir

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $ExtractedBinary = Join-Path $TempDir "codecrafters.exe"
    Move-Item -Path $ExtractedBinary -Destination $InstallPath -Force

    $InstalledVersion = & $InstallPath --version
    Write-Host "Installed $InstalledVersion"

    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
        $env:Path += ";$InstallDir" 
    }
    Write-Host "Done!"
} finally {
    # Cleanup
    if (Test-Path $TempDir) {
        Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}
