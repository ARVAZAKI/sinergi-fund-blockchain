# Deploy VaFund Chaincode Script
Write-Host "Deploying VaFund chaincode to Fabric network..." -ForegroundColor Green

# Set paths
$chaincodeSource = "c:\coding\vafund-blockchain\vafund-smartcontract\chaincode"
$wslChaincodePath = "/home/arva/fabric-workspace/vafund-chaincode"

# Copy chaincode to WSL
Write-Host "Copying chaincode to WSL..." -ForegroundColor Yellow
wsl rm -rf $wslChaincodePath
wsl mkdir -p $wslChaincodePath

# Copy files to WSL
Copy-Item "$chaincodeSource\*" -Destination "\\wsl`$\Ubuntu\home\arva\fabric-workspace\vafund-chaincode\" -Recurse -Force

Write-Host "Chaincode copied to WSL successfully" -ForegroundColor Green

# Deploy chaincode using fabric network script
Write-Host "Deploying chaincode..." -ForegroundColor Yellow
$deployResult = wsl bash -c 'cd ~/fabric-workspace/fabric-samples/test-network && ./network.sh deployCC -ccn vafund -ccp ~/fabric-workspace/vafund-chaincode -ccl go -c vafundchannel'

if ($LASTEXITCODE -eq 0) {
    Write-Host "VaFund chaincode deployed successfully!" -ForegroundColor Green
    
    # Test chaincode
    Write-Host "Testing chaincode..." -ForegroundColor Cyan
    $testCmd = @"
cd ~/fabric-workspace/fabric-samples/test-network
export PATH=`$PATH:../bin
export FABRIC_CFG_PATH=`$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=`$PWD/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=`$PWD/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
peer chaincode query -C vafundchannel -n vafund -c '{"function":"GetTotalDonations","Args":[]}'
"@

    $testResult = wsl bash -c $testCmd
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Chaincode test successful!" -ForegroundColor Green
        Write-Host "Test result: $testResult" -ForegroundColor White
    } else {
        Write-Host "Chaincode test failed, but deployment seems successful" -ForegroundColor Yellow
    }
} else {
    Write-Host "Chaincode deployment failed" -ForegroundColor Red
}

Write-Host ""
Write-Host "VaFund chaincode deployment complete!" -ForegroundColor Green