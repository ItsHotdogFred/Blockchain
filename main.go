package main

import (
	"fmt"
	"os"

	"github.com/ItsHotdogFred/blockchain/blockchain"
	"github.com/ItsHotdogFred/blockchain/cli"
	"github.com/ItsHotdogFred/blockchain/network"
)

func main() {
	defer os.Exit(0)

	if len(os.Args) > 1 && os.Args[1] == "server" {
		// Run server mode only if "server" argument is explicitly provided
		nodeID := os.Getenv("NODE_ID")

		// Try to continue existing blockchain, if it doesn't exist, create a new one
		var chain *blockchain.BlockChain

		// Check if blockchain exists
		path := fmt.Sprintf("./tmp/blocks_%s", nodeID)
		if blockchain.DBexists(path) {
			fmt.Println("Continuing existing blockchain...")
			chain = blockchain.ContinueBlockChain(nodeID)
		} else {
			fmt.Println("No existing blockchain found, creating new one...")
			// Create a genesis block with a default address for server initialization
			genesisAddress := "16X5ieK8C7M36wXNq1t3Uj7QSsGcwfsQaU"
			chain = blockchain.InitBlockChain(genesisAddress, nodeID)

			// Reindex UTXO set after creation
			UTXOSet := blockchain.UTXOSet{Blockchain: chain}
			UTXOSet.Reindex()
			fmt.Println("Blockchain created successfully!")
		}

		defer chain.Database.Close()

		go network.StartServer(nodeID, chain)
		network.StartApiServer(6969, nodeID, chain)
	} else {
		// Run CLI mode by default
		cli := cli.CommandLine{}
		cli.Run()
	}
}
