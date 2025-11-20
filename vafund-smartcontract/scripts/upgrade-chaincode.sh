#!/bin/bash

# Script untuk upgrade VaFund Chaincode tanpa restart network
# Jalankan script ini dari direktori vafund-smartcontract

echo "🚀 Starting VaFund Chaincode Upgrade Process..."

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SMARTCONTRACT_DIR="$(dirname "$SCRIPT_DIR")"
CHAINCODE_PATH="$SMARTCONTRACT_DIR/chaincode"

echo "📂 Script directory: $SCRIPT_DIR"
echo "📂 Smartcontract directory: $SMARTCONTRACT_DIR"
echo "📂 Chaincode path: $CHAINCODE_PATH"

# Check if chaincode directory exists
if [ ! -d "$CHAINCODE_PATH" ]; then
    echo "❌ Error: chaincode directory not found at $CHAINCODE_PATH"
    exit 1
fi

# Try to find test-network directory
# Look in common locations relative to sinergi-fund-blockchain
PROJECT_ROOT="$(dirname "$SMARTCONTRACT_DIR")"
echo "📂 Project root: $PROJECT_ROOT"

# Check multiple possible locations
TEST_NETWORK_PATHS=(
    "$(dirname "$PROJECT_ROOT")/fabric-workspace/fabric-samples/test-network"
    "$HOME/fabric-samples/test-network"
    "$PROJECT_ROOT/../fabric-samples/test-network"
    "$PROJECT_ROOT/../../fabric-samples/test-network"
)

TEST_NETWORK_PATH=""
for path in "${TEST_NETWORK_PATHS[@]}"; do
    if [ -d "$path" ]; then
        TEST_NETWORK_PATH="$path"
        echo "✅ Found test-network at: $TEST_NETWORK_PATH"
        break
    fi
done

if [ -z "$TEST_NETWORK_PATH" ]; then
    echo "❌ Error: test-network directory not found!"
    echo "Searched in:"
    for path in "${TEST_NETWORK_PATHS[@]}"; do
        echo "  - $path"
    done
    echo ""
    echo "Please set TEST_NETWORK_PATH environment variable manually:"
    echo "  export TEST_NETWORK_PATH=/path/to/fabric-samples/test-network"
    echo "  ./scripts/upgrade-chaincode.sh"
    exit 1
fi

cd "$TEST_NETWORK_PATH" || exit

echo "📍 Current directory: $(pwd)"

# Set environment variables
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

CHAINCODE_NAME="vafundcc"
CHANNEL_NAME="vafundchannel"
# CHAINCODE_PATH is already set above

# Get current version
echo "🔍 Checking current chaincode version..."
CURRENT_VERSION=$(peer lifecycle chaincode querycommitted -C $CHANNEL_NAME -n $CHAINCODE_NAME 2>/dev/null | grep "Version:" | awk '{print $2}' | sed 's/,//')

if [ -z "$CURRENT_VERSION" ]; then
    echo "❌ Error: Could not find current chaincode version"
    echo "Is the chaincode deployed? Try deploying first."
    exit 1
fi

echo "📦 Current version: $CURRENT_VERSION"

# Calculate new version (increment)
if [[ $CURRENT_VERSION =~ ^([0-9]+)\.([0-9]+)$ ]]; then
    MAJOR="${BASH_REMATCH[1]}"
    MINOR="${BASH_REMATCH[2]}"
    NEW_MINOR=$((MINOR + 1))
    NEW_VERSION="${MAJOR}.${NEW_MINOR}"
else
    echo "⚠️  Cannot parse version, using 1.1"
    NEW_VERSION="1.1"
fi

echo "🆕 New version: $NEW_VERSION"
echo ""

# Step 1: Package the chaincode
echo "📦 Step 1: Packaging chaincode..."
peer lifecycle chaincode package ${CHAINCODE_NAME}.tar.gz \
    --path $CHAINCODE_PATH \
    --lang golang \
    --label ${CHAINCODE_NAME}_${NEW_VERSION}

if [ $? -ne 0 ]; then
    echo "❌ Failed to package chaincode"
    exit 1
fi
echo "✅ Chaincode packaged successfully"
echo ""

# Step 2: Install on Org1
echo "📥 Step 2: Installing chaincode on Org1 peer..."
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

peer lifecycle chaincode install ${CHAINCODE_NAME}.tar.gz

if [ $? -ne 0 ]; then
    echo "❌ Failed to install chaincode on Org1"
    exit 1
fi
echo "✅ Chaincode installed on Org1"
echo ""

# Step 3: Install on Org2
echo "📥 Step 3: Installing chaincode on Org2 peer..."
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=localhost:9051

peer lifecycle chaincode install ${CHAINCODE_NAME}.tar.gz

if [ $? -ne 0 ]; then
    echo "❌ Failed to install chaincode on Org2"
    exit 1
fi
echo "✅ Chaincode installed on Org2"
echo ""

# Step 4: Query installed chaincode to get package ID
echo "🔍 Step 4: Getting package ID..."
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051

PACKAGE_ID=$(peer lifecycle chaincode queryinstalled | grep ${CHAINCODE_NAME}_${NEW_VERSION} | awk '{print $3}' | sed 's/,//')

if [ -z "$PACKAGE_ID" ]; then
    echo "❌ Failed to get package ID"
    exit 1
fi

echo "📦 Package ID: $PACKAGE_ID"
echo ""

# Step 5: Approve for Org1
echo "✅ Step 5: Approving chaincode for Org1..."
peer lifecycle chaincode approveformyorg \
    -o localhost:7050 \
    --ordererTLSHostnameOverride orderer.example.com \
    --tls \
    --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" \
    --channelID $CHANNEL_NAME \
    --name $CHAINCODE_NAME \
    --version $NEW_VERSION \
    --package-id $PACKAGE_ID \
    --sequence 2

if [ $? -ne 0 ]; then
    echo "❌ Failed to approve chaincode for Org1"
    exit 1
fi
echo "✅ Chaincode approved for Org1"
echo ""

# Step 6: Approve for Org2
echo "✅ Step 6: Approving chaincode for Org2..."
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=localhost:9051

peer lifecycle chaincode approveformyorg \
    -o localhost:7050 \
    --ordererTLSHostnameOverride orderer.example.com \
    --tls \
    --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" \
    --channelID $CHANNEL_NAME \
    --name $CHAINCODE_NAME \
    --version $NEW_VERSION \
    --package-id $PACKAGE_ID \
    --sequence 2

if [ $? -ne 0 ]; then
    echo "❌ Failed to approve chaincode for Org2"
    exit 1
fi
echo "✅ Chaincode approved for Org2"
echo ""

# Step 7: Check commit readiness
echo "🔍 Step 7: Checking commit readiness..."
peer lifecycle chaincode checkcommitreadiness \
    --channelID $CHANNEL_NAME \
    --name $CHAINCODE_NAME \
    --version $NEW_VERSION \
    --sequence 2 \
    --output json

echo ""

# Step 8: Commit the chaincode
echo "🚀 Step 8: Committing chaincode upgrade..."
peer lifecycle chaincode commit \
    -o localhost:7050 \
    --ordererTLSHostnameOverride orderer.example.com \
    --tls \
    --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" \
    --channelID $CHANNEL_NAME \
    --name $CHAINCODE_NAME \
    --version $NEW_VERSION \
    --sequence 2 \
    --peerAddresses localhost:7051 \
    --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" \
    --peerAddresses localhost:9051 \
    --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"

if [ $? -ne 0 ]; then
    echo "❌ Failed to commit chaincode"
    exit 1
fi
echo "✅ Chaincode committed successfully!"
echo ""

# Step 9: Verify the upgrade
echo "🔍 Step 9: Verifying chaincode upgrade..."
peer lifecycle chaincode querycommitted -C $CHANNEL_NAME -n $CHAINCODE_NAME

echo ""
echo "✨ Testing new gallery functions..."
peer chaincode query -C $CHANNEL_NAME -n $CHAINCODE_NAME -c '{"function":"GetAllGalleries","Args":[]}'

echo ""
echo "🎉 Chaincode upgrade completed successfully!"
echo "📋 Summary:"
echo "   Chaincode: $CHAINCODE_NAME"
echo "   Old Version: $CURRENT_VERSION"
echo "   New Version: $NEW_VERSION"
echo "   Channel: $CHANNEL_NAME"
echo ""
echo "🔄 Please restart your backend to use the new chaincode version"
