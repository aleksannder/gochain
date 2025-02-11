package domain

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}
var blocksInTransit [][]byte
var mempool = make(map[string]Transaction)

type addr struct {
	AddrList []string
}

type block struct {
	AddrFrom string
	Block    []byte
}

type tx struct {
	AddrFrom    string
	Transaction []byte
}

type getblocks struct {
	AddrFrom string
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type ver struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

func StartServer(nodeID, minerAddr string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddr
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Server listening on %s", miningAddress)
	defer func(ln net.Listener) {
		err := ln.Close()
		if err != nil {
			log.Panic(err)
		}
	}(ln)

	bc := NewBlockchain(nodeID)

	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn, bc)
	}
}

func sendVersion(addr string, bc *Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(ver{nodeVersion, bestHeight, nodeAddress})

	request := append(commandToBytes("ver"), payload...)

	sendData(addr, request)
}

func commandToBytes(cmd string) []byte {
	var b [commandLength]byte

	for i, c := range cmd {
		b[i] = byte(c)
	}

	return b[:]
}

func bytesToCommand(data []byte) string {
	var command []byte

	for _, b := range data {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func handleConnection(conn net.Conn, bc *Blockchain) {
	req, err := io.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(req[:commandLength])
	log.Printf("Received command %s\n", command)

	switch command {
	case "addr":
		handleAddr(req)
	case "block":
		handleBlock(req, bc)
	case "inv":
		handleInv(req, bc)
	case "getblocks":
		handleGetBlocks(req, bc)
	case "getdata":
		handleGetData(req, bc)
	case "tx":
		handleTx(req, bc)
	case "ver":
		handleVersion(req, bc)
	default:
		log.Printf("Unknown command %s\n", command)
	}

	err = conn.Close()
	if err != nil {
		log.Panic(err)
	}

}

func handleVersion(req []byte, bc *Blockchain) {
	var buf bytes.Buffer
	var payload ver

	buf.Write(req[commandLength:])
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	localBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if localBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if localBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}

	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}

func handleGetBlocks(req []byte, bc *Blockchain) {
	var buf bytes.Buffer
	var payload getblocks

	buf.Write(req[commandLength:])
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

func handleInv(req []byte, bc *Blockchain) {
	var buf bytes.Buffer
	var payload inv

	buf.Write(req[commandLength:])
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func handleGetData(req []byte, bc *Blockchain) {
	var buf bytes.Buffer
	var payload getdata

	buf.Write(req[commandLength:])
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		SendTx(payload.AddrFrom, &tx)
	}
}

func handleBlock(req []byte, bc *Blockchain) {
	var buf bytes.Buffer
	var payload block

	buf.Write(req[commandLength:])
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := DeserializeBlock(blockData)

	log.Printf("Received new block")

	bc.AddBlock(block)

	log.Printf("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := UTXOSet{Blockchain: bc}
		UTXOSet.Update(block)
	}

}

func handleTx(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx

	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddrFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(mempool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*Transaction

			for id := range mempool {
				tx := mempool[id]
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}

			cbTx := NewCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := bc.MineBlock(txs)
			UTXOSet := UTXOSet{bc}
			UTXOSet.Update(newBlock)

			fmt.Println("New block is mined!")

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}

			for _, node := range knownNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}

			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}

func handleAddr(req []byte) {
	var buf bytes.Buffer
	var payload addr

	buf.Write(req[commandLength:])
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	knownNodes = append(knownNodes, payload.AddrList...)
	log.Printf("Known nodes updated, there are %d known nodes now\n", len(knownNodes))
	requestBlocks()
}

func sendInv(addr, kind string, items [][]byte) {
	inv := inv{nodeAddress, kind, items}
	payload := gobEncode(inv)
	req := append(commandToBytes("inv"), payload...)

	sendData(addr, req)
}

func sendData(address string, data []byte) {
	conn, err := net.Dial(protocol, address)
	if err != nil {
		log.Printf("%s isn't available\n. Error message: %x", address, err)
		var updatedNodes []string

		for _, node := range knownNodes {
			if node != address {
				updatedNodes = append(updatedNodes, node)
			}
		}

		knownNodes = updatedNodes

		return
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Panic(err)
		}
	}(conn)

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func sendGetData(addr, kind string, id []byte) {
	payload := gobEncode(getdata{nodeAddress, kind, id})
	req := append(commandToBytes("getdata"), payload...)

	sendData(addr, req)
}

func SendTx(addr string, transx *Transaction) {
	data := tx{nodeAddress, transx.Serialize()}
	payload := gobEncode(data)
	req := append(commandToBytes("tx"), payload...)

	sendData(addr, req)
}

func sendBlock(addr string, b *Block) {
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	req := append(commandToBytes("block"), payload...)

	sendData(addr, req)
}

func sendGetBlocks(addr string) {
	payload := gobEncode(getblocks{nodeAddress})
	req := append(commandToBytes("getblocks"), payload...)

	sendData(addr, req)
}

func gobEncode(data interface{}) []byte {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buf.Bytes()
}

func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}
