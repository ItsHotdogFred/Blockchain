package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ItsHotdogFred/blockchain/wallet"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	Handle(err)
	return transaction
}

func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 24)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randData)
	}


	txin := TxInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(100, to)

	tx := Transaction{nil, []TxInput{txin}, []TxOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx

}

func NewTransaction(w *wallet.Wallet, to string, amount int, UTXO *UTXOSet) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)
	acc, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	from := string(w.Address())

	outputs = append(outputs, *NewTXOutput(amount, to))

	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	UTXO.Blockchain.SignTransaction(&tx, *w.PrivateKey.ToECDSA())

	return &tx
}

type GameResult struct {
	Transaction  *Transaction
	Won          bool
	Amount       int
	Change       int
	GameType     string
	ServerNumber int  // For Number Range game
}

func NewGameTransaction(w *wallet.Wallet, amount int, utxoSet *UTXOSet, gameType string, calculateWinnings func(int) (bool, int)) *GameResult {
	var inputs []TxInput
	var outputs []TxOutput

	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)
	acc, validOutputs := utxoSet.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("Error: not enough funds for game")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	from := string(w.Address())

	// Use the provided function to calculate game result
	won, winnings := calculateWinnings(amount)

	if winnings > 0 {
		outputs = append(outputs, *NewTXOutput(winnings, from))
	}

	// Add change if there was any excess input
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	utxoSet.Blockchain.SignTransaction(&tx, *w.PrivateKey.ToECDSA())

	return &GameResult{
		Transaction: &tx,
		Won:         won,
		Amount:      amount,
		Change:      acc - amount,
		GameType:    gameType,
	}
}

type CoinflipResult struct {
	Transaction *Transaction
	Won         bool
	Amount      int
	Change      int
}

func NewCoinflipTransaction(w *wallet.Wallet, amount int, UTXO *UTXOSet) *GameResult {
	// Coinflip-specific logic: 50/50 chance, double or nothing
	coinflipLogic := func(betAmount int) (bool, int) {
		// Generate random coinflip result
		randomBytes := make([]byte, 1)
		_, err := rand.Read(randomBytes)
		if err != nil {
			log.Panic(err)
		}
		coinflipResult := int(randomBytes[0]) % 2 // 0 or 1

		if coinflipResult == 1 {
			// Win: get bet amount back + winnings (net gain = bet amount)
			winnings := betAmount + betAmount  // Original bet + winnings
			fmt.Printf("Coinflip WIN! You gained %d coins\n", betAmount)
			return true, winnings
		} else {
			// Lose: no winnings
			fmt.Printf("Coinflip LOSS! You lost %d coins\n", betAmount)
			return false, 0
		}
	}

	return NewGameTransaction(w, amount, UTXO, "coinflip", coinflipLogic)
}

func NewDiceRollTransaction(w *wallet.Wallet, amount int, UTXO *UTXOSet) *GameResult {
	// Dice roll logic: 33% chance to win 3x the bet
	diceLogic := func(betAmount int) (bool, int) {
		// Generate random number 1-6
		randomBytes := make([]byte, 1)
		_, err := rand.Read(randomBytes)
		if err != nil {
			log.Panic(err)
		}
		diceRoll := (int(randomBytes[0]) % 6) + 1 // 1-6

		if diceRoll == 6 {
			// Win: 3x the bet amount
			winnings := betAmount * 3
			fmt.Printf("Dice Roll WIN! Rolled a 6! You won %d coins\n", winnings)
			return true, winnings
		} else {
			// Lose: no winnings
			fmt.Printf("Dice Roll LOSS! Rolled a %d. You lost %d coins\n", diceRoll, betAmount)
			return false, 0
		}
	}

	return NewGameTransaction(w, amount, UTXO, "dice", diceLogic)
}

func NewNumberRangeTransaction(w *wallet.Wallet, amount int, userGuess int, UTXO *UTXOSet) *GameResult {
	var inputs []TxInput
	var outputs []TxOutput

	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)
	acc, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("Error: not enough funds for game")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	from := string(w.Address())

	// Generate random server number (1-100)
	randomBytes := make([]byte, 1)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Panic(err)
	}
	serverNumber := (int(randomBytes[0]) % 100) + 1 // 1-100

	// Check if user's guess is within ±5 range
	lowerBound := userGuess - 5
	upperBound := userGuess + 5

	// Handle boundary conditions
	if lowerBound < 1 {
		lowerBound = 1
	}
	if upperBound > 100 {
		upperBound = 100
	}

	var won bool
	var winnings int

	if serverNumber >= lowerBound && serverNumber <= upperBound {
		// Win: 5x the bet amount
		winnings = amount * 5
		fmt.Printf("Number Range WIN! Server: %d, Your guess: %d (range %d-%d). You won %d coins\n", serverNumber, userGuess, lowerBound, upperBound, winnings)
		won = true
	} else {
		// Lose: no winnings
		fmt.Printf("Number Range LOSS! Server: %d, Your guess: %d (range %d-%d). You lost %d coins\n", serverNumber, userGuess, lowerBound, upperBound, amount)
		won = false
	}

	if winnings > 0 {
		outputs = append(outputs, *NewTXOutput(winnings, from))
	}

	// Add change if there was any excess input
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	UTXO.Blockchain.SignTransaction(&tx, *w.PrivateKey.ToECDSA())

	return &GameResult{
		Transaction:  &tx,
		Won:          won,
		Amount:       amount,
		Change:       acc - amount,
		GameType:     "numberrange",
		ServerNumber: serverNumber,
	}
}

func NewCoinflipTransactionLegacy(w *wallet.Wallet, amount int, UTXO *UTXOSet) *Transaction {
	result := NewCoinflipTransaction(w, amount, UTXO)
	return result.Transaction
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTxs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, in := range tx.Inputs {
		if prevTxs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inId, in := range txCopy.Inputs {
		prevTX := prevTxs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTX.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		Handle(err)
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[inId].Signature = signature
	}
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{in.ID, in.Out, nil, nil})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{out.Value, out.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("Previous transaction does not exist")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inId, in := range tx.Inputs {
		prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false
		}
	}

	return true
}

func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:     %x", input.ID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
