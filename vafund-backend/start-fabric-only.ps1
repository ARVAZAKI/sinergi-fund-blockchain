# VaFund WSL Network Starter
# Script untuk start Fabric Network di WSL saja (tanpa backend)

Write-Host "VaFund WSL Network Starter" -ForegroundColor Green

# Check WSL
Write-Host "Checking WSL..." -ForegroundColor Cyan
$wslTest = wsl echo "OK"
if (-not $wslTest) {
    Write-Host "WSL not running" -ForegroundColor Red
    exit
}

# Check Fabric network status
Write-Host "Checking Fabric network..." -ForegroundColor Cyan
$fabricRunning = wsl docker ps --filter "name=peer0.org1.example.com" --format "{{.Status}}" | Where-Object { $_ -like "*Up*" }

# Also check if containers exist but are stopped
$fabricStopped = wsl docker ps -a --filter "name=peer0.org1.example.com" --format "{{.Status}}" | Where-Object { $_ -like "*Exit*" }

if ($fabricStopped -and -not $fabricRunning) {
    Write-Host "❌ Fabric containers exist but stopped (probably crashed)" -ForegroundColor Red
    Write-Host "Cleaning up crashed containers and restarting..." -ForegroundColor Yellow
    
    # Clean up crashed containers
    wsl docker stop $(wsl docker ps -aq) 2>/dev/null
    wsl docker rm $(wsl docker ps -aq) 2>/dev/null
    
    # Clean up any existing network first
    Write-Host "Cleaning up existing network..." -ForegroundColor Yellow
    wsl bash -c 'cd ~/fabric-workspace/fabric-samples/test-network && ./network.sh down'
    
    $fabricRunning = $false
}

if (-not $fabricRunning) {
    Write-Host "Fabric network not running properly. Cleaning and restarting..." -ForegroundColor Yellow
    
    # Clean up any existing network first
    Write-Host "Cleaning up existing network..." -ForegroundColor Yellow
    wsl bash -c 'cd ~/fabric-workspace/fabric-samples/test-network && ./network.sh down'
    
    # Start fresh network
    Write-Host "Starting fresh Fabric network..." -ForegroundColor Yellow
    wsl bash -c 'cd ~/fabric-workspace/fabric-samples/test-network && ./network.sh up createChannel -c vafundchannel -ca'
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Fabric network started successfully" -ForegroundColor Green
        
        # Deploy VaFund chaincode
        Write-Host "Deploying VaFund chaincode..." -ForegroundColor Yellow  
        wsl bash -c 'cd ~/fabric-workspace/fabric-samples/test-network && ./network.sh deployCC -ccn vafund -ccp ~/fabric-workspace/vafund-chaincode -ccl go -c vafundchannel'
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Chaincode deployed successfully" -ForegroundColor Green
        } else {
            Write-Host "Chaincode deployment had issues, but network is running" -ForegroundColor Yellow
        }
    } else {
        Write-Host "Failed to start Fabric network" -ForegroundColor Red
        exit
    }
} else {
    Write-Host "Fabric network is already running properly" -ForegroundColor Green
}

# Update crypto materials
Write-Host "Updating crypto materials..." -ForegroundColor Cyan
if (Test-Path "copy-crypto-from-wsl.ps1") {
    try {
        & ".\copy-crypto-from-wsl.ps1" | Out-Null
        Write-Host "Crypto materials updated" -ForegroundColor Green
    } catch {
        Write-Host "Crypto update had issues" -ForegroundColor Yellow
    }
} else {
    Write-Host "copy-crypto-from-wsl.ps1 not found" -ForegroundColor Yellow
}

# Test chaincode connectivity
Write-Host "Testing chaincode connectivity..." -ForegroundColor Cyan
$chaincodeCmd = @"
cd ~/fabric-workspace/fabric-samples/test-network
export PATH=../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=$PWD/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=$PWD/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
peer chaincode query -C vafundchannel -n vafund -c '{"function":"GetTotalDonations","Args":[]}'
"@

$chaincodeTest = wsl bash -c $chaincodeCmd

if ($LASTEXITCODE -eq 0) {
    Write-Host "Chaincode is responsive" -ForegroundColor Green
    Write-Host "Test result: $chaincodeTest" -ForegroundColor White
} else {
    Write-Host "Chaincode test failed, but network should be running" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Fabric Network Setup Complete!" -ForegroundColor Green
Write-Host "Network Details:" -ForegroundColor Cyan
Write-Host "   Channel: vafundchannel" -ForegroundColor White
Write-Host "   Chaincode: vafund" -ForegroundColor White
Write-Host "   Peer: localhost:7051" -ForegroundColor White

Write-Host ""
Write-Host "To start the backend:" -ForegroundColor Yellow
Write-Host "   go run main.go" -ForegroundColor White

Write-Host ""
Write-Host "Backend URLs (when started):" -ForegroundColor Yellow
Write-Host "   API: http://localhost:3000" -ForegroundColor White
Write-Host "   Swagger: http://localhost:3000/swagger/" -ForegroundColor White