$repo = "neutron420/StackAudit"
$response = Invoke-WebRequest -Uri "https://api.github.com/repos/$repo/releases/latest" -UseBasicParsing
$release = $response.Content | ConvertFrom-Json
$asset = $release.assets | Where-Object { $_.name -match "windows" -and $_.name -match "amd64" -and $_.name -like "*.zip" }
if (!$asset) {
    Write-Host "Error: Could not find Windows release asset. The build might still be in progress on GitHub." -ForegroundColor Red
    exit
}
$url = $asset.browser_download_url
$dest = "$HOME\stack.zip"
$binDir = "$HOME\.stack\bin"

if (!(Test-Path $binDir)) { New-Item -ItemType Directory -Path $binDir }

Write-Host "Downloading stack from $url..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $url -OutFile $dest -UseBasicParsing
Expand-Archive -Path $dest -DestinationPath $binDir -Force
Remove-Item $dest

$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*$binDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$path;$binDir", "User")
    Write-Host "Added $binDir to PATH. Please restart your terminal." -ForegroundColor Green
}

Write-Host "stack installed successfully! Type 'stack' to begin." -ForegroundColor Green
