# Deploys go-api, node-api, and frontend as three separate Railway services
# in one project, wired together with the right env vars and private networking.
#
# Run this from the repo root: .\deploy-railway.ps1
#
# Why a script at all: Railway's Railpack builder looks at the repo root and
# gives up because there's no single-language app there, just three
# subdirectories plus a docker-compose.yml it doesn't orchestrate. Each
# service already has its own Dockerfile and a railway.json forcing the
# DOCKERFILE builder, so deploying each subdirectory as its own Railway
# service sidesteps Railpack's detection entirely.

$ErrorActionPreference = "Stop"
$repoRoot = $PSScriptRoot

function Assert-LastExitCode($step) {
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed: $step (exit code $LASTEXITCODE)"
        exit 1
    }
}

# 1. Railway CLI
if (-not (Get-Command railway -ErrorAction SilentlyContinue)) {
    Write-Host "Railway CLI not found, installing via npm..."
    npm install -g @railway/cli
    Assert-LastExitCode "npm install -g @railway/cli"
}

# 2. Login (opens a browser, only asks if you're not already logged in)
railway whoami 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "Not logged in, opening Railway login..."
    railway login
    Assert-LastExitCode "railway login"
}

# 3. Create the project (skip this line and run 'railway link' instead if you already have one)
Set-Location $repoRoot
railway init --name matrix-ops
Assert-LastExitCode "railway init"

# 4. A real shared secret, both APIs validate JWTs with it independently
$secretBytes = New-Object byte[] 32
[System.Security.Cryptography.RandomNumberGenerator]::Fill($secretBytes)
$jwtSecret = [Convert]::ToBase64String($secretBytes)

# 5. node-api first: no public domain, reachable only over Railway's private network
Set-Location "$repoRoot\node-api"
railway up --service node-api --detach
Assert-LastExitCode "deploy node-api"
railway variables --service node-api --set "JWT_SECRET=$jwtSecret"
Assert-LastExitCode "set node-api JWT_SECRET"

# 6. go-api: same secret, talks to node-api over Railway's private network (service-name.railway.internal)
Set-Location "$repoRoot\go-api"
railway up --service go-api --detach
Assert-LastExitCode "deploy go-api"
railway variables --service go-api --set "JWT_SECRET=$jwtSecret" --set "NODE_API_URL=http://node-api.railway.internal:3000"
Assert-LastExitCode "set go-api variables"

Write-Host ""
Write-Host "Generating a public domain for go-api..."
railway domain --service go-api
Assert-LastExitCode "railway domain go-api"

Write-Host ""
$goApiDomain = Read-Host "Paste the go-api domain Railway just printed above (e.g. go-api-production.up.railway.app, no https://)"
$goApiUrl = "https://$goApiDomain"

# 7. frontend: config.js gets generated from API_BASE_URL when the container starts (see docker-entrypoint.d)
Set-Location "$repoRoot\frontend"
railway up --service frontend --detach
Assert-LastExitCode "deploy frontend"
railway variables --service frontend --set "API_BASE_URL=$goApiUrl"
Assert-LastExitCode "set frontend API_BASE_URL"

Write-Host ""
Write-Host "Generating a public domain for frontend..."
railway domain --service frontend
Assert-LastExitCode "railway domain frontend"

Set-Location $repoRoot
Write-Host ""
Write-Host "Done. All three services are deploying. Check progress with: railway status"
Write-Host "The frontend and go-api need a moment to redeploy after the variables above were set - trigger it with 'railway up --service <name>' again if they don't pick it up automatically."
