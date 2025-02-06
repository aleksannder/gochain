package domain

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

type Block struct {
	Timestamp     int64  // Block creation timestamp
	Data          []byte // Valuable information in the block
	PrevBlockHash []byte // Hash of the previous block
	Hash          []byte // Hash of the block
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
	}
	block.SetHash()
	return block
}

func (b *Block) SetHash() {

	// Converts Timestamp -> string -> byte array
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	// Joins by creating new byte slices from the previous block hash, the data and the timestamp
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})

	// Hashes the concatenated headers
	hash := sha256.Sum256(headers)

	// Adds the new hash
	b.Hash = hash[:]
}
