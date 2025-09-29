
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