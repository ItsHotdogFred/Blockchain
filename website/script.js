
async function fetchBlockchainData() {
    try {
        const response = await fetch('http://localhost:6969/blockchain?limit=50');
        if (!response.ok) {
            throw new Error('Failed to fetch blockchain data');
        }
        return await response.json();
    } catch (error) {
        console.error('Error fetching blockchain:', error);
        return null;
    }
}

function displayBlockchainBlocks() {
    fetchBlockchainData().then(blockchainData => {
        if (!blockchainData || !blockchainData.blocks) return;

        const container = document.getElementById('blockchain-container');
        if (!container) return;

        // Remove initial notification if it exists
        const initialNotification = container.querySelector('.notification');
        if (initialNotification) {
            initialNotification.remove();
        }

        // Clear existing blocks
        container.innerHTML = '';

        // Reverse the blocks so they display oldest to newest (increasing block numbers)
        const orderedBlocks = [...blockchainData.blocks].reverse();

        // Display blocks with newest (highest number) at the top
        orderedBlocks.forEach(block => {
            const blockElement = document.createElement('div');
            blockElement.className = 'block';

            const timestamp = new Date(block.timestamp * 1000).toLocaleString();

            blockElement.innerHTML = `
                <div class="block-header">
                    <div>
                        <span class="block-game-type">Block #${block.height}</span>
                        <span class="tag is-info ml-2">
                            ${block.transactions.length} Transactions
                        </span>
                    </div>
                    <div class="block-timestamp">
                        ${timestamp}
                    </div>
                </div>
                <div class="block-hash">
                    <strong>Hash:</strong> ${block.hash}
                </div>
                <div class="block-hash">
                    <strong>Previous:</strong> ${block.prevHash}
                </div>
                <div class="block-data">
                    <strong>Nonce:</strong> ${block.nonce}<br>
                    <strong>Transactions:</strong><br>
                    ${block.transactions.map(tx =>
                        `TX: ${tx.id} (${tx.inputs} inputs, ${tx.outputs} outputs)`
                    ).join('<br>')}
                </div>
            `;

            // Add to top so newest blocks appear first
            container.prepend(blockElement);
        });
    });
}

function printBlockchainAPI(gameType, result, data) {
    // Print to console as API response
    console.log('=== BLOCKCHAIN TRANSACTION ===');
    console.log(`Game: ${gameType}`);
    console.log(`Result: ${result}`);
    console.log(`Timestamp: ${new Date().toISOString()}`);
    console.log(`Data:`, data);
    console.log('=============================');

    // Refresh the blockchain display to show the latest blocks
    setTimeout(displayBlockchainBlocks, 1000);

    return { gameType, result, data };
}

function fetchBalance(address) {
    fetch(`http://localhost:6969/balance?address=${address}`)
    .then(response => {
        if (!response.ok) {
            throw new Error('That not good, ' + response.statusText);
        }
        return response.json();
    })
    .then(data => {
        // Update main balance display
        const balanceElement = document.getElementById('wallet-balance');
        const balanceText = document.getElementById('balance-text');
        if (balanceText) {
            balanceText.textContent = data.balance;
            balanceElement.style.display = 'block';
        }
    })
    .catch(error => {
        console.error('Error:', error);
        const balanceElement = document.getElementById('wallet-balance');
        const balanceText = document.getElementById('balance-text');
        if (balanceText) {
            balanceText.textContent = 'Error: ' + error.message;
            balanceElement.style.display = 'block';
        }
    });
}

document.addEventListener('DOMContentLoaded', function() {
    const savedAddress = localStorage.getItem('walletAddress');
    if (savedAddress) {
        const addressElement = document.getElementById('create-wallet-address');
        const addressText = document.getElementById('address-text');
        if (addressText) {
            addressText.textContent = savedAddress;
            addressElement.style.display = 'block';
        }
        fetchBalance(savedAddress);
    }

    // Display initial blockchain data
    displayBlockchainBlocks();
});

document.getElementById('create-wallet-btn').onclick = function(e) {
    e.preventDefault();
    e.stopPropagation();

    fetch('http://localhost:6969/createwallet', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('That not good, ' + response.statusText);
        }
        return response.json();
    })
    .then(data => {
        localStorage.setItem('walletAddress', data.address);

        // Update address display
        const addressElement = document.getElementById('create-wallet-address');
        const addressText = document.getElementById('address-text');
        if (addressText) {
            addressText.textContent = data.address;
            addressElement.style.display = 'block';
        }

        // Update message display
        const messageElement = document.getElementById('create-wallet-message');
        const messageText = document.getElementById('message-text');
        if (messageText) {
            messageText.textContent = data.message;
            messageElement.style.display = 'block';
            messageElement.className = 'notification is-success';
        }

        // Auto-run blockchain print API for wallet creation
        printBlockchainAPI('WALLET_CREATION', 'SUCCESS', {
            address: data.address,
            message: data.message,
            timestamp: new Date().toISOString()
        });

        fetchBalance(data.address);
    })
    .catch(error => {
        console.error('Error:', error);
        const messageElement = document.getElementById('create-wallet-message');
        const messageText = document.getElementById('message-text');
        if (messageText) {
            messageText.textContent = 'Error: ' + error.message;
            messageElement.style.display = 'block';
            messageElement.className = 'notification is-danger';
        }
    });

    return false;
};

document.getElementById('coinflip-gamble-btn').onclick = function(e) {
    e.preventDefault();
    e.stopPropagation();

    const betAmount = document.getElementById('coinflip-bet-amount').value;
    const address = localStorage.getItem('walletAddress') || document.getElementById('create-wallet-address').innerText.replace('Address: ', '');
    if (!betAmount || isNaN(betAmount) || betAmount <= 0) {
        alert('Please enter a valid bet amount.');
        return false;
    }

    fetch('http://localhost:6969/coinflip', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ amount: parseFloat(betAmount), from: address })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('That not good ' + response.statusText);
        }
        return response.json();
    })
    .then(data => {
        const messageElement = document.getElementById('coinflip-message');
        const messageText = document.getElementById('coinflip-message-text');
        if (messageText) {
            messageText.textContent = data.message;
            messageElement.style.display = 'block';

            // Set color based on win/loss
            if (data.result === 'WIN') {
                messageElement.className = 'notification is-success';
            } else if (data.result === 'LOSS') {
                messageElement.className = 'notification is-danger';
            } else {
                messageElement.className = 'notification is-info';
            }
        }

        // Auto-run blockchain print API
        printBlockchainAPI('COINFLIP', data.result, {
            address: address,
            amount: parseFloat(betAmount),
            message: data.message,
            timestamp: new Date().toISOString()
        });

        fetchBalance(address);
    })
    .catch(error => {
        console.error('Error:', error);
        const messageElement = document.getElementById('coinflip-message');
        const messageText = document.getElementById('coinflip-message-text');
        if (messageText) {
            messageText.textContent = 'Error: ' + error.message;
            messageElement.style.display = 'block';
            messageElement.className = 'notification is-danger';
        }
    });

    return false;
}

document.getElementById('diceroll-gamble-btn').onclick = function(e) {
    e.preventDefault();
    e.stopPropagation();

    const betAmount = document.getElementById('diceroll-bet-amount').value;
    const address = localStorage.getItem('walletAddress') || document.getElementById('create-wallet-address').innerText.replace('Address: ', '');
    if (!betAmount || isNaN(betAmount) || betAmount <= 0) {
        alert('Please enter a valid bet amount.');
        return false;
    }

    fetch('http://localhost:6969/diceroll', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ amount: parseFloat(betAmount), from: address })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('That not good ' + response.statusText);
        }
        return response.json();
    })
    .then(data => {
        const messageElement = document.getElementById('diceroll-message');
        const messageText = document.getElementById('diceroll-message-text');
        if (messageText) {
            messageText.textContent = data.message;
            messageElement.style.display = 'block';

            // Set color based on win/loss
            if (data.result === 'WIN') {
                messageElement.className = 'notification is-success';
            } else if (data.result === 'LOSS') {
                messageElement.className = 'notification is-danger';
            } else {
                messageElement.className = 'notification is-info';
            }
        }

        // Auto-run blockchain print API
        printBlockchainAPI('DICE_ROLL', data.result, {
            address: address,
            amount: parseFloat(betAmount),
            message: data.message,
            timestamp: new Date().toISOString()
        });

        fetchBalance(address);
    })
    .catch(error => {
        console.error('Error:', error);
        const messageElement = document.getElementById('diceroll-message');
        const messageText = document.getElementById('diceroll-message-text');
        if (messageText) {
            messageText.textContent = 'Error: ' + error.message;
            messageElement.style.display = 'block';
            messageElement.className = 'notification is-danger';
        }
    });

    return false;
}

document.getElementById('NumberGuess-gamble-btn').onclick = function(e) {
    e.preventDefault();
    e.stopPropagation();

    const betAmount = document.getElementById('NumberGuess-bet-amount').value;
    const guess = document.getElementById('NumberGuess-bet-guess').value;
    if (!guess || isNaN(guess) || guess < 1 || guess > 100) {
        alert('Please enter a valid guess between 1 and 100.');
        return false;
    }
    const address = localStorage.getItem('walletAddress') || document.getElementById('create-wallet-address').innerText.replace('Address: ', '');
    if (!betAmount || isNaN(betAmount) || betAmount <= 0) {
        alert('Please enter a valid bet amount.');
        return false;
    }

    fetch('http://localhost:6969/numberrange', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ amount: parseFloat(betAmount), guess: parseInt(guess), from: address })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('That not good ' + response.statusText);
        }
        return response.json();
    })
    .then(data => {
        const messageElement = document.getElementById('NumberGuess-message');
        const messageText = document.getElementById('NumberGuess-message-text');
        if (messageText) {
            messageText.textContent = data.message;
            messageElement.style.display = 'block';

            // Set color based on win/loss
            if (data.result === 'WIN') {
                messageElement.className = 'notification is-success';
            } else if (data.result === 'LOSS') {
                messageElement.className = 'notification is-danger';
            } else {
                messageElement.className = 'notification is-info';
            }
        }

        // Auto-run blockchain print API
        printBlockchainAPI('NUMBER_GUESS', data.result, {
            address: address,
            amount: parseFloat(betAmount),
            guess: parseInt(guess),
            message: data.message,
            timestamp: new Date().toISOString()
        });

        fetchBalance(address);
    })
    .catch(error => {
        console.error('Error:', error);
        const messageElement = document.getElementById('NumberGuess-message');
        const messageText = document.getElementById('NumberGuess-message-text');
        if (messageText) {
            messageText.textContent = 'Error: ' + error.message;
            messageElement.style.display = 'block';
            messageElement.className = 'notification is-danger';
        }
    });

    return false;
}
