package main

import (
	"fmt"
	"github.com/aleksannder/gochain/domain"
	"strconv"

)

func main() {

	bc := domain.NewBlockchain()

	bc.AddBlock("Send 1 BTC to Zeko")
	bc.AddBlock("Send 3 BTC to Zeko")

	for _, block := range bc.Blocks {
		fmt.Printf("Previous hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Data: %s\n", block.Data)
		pow := domain.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
