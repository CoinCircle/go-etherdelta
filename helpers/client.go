package helpers

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"os"
)

var Client *ethclient.Client
var ETH_PROVIDER_URI = "wss://mainnet.infura.io/ws"

func init() {
	providerUriEnv := os.Getenv("ETH_PROVIDER_URI")

	if providerUriEnv != "" {
		ETH_PROVIDER_URI = providerUriEnv
	}

	if ETH_PROVIDER_URI == "" {
		log.Fatal("ETH_PROVIDER_URI environment variable is required")
	}

	// Create an IPC based RPC connection to a remote node
	client, err := ethclient.Dial(ETH_PROVIDER_URI)
	Client = client

	if err != nil {
		panic(err)
	}

	log.Println("Connected EtherDelta client to provider:", ETH_PROVIDER_URI)
}
