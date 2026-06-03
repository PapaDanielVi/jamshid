<#
.SYNOPSIS
    Jamshid installation script for Windows systems.

.DESCRIPTION
    Downloads and installs the latest jamshid release from GitHub.
    Automatically detects system architecture and installs to the appropriate location.

.PARAMETER InstallDir
    Optional. The installation directory. Defaults to $env:ProgramFiles\jamshid.

.EXAMPLE
    .\install-windows.ps1
    Installs jamshid to the default location.

.EXAMPLE
    .\install-windows.ps1 -InstallDir "C:\Tools\jamshid"
    Installs jamshid to a custom location.
#>

param(
    [string]$InstallDir = "$env:ProgramFiles\jamshid"
)

$ErrorActionPreference = "Stop"

$Repo = "PapaDanielVi/jamshid"
$BinaryName = "jamshid.exe"

# Get the latest release tag from GitHub API
function Get-LatestRelease {
    $response = Invoke-RestMethod "https://api.github.com/repos/${Repo}/releases/latest"
    return $response.tag_name
}

# Detect system architecture
function Get-Architecture {
    if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") {
        return "Windows_x86_64"
    } elseif ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
        return "Windows_arm64"
    } else {
        Write-Host "Error: Unsupported architecture $($env:PROCESSOR_ARCHITECTURE)"
        exit 1
    }
}

# Check if jamshid is already installed
function Test-Installed {
    $existing = Get-Command jamshid -ErrorAction SilentlyContinue
    if ($existing) {
        Write-Host "Jamshid is already installed: $($existing.Source)"
        return $true
    }
    return $false
}

# Download and install the binary
function Install-Binary {
    param(
        [string]$Architecture,
        [string]$Release
    )

    $url = "https://github.com/${Repo}/releases/download/${Release}/jamshid_${Architecture}.tar.gz"

    Write-Host "Downloading binary tarball for $Architecture..."
    $tempFile = Join-Path $env:TEMP "jamshid.tar.gz"

    Invoke-WebRequest -Uri $url -OutFile $tempFile

    # Create install directory if it doesn't exist
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    # Extract tarball using tar (available in Windows 10+)
    Write-Host "Extracting and installing to $InstallDir..."
    $tempDir = Join-Path $env:TEMP "jamshid_extract"
    if (-not (Test-Path $tempDir)) {
        New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
    }

    tar -xzf $tempFile -C $tempDir
    Copy-Item (Join-Path $tempDir $BinaryName) (Join-Path $InstallDir $BinaryName) -Force

    # Clean up
    Remove-Item $tempFile -Force
    Remove-Item $tempDir -Recurse -Force

    Write-Host "Installing to $InstallDir..."
    # Add to user PATH if not already present
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
        Write-Host "Added $InstallDir to user PATH. Restart your terminal for changes to take effect."
    }
}

# Main installation logic
function Main {
    $arch = Get-Architecture
    $release = Get-LatestRelease

    Write-Host "Installing jamshid $release for $arch..."
    Install-Binary -Architecture $arch -Release $release

    Write-Host "Jamshid installed successfully!"
}

# Check if already installed before proceeding
if (-not (Test-Installed)) {
    Main
} else {
    Write-Host "Jamshid is already installed. Re-running to update to latest version..."
    Main
}