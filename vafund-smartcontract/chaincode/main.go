package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	assetChaincode, err := contractapi.NewChaincode(NewSmartContract())
	if err != nil {
		log.Panicf("Error creating vafund chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting vafund chaincode: %v", err)
	}
}
