# Check if message.json exists
if (-not (Test-Path "message.json")) {
    Write-Output "Error: message.json not found!"
    exit 1
}

# Read JSON content
# Remove newlines and extra spaces
$message = Get-Content -Path "message.json" -Raw
$message = $message -replace '\s+', ' '
$message = $message -replace ' ,', ','

# Push the compacted message to Kafka topic using docker exec
Echo $message | docker exec -i kafka /usr/bin/kafka-console-producer --broker-list localhost:9092 --topic delivery-topic

# Confirmation
if ($LASTEXITCODE -eq 0) {
    Write-Output "Message pushed successfully to delivery-topic!"
} else {
    Write-Output "Error pushing message to Kafka!"
}
