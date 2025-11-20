#!/bin/bash

# Script untuk restart Fabric Network dan VaFund Chaincode
# Jalankan script ini dari direktori test-network

echo "🚀 Starting Fabric Network Restart Process..."

# Step 1: Bersihkan network yang ada (jika ada)
echo "🧹 Cleaning up existing network..."
./network.sh down

# Step 2: Start network dengan channel
echo "📡 Starting Fabric network with channel..."
./network.sh up createChannel -ca

# Step 3: Deploy VaFund chaincode
echo "📦 Deploying VaFund chaincode..."
./network.sh deployCC -ccn vafund -ccp ../../../vafund-smartcontract/chaincode -ccl go -c vafundchannel

# Step 4: Verifikasi deployment
echo "✅ Verifying chaincode deployment..."
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

echo "🔍 Testing chaincode with GetAllDonations..."
peer chaincode query -C vafundchannel -n vafund -c '{"function":"GetAllDonations","Args":[]}'

echo "📋 Current donations in chaincode:"
peer chaincode query -C vafundchannel -n vafund -c '{"function":"GetTotalDonations","Args":[]}'

echo "🎉 Fabric Network is ready!"
echo ""
echo "📝 Network Information:"
echo "   Channel: vafundchannel"
echo "   Chaincode: vafund"
echo "   Peer: localhost:7051"
echo ""
echo "🔗 To connect from Windows backend:"
echo "   1. Make sure backend .env has correct settings"
echo "   2. Run: go run main.go"
echo "   3. Test API: http://localhost:3000/swagger/"