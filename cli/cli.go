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
	"github.com/ItsHotdogFred/blockchain/network"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address ADDRESS - get the balance for")
	fmt.Println(" printchain - Prints the blocks in the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT - Send amount of coins")
	fmt.Println(" createwallet - Creates a new Wallet")
	fmt.Println(" listaddresses - Lists the addresses in our wallet file")
	fmt.Println(" reindexutxo - Rebuilds the UTXO set")
	fmt.Println(" startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env.")
	fmt.Println(" coinflip -from FROM -amount AMOUNT - Coinflip to double or lose your coins")
	fmt.Println(" diceroll -from FROM -amount AMOUNT - Dice roll (33% chance to win 3x)")
	fmt.Println(" numberrange -from FROM -amount AMOUNT -guess NUMBER - Number Range (±5 range wins 5x)")
}

func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) StartNode(nodeID string) {
	fmt.Printf("Starting Node %s\n", nodeID)
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	network.StartServer(nodeID, chain)
}

func (cli *CommandLine) reindexUTXO(nodeID string) {
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{chain}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set. \n", count)
} 

func (cli *CommandLine) listAddresses(nodeID string) {
	wallets, _ := wallet.CreateWallets(nodeID)
	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) createWallet(nodeID string) {
	wallets, _ := wallet.CreateWallets(nodeID)
	address := wallets.AddWallet()
	wallets.SaveFile(nodeID)

	// Create initial balance transaction (100 coins)
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()

	UTXOSet := blockchain.UTXOSet{Blockchain: chain}

	// Create coinbase transaction for initial balance
	cbTx := blockchain.CoinbaseTx(address, "Initial balance")
	txs := []*blockchain.Transaction{cbTx}
	block := chain.MineBlock(txs)
	UTXOSet.Update(block)

	fmt.Printf("New address is: %s with 100 initial balance\n", address)
}

func (cli *CommandLine) printChain(nodeID string) {
	chain := blockchain.ContinueBlockChain(nodeID)
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


func (cli *CommandLine) getBalance(address string, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not Valid")
	}

	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOSet := blockchain.UTXOSet{chain}
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

func (cli *CommandLine) send(from, to string, amount int, nodeID string) {
	if !wallet.ValidateAddress(from) {
		log.Panic("Address is not Valid")
	}
	if !wallet.ValidateAddress(to) {
		log.Panic("Address is not Valid")
	}

	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOSet := blockchain.UTXOSet{chain}
	defer chain.Database.Close()

	wallets, err := wallet.CreateWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := blockchain.NewTransaction(&wallet, to, amount, &UTXOSet)
	cbTx := blockchain.CoinbaseTx(from, "")
	txs := []*blockchain.Transaction{cbTx, tx}
	block := chain.MineBlock(txs)
	UTXOSet.Update(block)

	fmt.Println("Success!")
}

func (cli *CommandLine) coinflip(from string, amount int, nodeID string) {
	if !wallet.ValidateAddress(from) {
		log.Panic("Address is not Valid")
	}

	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOSet := blockchain.UTXOSet{chain}
	defer chain.Database.Close()

	wallets, err := wallet.CreateWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	result := blockchain.NewCoinflipTransaction(&wallet, amount, &UTXOSet)
	txs := []*blockchain.Transaction{result.Transaction}
	block := chain.MineBlock(txs)
	UTXOSet.Update(block)

	fmt.Println("Coinflip transaction completed!")
}

func (cli *CommandLine) diceroll(from string, amount int, nodeID string) {
	if !wallet.ValidateAddress(from) {
		log.Panic("Address is not Valid")
	}

	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOSet := blockchain.UTXOSet{chain}
	defer chain.Database.Close()

	wallets, err := wallet.CreateWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	result := blockchain.NewDiceRollTransaction(&wallet, amount, &UTXOSet)
	txs := []*blockchain.Transaction{result.Transaction}
	block := chain.MineBlock(txs)
	UTXOSet.Update(block)

	fmt.Println("Dice roll transaction completed!")
}

func (cli *CommandLine) numberrange(from string, amount int, guess int, nodeID string) {
	if !wallet.ValidateAddress(from) {
		log.Panic("Address is not Valid")
	}

	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOSet := blockchain.UTXOSet{chain}
	defer chain.Database.Close()

	wallets, err := wallet.CreateWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	result := blockchain.NewNumberRangeTransaction(&wallet, amount, guess, &UTXOSet)
	txs := []*blockchain.Transaction{result.Transaction}
	block := chain.MineBlock(txs)
	UTXOSet.Update(block)

	fmt.Println("Number Range transaction completed!")
}

func (cli *CommandLine) Run() {
	cli.ValidateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID env is not set!")
		runtime.Goexit()
	}

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	coinflipCmd := flag.NewFlagSet("coinflip", flag.ExitOnError)
	diceRollCmd := flag.NewFlagSet("diceroll", flag.ExitOnError)
	numberRangeCmd := flag.NewFlagSet("numberrange", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	_ = printChainCmd
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	coinflipFrom := coinflipCmd.String("from", "", "Source wallet address")
	coinflipAmount := coinflipCmd.Int("amount", 0, "Amount to bet")
	diceRollFrom := diceRollCmd.String("from", "", "Source wallet address")
	diceRollAmount := diceRollCmd.Int("amount", 0, "Amount to bet")
	numberRangeFrom := numberRangeCmd.String("from", "", "Source wallet address")
	numberRangeAmount := numberRangeCmd.Int("amount", 0, "Amount to bet")
	numberRangeGuess := numberRangeCmd.Int("guess", 0, "Number guess (1-100)")
	_ = startNodeCmd.String("miner", "", "Mining address (deprecated - transactions are self-mined)")

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
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
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
	case "coinflip":
		err := coinflipCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "diceroll":
		err := diceRollCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "numberrange":
		err := numberRangeCmd.Parse(os.Args[2:])
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
		cli.getBalance(*getBalanceAddress, nodeID)
	}


	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmount, nodeID)
	}

	if coinflipCmd.Parsed() {
		if *coinflipFrom == "" || *coinflipAmount <= 0 {
			coinflipCmd.Usage()
			runtime.Goexit()
		}
		cli.coinflip(*coinflipFrom, *coinflipAmount, nodeID)
	}

	if diceRollCmd.Parsed() {
		if *diceRollFrom == "" || *diceRollAmount <= 0 {
			diceRollCmd.Usage()
			runtime.Goexit()
		}
		cli.diceroll(*diceRollFrom, *diceRollAmount, nodeID)
	}

	if numberRangeCmd.Parsed() {
		if *numberRangeFrom == "" || *numberRangeAmount <= 0 || *numberRangeGuess < 1 || *numberRangeGuess > 100 {
			numberRangeCmd.Usage()
			runtime.Goexit()
		}
		cli.numberrange(*numberRangeFrom, *numberRangeAmount, *numberRangeGuess, nodeID)
	}

	if startNodeCmd.Parsed() {
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			startNodeCmd.Usage()
			runtime.Goexit()
		}
		cli.StartNode(nodeID)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}
	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeID)
	}
	if reindexUTXOCmd.Parsed() {
		cli.reindexUTXO(nodeID)
	} 
}
