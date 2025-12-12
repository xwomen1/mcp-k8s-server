# PowerShell script to test MCP Kubernetes Server
# Usage: .\test.ps1

Write-Host "Testing MCP Kubernetes Server" -ForegroundColor Green
Write-Host "==============================" -ForegroundColor Green
Write-Host ""

# Start server in background
Write-Host "Starting server..." -ForegroundColor Yellow
$serverProcess = Start-Process -FilePath "go" -ArgumentList "run", "../cmd/server/main.go" -NoNewWindow -PassThru -RedirectStandardInput "server_input.txt" -RedirectStandardOutput "server_output.txt" -RedirectStandardError "server_error.txt"

Start-Sleep -Seconds 2

# Test 1: List tools
Write-Host "Test 1: Listing available tools..." -ForegroundColor Cyan
$request1 = '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | Out-File -FilePath "request1.json" -Encoding utf8 -NoNewline
Write-Host "Request: $request1"
Write-Host "Check server_output.txt for response"
Write-Host ""

Start-Sleep -Seconds 1

# Cleanup
Stop-Process -Id $serverProcess.Id -Force -ErrorAction SilentlyContinue

Write-Host "Test completed. Check server_output.txt for responses." -ForegroundColor Green

