# Copy crypto materials from WSL to Windows
Write-Host "Copying crypto materials from WSL..." -ForegroundColor Yellow

# Create crypto directory if not exists
if (-not (Test-Path "crypto")) {
    New-Item -ItemType Directory -Path "crypto"
    Write-Host "Created crypto directory" -ForegroundColor Green
}

# Copy Admin cert
$adminCertPath = "organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/cert.pem"
wsl cp ~/fabric-workspace/fabric-samples/test-network/$adminCertPath /mnt/c/coding/vafund-blockchain/vafund-backend/crypto/cert.pem

# Copy Admin private key (find the key file)
$adminKeyDir = "organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore"
$keyFile = wsl bash -c "ls ~/fabric-workspace/fabric-samples/test-network/$adminKeyDir/ | head -1"
if ($keyFile) {
    wsl cp ~/fabric-workspace/fabric-samples/test-network/$adminKeyDir/$keyFile /mnt/c/coding/vafund-blockchain/vafund-backend/crypto/key.pem
    Write-Host "Copied admin private key: $keyFile" -ForegroundColor Green
} else {
    Write-Host "Admin private key not found!" -ForegroundColor Red
}

# Copy TLS CA cert
$tlsCaPath = "organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem"
wsl cp ~/fabric-workspace/fabric-samples/test-network/$tlsCaPath /mnt/c/coding/vafund-blockchain/vafund-backend/crypto/tls-ca-cert.pem

# Copy peer TLS CA cert (alternative)
$peerTlsCaPath = "organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
wsl cp ~/fabric-workspace/fabric-samples/test-network/$peerTlsCaPath /mnt/c/coding/vafund-blockchain/vafund-backend/crypto/peer-tls-ca.pem

# Copy orderer TLS CA cert
$ordererTlsCaPath = "organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem"
wsl cp ~/fabric-workspace/fabric-samples/test-network/$ordererTlsCaPath /mnt/c/coding/vafund-blockchain/vafund-backend/crypto/orderer-tls-ca.pem

Write-Host "Crypto materials copied successfully!" -ForegroundColor Green

# List copied files
Write-Host "Files in crypto directory:" -ForegroundColor Cyan
Get-ChildItem -Path "crypto" | Format-Table Name, Length, LastWriteTime