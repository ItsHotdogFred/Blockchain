# Blockchain Casino

A fully functional blockchain implementation with built-in gambling games, featuring decentralized transactions, proof-of-work consensus, and a web-based casino interface.

![Blockchain Casino](https://img.shields.io/badge/Go-1.25.1-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)

## 🎰 Features

### Core Blockchain
- **Proof-of-Work Consensus**: Secure mining with adjustable difficulty
- **Cryptographic Wallets**: ECDSA-based key management
- **UTXO Model**: Efficient transaction processing
- **Merkle Tree Integrity**: Tamper-proof transaction verification
- **P2P Network**: Decentralized node communication
- **RESTful API**: HTTP endpoints for blockchain operations

### Gambling Games
- **Coin Flip**: 50/50 chance to double your coins
- **Dice Roll**: 33% chance to win 3x your bet
- **Number Range**: Guess within ±5 range to win 5x your bet

### Web Interface
- **Real-time Balance Tracking**: Live wallet balance updates
- **Interactive Wallet Management**: Create and manage multiple wallets
- **Transaction History**: View all blockchain transactions
- **Game Dashboard**: Integrated gambling interface
- **Responsive Design**: Works on desktop and mobile

## 🏗️ Architecture

```
├── blockchain/           # Core blockchain implementation
│   ├── block.go         # Block structure and creation
│   ├── blockchain.go    # Blockchain management
│   ├── merkle.go        # Merkle tree implementation
│   ├── proof.go         # Proof-of-work algorithm
│   ├── transaction.go   # Transaction handling
│   ├── tx.go           # Transaction input/output
│   └── utxo.go         # UTXO set management
├── cli/                 # Command-line interface
├── network/             # P2P networking
├── wallet/              # Cryptographic wallet management
└── website/             # Web-based casino interface
```

## 🚀 Quick Start

### Prerequisites
- Go 1.25.1 or later
- Modern web browser

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/ItsHotdogFred/blockchain.git
   cd blockchain
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Build the application**
   ```bash
   go build -o main
   ```

### Running the Blockchain

#### Option 1: Server Mode (Recommended for Web Interface)
```bash
# Set node ID and start server
export NODE_ID="3000"
./main server
```

#### Option 2: CLI Mode
```bash
# Create a new wallet
./main createwallet

# Get balance
./main getbalance -address YOUR_WALLET_ADDRESS

# Send transactions
./main send -from FROM_ADDRESS -to TO_ADDRESS -amount 100

# Start mining node
./main startnode -miner YOUR_WALLET_ADDRESS
```

#### Option 3: Gambling Games
```bash
# Coin flip (50/50 chance to double)
./main coinflip -from YOUR_ADDRESS -amount 100

# Dice roll (33% chance to win 3x)
./main diceroll -from YOUR_ADDRESS -amount 100

# Number range (guess number 1-100, win 5x if ±5)
./main numberrange -from YOUR_ADDRESS -amount 100 -guess 50
```

### Web Interface

1. **Start the blockchain server** (see above)
2. **Open the web interface**
   ```bash
   cd website
   # Open index.html in your browser
   ```
3. **Access the casino** at `http://localhost:6969`

## 🎮 Web Interface Features

### Wallet Management
- **Create Wallet**: Generate new cryptographic wallets
- **View Balance**: Real-time balance updates
- **Address Display**: Copy wallet addresses with one click

### Gambling Games
- **Coin Flip**: Simple 50/50 game
- **Dice Roll**: Higher risk, higher reward
- **Number Range**: Skill-based guessing game

### Transaction Management
- **Send Transactions**: Transfer coins between wallets
- **View History**: Complete transaction ledger
- **Network Status**: Real-time blockchain information

## 🔧 API Endpoints

The blockchain exposes the following HTTP endpoints:

- `GET /balance?address=ADDRESS` - Get wallet balance
- `POST /createwallet` - Create new wallet
- `POST /send` - Send transaction
- `GET /chain` - Get full blockchain
- `GET /transactions` - Get transaction pool
- `POST /coinflip` - Play coin flip game
- `POST /diceroll` - Play dice roll game
- `POST /numberrange` - Play number range game

## ⛏️ Mining

The blockchain uses a proof-of-work consensus mechanism:

- **Difficulty Adjustment**: Automatically adjusts based on network hash rate
- **Block Rewards**: Miners receive rewards for validating transactions
- **Transaction Fees**: Small fees for processing transactions
- **Merkle Root**: Efficient transaction verification

### Starting a Mining Node
```bash
export NODE_ID="3000"
./main startnode -miner YOUR_WALLET_ADDRESS
```

## 🌐 Network Configuration

### Node Communication
- **Protocol**: TCP
- **Default Port**: 3000 (P2P), 6969 (API)
- **Discovery**: Nodes can discover each other automatically
- **Syncing**: Automatic blockchain synchronization

### Multi-Node Setup
```bash
# Terminal 1 - Node 3000
export NODE_ID="3000"
./main startnode -miner MINER_ADDRESS

# Terminal 2 - Node 3001
export NODE_ID="3001"
./main startnode

# Terminal 3 - Node 3002
export NODE_ID="3002"
./main startnode
```

## 📊 CLI Commands

### Wallet Operations
```bash
createwallet              # Create new wallet
listaddresses           # List all wallet addresses
getbalance -address ADDR # Get wallet balance
```

### Blockchain Operations
```bash
printchain              # Display all blocks
reindexutxo            # Rebuild UTXO set
```

### Transaction Operations
```bash
send -from FROM -to TO -amount AMOUNT  # Send coins
```

### Network Operations
```bash
startnode -miner ADDRESS  # Start mining node
```

### Gambling Games
```bash
coinflip -from FROM -amount AMOUNT
diceroll -from FROM -amount AMOUNT
numberrange -from FROM -amount AMOUNT -guess NUMBER
```

## 🔒 Security Features

### Cryptography
- **ECDSA**: Elliptic Curve Digital Signature Algorithm
- **SHA-256**: Secure hashing for blocks and transactions
- **Merkle Trees**: Tamper-proof transaction verification
- **Digital Signatures**: Transaction authentication

### Consensus
- **Proof-of-Work**: Prevents double-spending and ensures consensus
- **Longest Chain Rule**: Resolves conflicts automatically
- **Difficulty Adjustment**: Maintains consistent block times

## 🛠️ Development

### Project Structure
- **Modular Design**: Clean separation of concerns
- **BadgerDB**: High-performance key-value database
- **Gorilla Mux**: HTTP routing for API endpoints
- **Base58**: Address encoding

### Adding New Features
1. **Core Logic**: Add blockchain functions in `/blockchain/`
2. **CLI Commands**: Extend `/cli/cli.go`
3. **API Endpoints**: Modify `/network/network.go`
4. **Web Interface**: Update `/website/` files

## 🐛 Troubleshooting

### Common Issues

**Database Errors**
```bash
# Clear blockchain data
rm -rf tmp/
```

**Network Issues**
```bash
# Check if ports are available
netstat -tulpn | grep :3000
netstat -tulpn | grep :6969
```

**Web Interface Issues**
- Ensure the blockchain server is running
- Check browser console for errors
- Refresh the page if connection issues occur

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Blockchain Tutorial**: Based on the excellent YouTube series by [Laurence Bradford](https://www.youtube.com/watch?v=mYlHT9bB6OE&list=PLJbE2Yu2zumC5QE39TQHBLYJDB2gfFE5Q)
- **Go Libraries**: Built with amazing open-source libraries
- **Bulma CSS**: Beautiful UI framework for the web interface

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📞 Contact

Project Link: [https://github.com/ItsHotdogFred/blockchain](https://github.com/ItsHotdogFred/blockchain)

---

⚠️ **Disclaimer**: This project is for educational purposes only. The gambling features are simulated and should not be used for real gambling. Please gamble responsibly.