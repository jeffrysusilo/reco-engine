# PowerShell script to generate sample events

param(
    [string]$BaseUrl = "http://localhost:8080",
    [int]$NumEvents = 1000
)

Write-Host "Generating $NumEvents sample events to $BaseUrl"

$EventTypes = @("VIEW", "CLICK", "CART", "PURCHASE")
$Sessions = @()

# Generate some session IDs
for ($i = 1; $i -le 50; $i++) {
    $Sessions += "session_" + [guid]::NewGuid().ToString()
}

for ($i = 1; $i -le $NumEvents; $i++) {
    $UserId = Get-Random -Minimum 1 -Maximum 101
    $ItemId = Get-Random -Minimum 1 -Maximum 51
    $EventType = $EventTypes[(Get-Random -Minimum 0 -Maximum 4)]
    $SessionId = $Sessions[(Get-Random -Minimum 0 -Maximum 50)]
    
    $Body = @{
        user_id = $UserId
        item_id = $ItemId
        event_type = $EventType
        session_id = $SessionId
    } | ConvertTo-Json

    try {
        Invoke-RestMethod -Uri "$BaseUrl/events" -Method Post -Body $Body -ContentType "application/json" | Out-Null
        
        if ($i % 100 -eq 0) {
            Write-Host "Generated $i events..."
        }
        
        Start-Sleep -Milliseconds 10
    }
    catch {
        Write-Warning "Failed to send event $i : $_"
    }
}

Write-Host "Done! Generated $NumEvents events"
