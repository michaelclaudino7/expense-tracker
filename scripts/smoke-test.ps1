# Smoke test: run with the API already running on localhost:8080
$base = "http://localhost:8080"
$json = "application/json"

Write-Host "Health:"
Invoke-RestMethod -Uri "$base/health"

Write-Host "Creating categories..."
$food = Invoke-RestMethod -Method Post -Uri "$base/categories" -ContentType $json -Body '{"name":"Smoke Food","type":"expense"}'
$pay  = Invoke-RestMethod -Method Post -Uri "$base/categories" -ContentType $json -Body '{"name":"Smoke Pay","type":"income"}'

Write-Host "Creating transactions..."
$today = Get-Date -Format "yyyy-MM-dd"
$expense = '{"category_id":' + $food.id + ',"description":"Lunch","amount_cents":3500,"occurred_on":"' + $today + '"}'
$income  = '{"category_id":' + $pay.id + ',"description":"Freelance","amount_cents":120000,"occurred_on":"' + $today + '"}'
Invoke-RestMethod -Method Post -Uri "$base/transactions" -ContentType $json -Body $expense | Out-Null
Invoke-RestMethod -Method Post -Uri "$base/transactions" -ContentType $json -Body $income | Out-Null

Write-Host "Monthly report:"
$month = Get-Date -Format "yyyy-MM"
Invoke-RestMethod -Uri "$base/reports/monthly?month=$month" | ConvertTo-Json -Depth 5