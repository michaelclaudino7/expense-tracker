# Smoke test: run with the API already running on localhost:8080
# Needs JWT_SECRET set in the terminal where the API runs.
$base = "http://localhost:8080"
$json = "application/json"

Write-Host "Health:"
Invoke-RestMethod -Uri "$base/health"

Write-Host "Registering a new user..."
$stamp = Get-Date -Format "yyyyMMddHHmmss"
$body = '{"email":"smoke' + $stamp + '@example.com","password":"smoke-test-123"}'
$auth = Invoke-RestMethod -Method Post -Uri "$base/auth/register" -ContentType $json -Body $body
$headers = @{ Authorization = "Bearer $($auth.token)" }

Write-Host "Creating categories..."
$food = Invoke-RestMethod -Method Post -Uri "$base/categories" -Headers $headers -ContentType $json -Body '{"name":"Smoke Food","type":"expense"}'
$pay  = Invoke-RestMethod -Method Post -Uri "$base/categories" -Headers $headers -ContentType $json -Body '{"name":"Smoke Pay","type":"income"}'

Write-Host "Creating transactions..."
$today = Get-Date -Format "yyyy-MM-dd"
$expense = '{"category_id":' + $food.id + ',"description":"Lunch","amount_cents":3500,"occurred_on":"' + $today + '"}'
$income  = '{"category_id":' + $pay.id + ',"description":"Freelance","amount_cents":120000,"occurred_on":"' + $today + '"}'
Invoke-RestMethod -Method Post -Uri "$base/transactions" -Headers $headers -ContentType $json -Body $expense | Out-Null
Invoke-RestMethod -Method Post -Uri "$base/transactions" -Headers $headers -ContentType $json -Body $income | Out-Null

Write-Host "Monthly report:"
$month = Get-Date -Format "yyyy-MM"
Invoke-RestMethod -Uri "$base/reports/monthly?month=$month" -Headers $headers | ConvertTo-Json -Depth 5