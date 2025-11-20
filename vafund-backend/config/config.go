package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Application Settings
	Port        string
	Environment string

	// Fabric Configuration
	FabricPeerHost    string
	FabricPeerPort    string
	FabricPeerTLSName string
	FabricOrgMSP      string
	FabricChannel     string
	FabricChaincode   string

	// Crypto Paths
	FabricCertPath  string
	FabricKeyPath   string
	FabricTLSCAPath string

	// Application Settings
	SimulationMode bool

	// Fallback Hosts
	FabricFallbackHosts []string
}

var AppConfig *Config

func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	config := &Config{
		Port:        getEnv("APP_PORT", "3000"),
		Environment: getEnv("APP_MODE", "development"),

		FabricPeerHost:    getEnv("FABRIC_PEER_HOST", "172.31.201.95"),
		FabricPeerPort:    getEnv("FABRIC_PEER_PORT", "7051"),
		FabricPeerTLSName: getEnv("FABRIC_PEER_TLS_NAME", "peer0.org1.example.com"),
		FabricOrgMSP:      getEnv("FABRIC_ORG_MSP", "Org1MSP"),
		FabricChannel:     getEnv("FABRIC_CHANNEL", "vafundchannel"),
		FabricChaincode:   getEnv("FABRIC_CHAINCODE", "vafund"),

		FabricCertPath:  getEnv("FABRIC_CERT_PATH", "./crypto/cert.pem"),
		FabricKeyPath:   getEnv("FABRIC_KEY_PATH", "./crypto/key.pem"),
		FabricTLSCAPath: getEnv("FABRIC_TLS_CA_PATH", "./crypto/tls-ca-cert.pem"),

		SimulationMode: getEnvBool("SIMULATION_MODE", false),
	}

	// Parse fallback hosts
	fallbackHosts := getEnv("FABRIC_FALLBACK_HOSTS", "localhost:7051,127.0.0.1:7051")
	config.FabricFallbackHosts = strings.Split(fallbackHosts, ",")

	AppConfig = config
	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func (c *Config) GetFabricPeerAddress() string {
	return c.FabricPeerHost + ":" + c.FabricPeerPort
}

func (c *Config) GetAllHosts() []string {
	hosts := []string{c.GetFabricPeerAddress()}
	hosts = append(hosts, c.FabricFallbackHosts...)
	return hosts
}
