$body = '{"username":"testuser","password":"123456","name":"Test User"}'
Write-Host "=== Testing Register ==="
try {
  $r = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/register" -Method POST -Body $body -ContentType "application/json"
  Write-Host $r | ConvertTo-Json
} catch {
  Write-Host "Register Error: $_"
  Write-Host "Status: $($_.Exception.Response.StatusCode.value__)"
  if ($_.Exception.Response) {
    $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
    $responseBody = $reader.ReadToEnd()
    $reader.Close()
    Write-Host "Body: $responseBody"
  }
}

$body2 = '{"username":"testuser","password":"123456"}'
Write-Host "`n=== Testing Login ==="
try {
  $r2 = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -Body $body2 -ContentType "application/json"
  Write-Host ($r2 | ConvertTo-Json)
} catch {
  Write-Host "Login Error: $_"
  Write-Host "Status: $($_.Exception.Response.StatusCode.value__)"
  if ($_.Exception.Response) {
    $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
    $responseBody = $reader.ReadToEnd()
    $reader.Close()
    Write-Host "Body: $responseBody"
  }
}
