const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const repo = "neutron420/StackAudit";
const binDir = path.join(__dirname, '..', 'bin');
if (!fs.existsSync(binDir)) fs.mkdirSync(binDir);

console.log("Installing StackAudit binary...");

try {
  if (process.platform === 'win32') {
    execSync(`powershell -NoProfile -Command "iwr -UseBasicParsing https://raw.githubusercontent.com/${repo}/main/scripts/install.ps1 | iex"`, { stdio: 'inherit' });
    const installedBin = path.join(process.env.USERPROFILE, '.stack', 'bin', 'stack.exe');
    if (fs.existsSync(installedBin)) {
        fs.copyFileSync(installedBin, path.join(binDir, 'stack.exe'));
    }
  } else {
    execSync(`curl -sSL https://raw.githubusercontent.com/${repo}/main/scripts/install.sh | sh`, { stdio: 'inherit' });
    if (fs.existsSync('/usr/local/bin/stack')) {
        fs.copyFileSync('/usr/local/bin/stack', path.join(binDir, 'stack'));
    }
  }
  console.log("NPM installation complete!");
} catch (err) {
  console.error("Failed to install binary:", err.message);
  process.exit(1);
}
