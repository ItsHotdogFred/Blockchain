package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/ItsHotdogFred/blockchain/blockchain"
	"github.com/ItsHotdogFred/blockchain/wallet"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address ADDRESS - get the balance for")
	fmt.Println(" createblockchain -address ADDRESS creates a blockchain")
	fmt.Println(" printchain - Prints the blocks in the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT - Send amount")
	fmt.Println(" createwallet - Creates a new Wallet")
	fmt.Println(" listaddresses - Lists the addresses in our wallet file")
	fmt.Println(" reindexutxo - Rebuilds the UTXO set")
}

func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) reindexUTXO() {
	chain := blockchain.ContinueBlockchain("")
	defer chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set. \n", count)
} 

func (cli *CommandLine) listAddresses() {
	wallets, _ := wallet.CreateWallets()
	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) createWallet() {
	wallets, _ := wallet.CreateWallets()
	address := wallets.AddWallet()
	wallets.SaveFile()

	fmt.Printf("New address is: %s\n", address)
}

func (cli *CommandLine) printChain() {
	chain := blockchain.ContinueBlockchain("")
	defer chain.Database.Close()
	iter := chain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)

		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}

}

func (cli *CommandLine) createBlockchain(address string) {
	if !wallet.VaildateAddress(address) {
		log.Panic("Address is not Valid")
	}
	chain := blockchain.InitBlockchain(address)
	defer chain.Database.Close()
	fmt.Println("Finished!")

	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()
}

func (cli *CommandLine) getBalance(address string) {
	if !wallet.VaildateAddress(address) {
		log.Panic("Address is not Valid")
	}

	chain := blockchain.ContinueBlockchain(address)
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUnspentTransactions(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int) {
	if !wallet.VaildateAddress(from) {
		log.Panic("Address is not Valid")
	}
	if !wallet.VaildateAddress(to) {
		log.Panic("Address is not Valid")
	}

	chain := blockchain.ContinueBlockchain(from)
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	tx := blockchain.NewTransaction(from, to, amount, &UTXOSet)
	cbTx := blockchain.CoinbaseTx(from, "")
	block := chain.AddBlock([]*blockchain.Transaction{cbTx, tx})
	UTXOSet.Update(block)
	fmt.Println("Success!")
}

func (cli *CommandLine) Run() {
	cli.ValidateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	_ = printChainCmd
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "Create blockchain with address")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}
	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}
	if reindexUTXOCmd.Parsed() {
		cli.reindexUTXO()
	} 
}
