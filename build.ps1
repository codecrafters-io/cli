param($Command)

$BinDir = "$env:USERPROFILE\bin"
$ExePath = "$BinDir\codecrafters.exe"

switch ($Command) {
    "install" {
        if (!(Test-Path $BinDir)) { mkdir $BinDir }
        go build -o $ExePath cmd/codecrafters/main.go
    }
    
    "uninstall" {
        if (Test-Path $ExePath) { Remove-Item $ExePath }
    }
    
    "release" {
        $currentVersion = (git tag --list "v*" | Sort-Object | Select-Object -Last 1) -replace "v", ""
        if (!$currentVersion) { $currentVersion = "0" }
        $nextVersion = [int]$currentVersion + 1
        git tag "v$nextVersion"
        git push origin main "v$nextVersion"
    }
    
    "test" {
        go test -v ./...
    }
    
    default {
        Write-Host "Usage: .\build.ps1 [install|uninstall|release|test]"
    }
}