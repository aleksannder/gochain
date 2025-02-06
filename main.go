package main

import (
	"github.com/aleksannder/gochain/domain"
)

func main() {

	bc := domain.NewBlockchain()

	defer bc.Db.Close()

	cli := domain.CLI{Bc: bc}
	cli.Run()
}
