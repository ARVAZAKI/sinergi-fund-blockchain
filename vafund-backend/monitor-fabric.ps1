# Monitor Fabric Network Health
param(
    [switch]$Continuous,
    [int]$Interval = 30
)

function Check-FabricHealth {
    Write-Host "🔍 Checking Fabric network health..." -ForegroundColor Cyan
    
    # Check if containers are running
    $runningContainers = @(wsl docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "(peer0|orderer|ca_)")
    $stoppedContainers = @(wsl docker ps -a --format "table {{.Names}}\t{{.Status}}" | grep -E "(peer0|orderer|ca_)" | grep -v "Up")
    
    Write-Host "Running containers:" -ForegroundColor Green
    if ($runningContainers.Count -gt 0) {
        $runningContainers | ForEach-Object { Write-Host "  $_" -ForegroundColor White }
    } else {
        Write-Host "  None" -ForegroundColor Red
    }
    
    if ($stoppedContainers.Count -gt 0) {
        Write-Host "Stopped/crashed containers:" -ForegroundColor Red
        $stoppedContainers | ForEach-Object { Write-Host "  $_" -ForegroundColor Yellow }
    }
    
    # Check chaincode connectivity if network is running
    if ($runningContainers.Count -ge 3) {
        Write-Host "Testing chaincode connectivity..." -ForegroundColor Cyan
        $chaincodeCmd = @"
cd ~/fabric-workspace/fabric-samples/test-network
export PATH=../bin:`$PATH
export FABRIC_CFG_PATH=`$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=`$PWD/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=`$PWD/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
peer chaincode query -C vafundchannel -n vafund -c '{"function":"GetTotalDonations","Args":[]}' 2>/dev/null
"@
        
        $result = wsl bash -c $chaincodeCmd
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Chaincode is responsive: $result" -ForegroundColor Green
        } else {
            Write-Host "❌ Chaincode test failed" -ForegroundColor Red
        }
    }
    
    Write-Host "Health check completed at $(Get-Date)" -ForegroundColor Gray
    Write-Host "-" * 50 -ForegroundColor Gray
}

if ($Continuous) {
    Write-Host "Starting continuous monitoring every $Interval seconds..." -ForegroundColor Yellow
    Write-Host "Press Ctrl+C to stop" -ForegroundColor Yellow
    
    while ($true) {
        Check-FabricHealth
        Start-Sleep -Seconds $Interval
    }
} else {
    Check-FabricHealth
}