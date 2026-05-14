$repo = "neutron420/StackAudit"
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
$asset = $release.assets | Where-Object { $_.name -like "*Windows_x86_64.zip" }
$url = $asset.browser_download_url
$dest = "$HOME\stack.zip"
$binDir = "$HOME\.stack\bin"

if (!(Test-Path $binDir)) { New-Item -ItemType Directory -Path $binDir }

Write-Host "Downloading stack from $url..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $url -OutFile $dest
Expand-Archive -Path $dest -DestinationPath $binDir -Force
Remove-Item $dest

$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*$binDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$path;$binDir", "User")
    Write-Host "Added $binDir to PATH. Please restart your terminal." -ForegroundColor Green
}

Write-Host "stack installed successfully! Type 'stack' to begin." -ForegroundColor Green
