# Launches Swagger UI locally for docs/openapi.yaml

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

docker run --rm `
  -p 8081:8080 `
  -e SWAGGER_JSON=/spec/openapi.yaml `
  -v "${repoRoot}/docs:/spec" `
  swaggerapi/swagger-ui
