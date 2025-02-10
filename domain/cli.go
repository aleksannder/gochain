package domain

import (
	"flag"
	"fmt"
	"github.com/aleksannder/gochain/util"
	"github.com/boltdb/bolt"
	"log"
	"os"
	"strconv"
)

type CLI struct{}

// Usage CLI

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\t getbalance -address ADDRESS -> Get balance of given address")
	fmt.Println("\t createblockchain -address ADDRESS -> Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("\t printchain -> Print all the blocks of the blockchain")
	fmt.Println("\t send -from FROM -to TO -amount AMOUNT -> Send AMOUNT of coins from FROM address to TO recipient")
	fmt.Println("\t createwallet -> Generates new key-pair and saves to wallet file")
	fmt.Println("\t listaddresses -> Lists all addresses from wallet file")

}

// Validate CLI args

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// Run CLI

func (cli *CLI) Run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	reindexUtxoCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "Address of wallet")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "Address of wallet")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
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
	case "reindexutxo":
		err := reindexUtxoCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}
	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount == 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
	if createWalletCmd.Parsed() {
		cli.createWallet()
	}
	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}
	if reindexUtxoCmd.Parsed() {
		cli.reindexUtxo()
	}
}

func (cli *CLI) createBlockchain(address string) {
	bc := CreateBlockchain(address)
	defer func(Db *bolt.DB) {
		err := Db.Close()
		if err != nil {
			log.Panic(err)
		}
	}(bc.Db)

	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()
	fmt.Println("Done")
}

func (cli *CLI) getBalance(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address invalid")
	}

	bc := NewBlockchain()
	UTXOSet := UTXOSet{bc}
	defer func(Db *bolt.DB) {
		err := Db.Close()
		if err != nil {
			log.Panic(err)
		}
	}(bc.Db)

	balance := 0
	pubKeyHash := util.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) printChain() {
	bc := NewBlockchain()
	defer func(Db *bolt.DB) {
		err := Db.Close()
		if err != nil {
			log.Panic(err)
		}
	}(bc.Db)

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("======= Block %x =======\n", block.Hash)
		fmt.Printf("Previous block: %x\n", block.PrevBlockHash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) send(from, to string, amount int) {
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient Address invalid")
	}
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender Address invalid")
	}

	bc := NewBlockchain()
	UTXOSet := UTXOSet{bc}
	defer func(Db *bolt.DB) {
		err := Db.Close()
		if err != nil {
			log.Panic(err)
		}
	}(bc.Db)

	tx := NewUTXOTransaction(from, to, amount, &UTXOSet)
	cbTx := NewCoinbaseTX(from, "")
	txs := []*Transaction{cbTx, tx}

	newBlock := bc.MineBlock(txs)
	UTXOSet.Update(newBlock)
	fmt.Printf("\n\n Success")
}

func (cli *CLI) listAddresses() {
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Success, your new address is: %s\n", address)
}

func (cli *CLI) reindexUtxo() {
	bc := NewBlockchain()

	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Reindexed, there are %d transactions in the UTXO set.\n", count)
}
