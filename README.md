# BLOCKCHAIN IN GO (GOCHAIN)
**Gochain** is a mini blockchain project of mine that implements fundamental blockchain features.  
It includes core functionalities such as  
- **block creation**,
- **transaction handling**,
- **a proof-of-work consensus mechanism**,
- **basic data integrity and validation**, 
- **a reward system for mining**  
  
This project serves as a lightweight demonstration of how blockchain technology works, focusing on 
- **decentralization principles**, 
- **cryptographic hashing**,
- **and chain verification**

  
The project really helped me understand the core of blockchain technology and how its unique transactional system
enables the development of a decentralized ledger. This decentralized approach can compete with traditional centralized 
ledger systems while offering even better performance in high-throughput scenarios.  
One key component missing in this project is a true peer-to-peer system, which is essential to blockchain technology. 
This is definitely a goal I plan to pursue in the future, as it would showcase the performance capabilities of
a decentralized ledger and highlight the core of blockchains like Bitcoin and Ethereum.

# Key Principles

This blockchain is inspired by the Bitcoin Core blockchain implementation.  
The key components include:

- **Block hashing** using RIPEMD-160 and SHA-256 algorithms
- **Proof of Work mechanism** (Hashcash PoW) for block validation
- **A simple key-value database** for storing *blocks* and *chainstate* metadata
- **Bitcoin-like transaction system** with UTXO management
- **ECDSA for standardized addressing and digital signatures**
- **Mining reward system** for incentivizing participants
- **Unspent Transaction Output (UTXO) set** for efficient transaction processing
- **Simplified Payment Verification (SPV) with Merkle trees**
- **Basic implementation of core nodes** found in a blockchain network
- **A simple CLI** for interacting with the blockchain


### AVAILABLE COMMANDS
- `getbalance -address ADDRESS`
    - Get the balance of the given address.
  

- `createblockchain -address ADDRESS`
    - Create a blockchain and send the genesis block reward to the specified address.
  

- `printchain`
    - Print all the blocks of the blockchain.
  

- `send -from FROM -to TO -amount AMOUNT -mine`
    - Send a specified amount of coins from the `FROM` address to the `TO` recipient.  
      The `-mine` flag mines on the same node when set.
  

- `createwallet`
    - Generate a new key-pair and save it to the wallet file.
  

- `listaddresses`
    - List all addresses stored in the wallet file.
  

- `startnode -miner ADDRESS`
    - Start a node with the ID specified in the `NODE_ID` environment variable.  
      The `-miner` flag enables mining on that node.

## Running the App + Example

To run the app, make sure you have Go installed on your system.

### Steps to run:

1. Run `go mod download` to download and cache dependencies for faster builds.
2. Run `go build` to build the app binary.
3. **Ensure the `NODE_ID` environment variable is set** before running the app.
4. After building the app, run it in the terminal using:  
   `./gochain <command>`

### Example:

In the example we will have three separate terminal instances with their NODE_ID env. variables respectively set
to 3000, 3001 and 3002.  

- NODE 3000
  1. Create a wallet and a blockchain  
    `./gochain createwallet`  
     `./gochain createblockchain -address NODE_3000_WALLET`
  2. The blockchain database file will now contain a single genesis block, we need to save this block so we can use it in other nodes  
    `cp blockchain_3000.db blockchain_genesis.db`
  

- NODE 3001
  1. Create a few wallet addresses (in this example we will have three wallets respectively named WALLET_1, WALLET_2 and WALLET_3  
  `./gochain createwallet`  
 

- NODE 3000  
    1. Send some coins to the wallet addresses made on NODE 3001  
       `./gochain send -from NODE_3000_WALLET -to WALLET_1 -amount 10 -mine `  
       `./gochain send -from NODE_3000_WALLET -to WALLET_2 -amount 10 -mine `
  2. After this start the node and leave it running until the end of the scenario
  

- NODE 3001
    1. Copy the genesis block and start the node  
       `cp blockchain_genesis.db blockchain_3001.db`  
        `./gochain startnode`
  2. Now this node will download all the blocks from the central node, you can now stop the node and check the balances of all wallets  
     `./gochain getbalance -address WALLET_1`  
     `./gochain getbalance -address WALLET_2`  
     `./gochain getbalance -address NODE_3000_WALLET`
  

- NODE 3002
    1. This will be our miner node, but first we have to initialize the blockchain  
       `cp blockchain_genesis.db blockchain_3002.db`  
       `./gochain createwallet`
  2. Now you can start the node as a miner node  
     `./gochain startnode -miner MINER_WALLET`  
  

- NODE 3001  
    1. Send some coins  
       `./gochain send -from WALLET_1 -to WALLET_2 -amount 3`  
       `./gochain send -from WALLET_1 -to WALLET_2 -amount 1`
  

- NODE 3002  
    1. Quickly switch to this node so you can see the mining process in progress. Also check the output of the central node  
  

- NODE 3001  
    1. Start the node and let it download the newly mined block  
        `./gochain startnode`
  2. Now stop and check the balances of all wallets.

# REFERENCES
This project was inspired by a project I found in the [Project-based learning](https://github.com/practical-tutorials/project-based-learning?tab=readme-ov-file#go) GitHub repository.  
The original project was created by [Jeiwan](https://github.com/Jeiwan) and their GitHub repository can be found [here](https://github.com/Jeiwan).