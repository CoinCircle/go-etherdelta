package helpers

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/ethclient"
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

	log.Infof("Connected EtherDelta client to provider: %s", ETH_PROVIDER_URI)
}
