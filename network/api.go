package network

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/ItsHotdogFred/blockchain/blockchain"
	"github.com/ItsHotdogFred/blockchain/wallet"
)

type helloworld struct {
	Text         string `json:"text"`
	Time         string `json:"time"`
	RandomNumber int    `json:"number"`
}

type TransactionRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

type CoinflipRequest struct {
	From   string `json:"from"`
	Amount int    `json:"amount"`
}

type DiceRollRequest struct {
	From   string `json:"from"`
	Amount int    `json:"amount"`
}

type NumberRangeRequest struct {
	From   string `json:"from"`
	Amount int    `json:"amount"`
	Guess  int    `json:"guess"`
}

var nodeID string

// Rate limiting structures
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
}

var rateLimiter = &RateLimiter{
	requests: make(map[string][]time.Time),
}

func (rl *RateLimiter) AllowRequest(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Remove requests older than 1 minute
	if requests, exists := rl.requests[clientIP]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < time.Minute {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[clientIP] = validRequests

		// Check if client has exceeded limit
		if len(validRequests) >= 30 {
			return false
		}
	}

	// Add current request
	rl.requests[clientIP] = append(rl.requests[clientIP], now)
	return true
}

type GameRequest struct {
	From   string `json:"from"`
	Amount int    `json:"amount"`
}

type GameHandler func(*wallet.Wallet, int, *blockchain.UTXOSet) *blockchain.GameResult

func handleGameTransaction(w http.ResponseWriter, r *http.Request, chain *blockchain.BlockChain, req GameRequest, handler GameHandler, gameName string) {
	// Validate address with error handling
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, fmt.Sprintf("Invalid address: %v", rec), http.StatusBadRequest)
				return
			}
		}()
		if !wallet.ValidateAddress(req.From) {
			http.Error(w, "Invalid address", http.StatusBadRequest)
			return
		}
	}()
	if req.Amount <= 0 {
		http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
		return
	}

	UTXOSet := blockchain.UTXOSet{Blockchain: chain}

	// Load wallets and get sender wallet
	wallets, err := wallet.CreateWallets(nodeID)
	if err != nil {
		http.Error(w, "Failed to load wallets", http.StatusInternalServerError)
		return
	}
	// Check if wallet exists
	if _, exists := wallets.Wallets[req.From]; !exists {
		http.Error(w, "Wallet not found", http.StatusNotFound)
		return
	}
	senderWallet := wallets.GetWallet(req.From)

	// Create game transaction with error handling
	var gameResult *blockchain.GameResult
	var txPanicked bool

	func() {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, fmt.Sprintf("Failed to create %s transaction: %v", gameName, rec), http.StatusBadRequest)
				txPanicked = true
			}
		}()
		gameResult = handler(&senderWallet, req.Amount, &UTXOSet)
	}()

	if txPanicked || gameResult == nil {
		return
	}

	// Mine block with error handling - no coinbase transaction for games (same as CLI)
	txs := []*blockchain.Transaction{gameResult.Transaction}
	var block *blockchain.Block
	var panicked bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				http.Error(w, fmt.Sprintf("Failed to mine %s transaction: %v", gameName, r), http.StatusInternalServerError)
				panicked = true
			}
		}()
		block = chain.MineBlock(txs)
	}()

	if panicked || block == nil {
		return
	}

	UTXOSet.Update(block)

	// Calculate amount change and winnings
	var amountChange int
	var resultStr string
	var message string
	if gameResult.Won {
		resultStr = "WIN"
		// Calculate total coins received (includes original bet)
		var totalReceived int
		// Look at the transaction outputs to see what the player actually received
		for _, output := range gameResult.Transaction.Outputs {
			// Find the output that goes back to the sender (assuming it's the winnings)
			if output.Value > req.Amount { // This indicates winnings (more than the bet)
				totalReceived = output.Value
				break
			}
		}
		if totalReceived == 0 {
			// Default calculation if we can't find specific output
			if gameResult.GameType == "coinflip" {
				totalReceived = req.Amount * 2 // Double the bet
			} else if gameResult.GameType == "dice" {
				totalReceived = req.Amount * 3 // Triple the bet
			} else if gameResult.GameType == "numberrange" {
				totalReceived = req.Amount * 5 // 5x the bet
			}
		}
		amountChange = totalReceived - req.Amount // Net gain
		message = fmt.Sprintf("%s %s! You received %d coins (net gain: %d)", gameName, resultStr, totalReceived, amountChange)
	} else {
		resultStr = "LOSS"
		amountChange = -req.Amount
		message = fmt.Sprintf("%s %s! You lost %d coins", gameName, resultStr, req.Amount)
	}

	// Return success response
	response := map[string]interface{}{
		"status":       "success",
		"result":       resultStr,
		"amountChange": amountChange,
		"betAmount":    req.Amount,
		"change":       gameResult.Change,
		"message":      message,
		"block":        fmt.Sprintf("%x", block.Hash),
		"tx":           fmt.Sprintf("%x", gameResult.Transaction.ID),
	}

	// Add server number for Number Range game
	if gameResult.GameType == "numberrange" && gameResult.ServerNumber != 0 {
		response["serverNumber"] = gameResult.ServerNumber
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for balance endpoint and OPTIONS requests
		if r.URL.Path == "/balance" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Get client IP (handle reverse proxy by checking X-Forwarded-For first)
		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = r.Header.Get("X-Real-IP")
		}
		if clientIP == "" {
			clientIP = r.RemoteAddr
		}

		// Check if request is allowed
		if !rateLimiter.AllowRequest(clientIP) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error":       "Rate limit exceeded. Maximum 20 requests per minute allowed.",
				"retry_after": "60",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func StartApiServer(port int, ID string, chain *blockchain.BlockChain) {
	nodeID = ID

	router := mux.NewRouter()
	portStr := fmt.Sprintf(":%s", strconv.Itoa(port))

	// Add middlewares
	router.Use(corsMiddleware)
	router.Use(rateLimitMiddleware)

	router.HandleFunc("/hello", HelloWorld).Methods("GET", "OPTIONS")
	router.HandleFunc("/createwallet", func(w http.ResponseWriter, r *http.Request) {
		APICreateWallet(w, r, chain)
	}).Methods("POST", "OPTIONS")
	router.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		SendTransaction(w, r, chain)
	}).Methods("POST", "OPTIONS")
	router.HandleFunc("/balance", func(w http.ResponseWriter, r *http.Request) {
		GetBalance(w, r, chain)
	}).Methods("GET", "OPTIONS")
	router.HandleFunc("/coinflip", func(w http.ResponseWriter, r *http.Request) {
		CoinFlip(w, r, chain)
	}).Methods("POST", "OPTIONS")
	router.HandleFunc("/diceroll", func(w http.ResponseWriter, r *http.Request) {
		DiceRoll(w, r, chain)
	}).Methods("POST", "OPTIONS")
	router.HandleFunc("/numberrange", func(w http.ResponseWriter, r *http.Request) {
		NumberRange(w, r, chain)
	}).Methods("POST", "OPTIONS")
	router.HandleFunc("/blockchain", func(w http.ResponseWriter, r *http.Request) {
		GetBlockchain(w, r, chain)
	}).Methods("GET", "OPTIONS")

	http.ListenAndServe(portStr, router)
}

func HelloWorld(w http.ResponseWriter, r *http.Request) {

	jsonResponse := helloworld{
		Text:         "Hello World",
		Time:         time.Now().String(),
		RandomNumber: rand.Int(),
	}
	jsonResponseByte, _ := json.Marshal(jsonResponse)

	w.Header().Set("Content-Type", "application/json")

	w.Write(jsonResponseByte)
}

func APICreateWallet(w http.ResponseWriter, r *http.Request, chain *blockchain.BlockChain) {
	var wallets *wallet.Wallets
	var err error

	// Try to load existing wallets, create empty if doesn't exist
	wallets, err = wallet.CreateWallets(nodeID)
	if err != nil {
		wallets = &wallet.Wallets{}
		wallets.Wallets = make(map[string]*wallet.Wallet)
	}

	address := wallets.AddWallet()
	wallets.SaveFile(nodeID)

	// Create initial balance transaction (100 coins) with error handling
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}

	// Create coinbase transaction for initial balance
	cbTx := blockchain.CoinbaseTx(address, "Initial balance")
	txs := []*blockchain.Transaction{cbTx}

	var block *blockchain.Block
	func() {
		defer func() {
			if r := recover(); r != nil {
				http.Error(w, fmt.Sprintf("Failed to mine initial balance: %v", r), http.StatusInternalServerError)
				return
			}
		}()
		block = chain.MineBlock(txs)
	}()

	if block == nil {
		http.Error(w, "Failed to mine initial balance block", http.StatusInternalServerError)
		return
	}

	UTXOSet.Update(block)

	fmt.Printf("New address is: %s\n", address)

	response := map[string]string{
		"address": address,
		"message": "Wallet created with 100 initial balance",
	}

	jsonResponseByte, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponseByte)
}

func SendTransaction(w http.ResponseWriter, r *http.Request, chain *blockchain.BlockChain) {
	var txReq TransactionRequest

	// Decode JSON request body
	err := json.NewDecoder(r.Body).Decode(&txReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate addresses with error handling
	func() {
		defer func() {
			if r := recover(); r != nil {
				http.Error(w, fmt.Sprintf("Invalid 'from' address: %v", r), http.StatusBadRequest)
				return
			}
		}()
		if !wallet.ValidateAddress(txReq.From) {
			http.Error(w, "Invalid 'from' address", http.StatusBadRequest)
			return
		}
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				http.Error(w, fmt.Sprintf("Invalid 'to' address: %v", r), http.StatusBadRequest)
				return
			}
		}()
		if !wallet.ValidateAddress(txReq.To) {
			http.Error(w, "Invalid 'to' address", http.StatusBadRequest)
			return
		}
	}()
	if txReq.Amount <= 0 {
		http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
		return
	}

	UTXOSet := blockchain.UTXOSet{Blockchain: chain}

	// Load wallets and get sender wallet
	wallets, err := wallet.CreateWallets(nodeID)
	if err != nil {
		http.Error(w, "Failed to load wallets", http.StatusInternalServerError)
		return
	}
	// Check if wallet exists
	if _, exists := wallets.Wallets[txReq.From]; !exists {
		http.Error(w, "Sender wallet not found", http.StatusNotFound)
		return
	}
	senderWallet := wallets.GetWallet(txReq.From)

	// Create transaction with error handling
	var tx *blockchain.Transaction
	var txPanicked bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				http.Error(w, fmt.Sprintf("Failed to create transaction: %v", r), http.StatusBadRequest)
				txPanicked = true
			}
		}()
		tx = blockchain.NewTransaction(&senderWallet, txReq.To, txReq.Amount, &UTXOSet)
	}()

	if txPanicked || tx == nil {
		return
	}

	// Mine block with error handling
	txs := []*blockchain.Transaction{tx}
	var block *blockchain.Block
	var panicked bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				http.Error(w, fmt.Sprintf("Failed to mine transaction: %v", r), http.StatusInternalServerError)
				panicked = true
			}
		}()
		block = chain.MineBlock(txs)
	}()

	if panicked || block == nil {
		return
	}

	UTXOSet.Update(block)

	// Return success response
	response := map[string]string{
		"status":  "success",
		"message": "Transaction mined successfully",
		"block":   fmt.Sprintf("%x", block.Hash),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetBalance(w http.ResponseWriter, r *http.Request, chain *blockchain.BlockChain) {
	// Get query parameters from URL
	query := r.URL.Query()
	address := query.Get("address")

	if address == "" {
		http.Error(w, "Address parameter is required", http.StatusBadRequest)
		return
	}

	// Validate address with error handling
	validAddress := true
	func() {
		defer func() {
			if r := recover(); r != nil {
				http.Error(w, fmt.Sprintf("Invalid address: %v", r), http.StatusBadRequest)
				validAddress = false
				return
			}
		}()
		validAddress = wallet.ValidateAddress(address)
	}()

	if !validAddress {
		return
	}

	// Get balance with error handling
	var balance int
	var UTXOs []blockchain.TxOutput
	func() {
		defer func() {
			if r := recover(); r != nil {
				http.Error(w, fmt.Sprintf("Failed to get balance: %v", r), http.StatusInternalServerError)
				return
			}
		}()

		UTXOSet := blockchain.UTXOSet{Blockchain: chain}

		pubKeyHash := wallet.Base58Decode([]byte(address))
		if len(pubKeyHash) < 5 {
			http.Error(w, "Invalid address format", http.StatusBadRequest)
			return
		}
		pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
		UTXOs = UTXOSet.FindUnspentTransactions(pubKeyHash)

		balance = 0
		for _, out := range UTXOs {
			balance += out.Value
		}
	}()

	response := map[string]interface{}{
		"address": address,
		"balance": balance,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func DiceRoll(w http.ResponseWriter, r *http.Request, chain *blockchain.BlockChain) {
	var drReq DiceRollRequest

	// Decode JSON request body
	err := json.NewDecoder(r.Body).Decode(&drReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Use common game handler with dice roll function
	handleGameTransaction(w, r, chain, GameRequest{
		From:   drReq.From,
		Amount: drReq.Amount,
	}, func(w *wallet.Wallet, amount int, utxo *blockchain.UTXOSet) *blockchain.GameResult {
		return blockchain.NewDiceRollTransaction(w, amount, utxo)
	}, "Dice Roll")
}

func CoinFlip(w http.ResponseWriter, r *http.Request, chain *blockchain.BlockChain) {
	var cfReq CoinflipRequest

	// Decode JSON request body
	err := json.NewDecoder(r.Body).Decode(&cfReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Use common game handler with coinflip function
	handleGameTransaction(w, r, chain, GameRequest{
		From:   cfReq.From,
		Amount: cfReq.Amount,
	}, func(w *wallet.Wallet, amount int, utxo *blockchain.UTXOSet) *blockchain.GameResult {
		return blockchain.NewCoinflipTransaction(w, amount, utxo)
	}, "Coinflip")
}

func NumberRange(w http.ResponseWriter, r *http.Request, chain *blockchain.BlockChain) {
	var nrReq NumberRangeRequest

	// Decode JSON request body
	err := json.NewDecoder(r.Body).Decode(&nrReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate guess parameter
	if nrReq.Guess < 1 || nrReq.Guess > 100 {
		http.Error(w, "Guess must be between 1 and 100", http.StatusBadRequest)
		return
	}

	// Use common game handler with number range function
	handleGameTransaction(w, r, chain, GameRequest{
		From:   nrReq.From,
		Amount: nrReq.Amount,
	}, func(w *wallet.Wallet, amount int, utxo *blockchain.UTXOSet) *blockchain.GameResult {
		return blockchain.NewNumberRangeTransaction(w, amount, nrReq.Guess, utxo)
	}, "Number Range")
}

type BlockInfo struct {
	Height       int               `json:"height"`
	Hash         string            `json:"hash"`
	PrevHash     string            `json:"prevHash"`
	Timestamp    int64             `json:"timestamp"`
	Nonce        int               `json:"nonce"`
	Transactions []TransactionInfo `json:"transactions"`
}

type TransactionInfo struct {
	ID      string `json:"id"`
	Inputs  int    `json:"inputs"`
	Outputs int    `json:"outputs"`
}

type BlockchainResponse struct {
	Blocks []BlockInfo `json:"blocks"`
	Total  int         `json:"totalBlocks"`
}

func GetBlockchain(w http.ResponseWriter, r *http.Request, chain *blockchain.BlockChain) {
	// Get limit parameter from query (default 20, max 100)
	query := r.URL.Query()
	limitStr := query.Get("limit")
	limit := 20 // default

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	var blocks []BlockInfo
	count := 0

	// Iterate through blockchain from newest to oldest
	iter := chain.Iterator()
	for count < limit {
		block := iter.Next()
		if block == nil {
			break
		}

		// Convert transactions to summary info
		var txInfos []TransactionInfo
		for _, tx := range block.Transactions {
			txInfos = append(txInfos, TransactionInfo{
				ID:      fmt.Sprintf("%x", tx.ID),
				Inputs:  len(tx.Inputs),
				Outputs: len(tx.Outputs),
			})
		}

		// Create block info
		blockInfo := BlockInfo{
			Height:       block.Height,
			Hash:         fmt.Sprintf("%x", block.Hash),
			PrevHash:     fmt.Sprintf("%x", block.PrevHash),
			Timestamp:    block.Timestamp,
			Nonce:        block.Nonce,
			Transactions: txInfos,
		}

		blocks = append(blocks, blockInfo)
		count++

		// Stop if we've reached genesis block
		if len(block.PrevHash) == 0 {
			break
		}
	}

	response := BlockchainResponse{
		Blocks: blocks,
		Total:  count,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
