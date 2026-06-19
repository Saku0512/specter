$ErrorActionPreference = "Stop"

$repo    = "Saku0512/specter"
$installDir = "$env:USERPROFILE\.local\bin"

function Show-Banner {
    Write-Host "        .-." -ForegroundColor Cyan
    Write-Host "      (o o) boo" -ForegroundColor Cyan
    Write-Host "      | O \" -ForegroundColor Cyan
    Write-Host "       \   \" -ForegroundColor Cyan
    Write-Host "        ``~~~'" -ForegroundColor Cyan
    Write-Host "specter installer" -ForegroundColor White
    Write-Host ""
}

function Write-Step([string]$Message) {
    Write-Host "==> " -NoNewline -ForegroundColor Cyan
    Write-Host $Message -ForegroundColor White
}

function Write-Warn([string]$Message) {
    Write-Host "==> " -NoNewline -ForegroundColor Yellow
    Write-Host $Message -ForegroundColor White
}

function Write-Success([string]$Message) {
    Write-Host "==> " -NoNewline -ForegroundColor Green
    Write-Host $Message -ForegroundColor White
}

Show-Banner

Write-Step "Fetching latest release metadata"
$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$version = $release.tag_name

Write-Step "Installing specter $version"

$url  = "https://github.com/$repo/releases/download/$version/specter_windows_amd64.exe"
$dest = "$installDir\specter.exe"

Write-Step "Creating install directory"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Write-Step "Downloading specter_windows_amd64.exe"
Invoke-WebRequest -Uri $url -OutFile $dest

# Add to PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Warn "Added $installDir to PATH (restart your terminal to apply)"
}

Write-Success "specter installed to $dest"
Write-Host "Run: " -NoNewline -ForegroundColor White
Write-Host "specter -c config.yml -p 8080" -ForegroundColor Cyan
