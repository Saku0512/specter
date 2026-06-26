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

$filename = "specter_windows_amd64.exe"
$url  = "https://github.com/$repo/releases/download/$version/$filename"
$checksumUrl = "https://github.com/$repo/releases/download/$version/SHA256SUMS.txt"
$dest = "$installDir\specter.exe"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) "specter-$([guid]::NewGuid()).exe"
$checksums = Join-Path ([System.IO.Path]::GetTempPath()) "specter-$([guid]::NewGuid())-SHA256SUMS.txt"

Write-Step "Creating install directory"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Write-Step "Downloading $filename"
Invoke-WebRequest -Uri $url -OutFile $tmp

Write-Step "Verifying $filename"
Invoke-WebRequest -Uri $checksumUrl -OutFile $checksums
$expectedLine = Get-Content $checksums | Where-Object { $_ -match "\s+$([regex]::Escape($filename))$" } | Select-Object -First 1
if (-not $expectedLine) {
    throw "Checksum for $filename was not found in SHA256SUMS.txt"
}

$expected = (($expectedLine -split "\s+")[0]).ToLowerInvariant()
$actual = (Get-FileHash $tmp -Algorithm SHA256).Hash.ToLowerInvariant()
if ($actual -ne $expected) {
    throw "Checksum mismatch for $filename"
}

Move-Item -Path $tmp -Destination $dest -Force
Remove-Item -Path $checksums -Force -ErrorAction SilentlyContinue

# Add to PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Warn "Added $installDir to PATH (restart your terminal to apply)"
}

Write-Success "specter installed to $dest"
Write-Host "Run: " -NoNewline -ForegroundColor White
Write-Host "specter -c config.yml -p 8080" -ForegroundColor Cyan
