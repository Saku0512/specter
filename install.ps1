$ErrorActionPreference = "Stop"

$repo    = "Saku0512/specter"
$installDir = "$env:USERPROFILE\.local\bin"

Write-Host "Fetching latest release..."
$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$version = $release.tag_name

Write-Host "Installing specter $version..."

$url  = "https://github.com/$repo/releases/download/$version/specter_windows_amd64.exe"
$dest = "$installDir\specter.exe"

New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Invoke-WebRequest -Uri $url -OutFile $dest

# Add to PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Host "Added $installDir to PATH (restart your terminal to apply)"
}

Write-Host "specter installed to $dest"
Write-Host "Run: specter -c config.yml -p 8080"
