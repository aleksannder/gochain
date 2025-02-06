package main

import (
	"fmt"
	"github.com/aleksannder/gochain/domain"
)

func main() {

	bc := domain.NewBlockchain()

	bc.AddBlock("Send 1 BTC to Zeko")
	bc.AddBlock("Send 3 BTC to Zeko")

	for _, block := range bc.Blocks {
		fmt.Printf("Previous hash: %s\n", block.PrevBlockHash)
		fmt.Printf("Hash: %s\n", block.Hash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Println()
	}
}
