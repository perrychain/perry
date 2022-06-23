package blockdb

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Hash [32]byte

type BlockDB struct {
	File     *os.File
	Filename string
	Version  uint8
	Blocks   []BlockKV
	Mu       sync.RWMutex
}

type BlockKV struct {
	Key   Hash  `json:"hash"`
	Value Block `json:"block"`
}

type Block struct {
	Header  BlockHeader `json:"header"`
	Payload []TxPayload `json:"payload"`
}

type BlockHeader struct {
	Parent  Hash      `json:"parent"`
	SeqID   uint64    `json:"seqid"`
	SeqTime time.Time `json:"seqtime"`
}

type TxPayload struct {
	Sender    []byte `json:"sender"`
	Recipient []byte `json:"recipient"`
	Signature []byte `json:"signature"`
	Data      []byte `json:"data"`
	Header    []byte `json:"header"`
	Type      uint8  `json:"type"`
	Reserved  uint8  `json:"reserved"`
	Output    []byte `json:"output"`
	Block     uint64
}

type SyncBlocks struct {
	PublicKey []byte    `json:"public_key"`
	Blocks    []BlockKV `json:"blocks"`
}

func New(filename string) BlockDB {

	return BlockDB{Version: 1, Filename: filename}

}

// Open the specified database file
func (blockdb *BlockDB) Open() (err error) {

	// Return if no path specified
	if blockdb.Filename == "" {
		return
	}

	fmt.Println("Opening => ", blockdb.Filename)

	// Open the specified file, create a new file if missing
	f, err := os.OpenFile(blockdb.Filename, os.O_RDONLY|os.O_CREATE, 0755)

	if err != nil {
		return err
	}

	var buf []byte
	scanner := bufio.NewScanner(f)
	// Increase the buffer to read the larger blocks stored on disk
	// TODO: Limit each block size to 2MB max
	scanner.Buffer(buf, 2048*1024)

	for scanner.Scan() {

		if err := scanner.Err(); err != nil {
			return err
		}

		var blockKv BlockKV
		err = json.Unmarshal(scanner.Bytes(), &blockKv)

		if err != nil {
			return err
		}

		blockdb.Blocks = append(blockdb.Blocks, blockKv)

	}

	if err := f.Close(); err != nil {
		return err
	}

	return

}

// Append a new block to disk, new-line seperated
func (blockdb *BlockDB) Append(block []byte) (err error) {

	f, err := os.OpenFile(blockdb.Filename, os.O_RDWR|os.O_APPEND, 0755)

	if err != nil {
		return err
	}

	_, err = f.Write(append(block, '\n'))

	return

}

// Verfiy the blockchain DB on disk matches the specified signature
// TODO: Use multiple go routines
func (blockdb *BlockDB) Verify() (err error) {

	blockdb.Mu.RLock()

	fmt.Printf("Verifying (%d) block signatures on disk ...", len(blockdb.Blocks))

	var currentHash Hash

	for i := 0; i < len(blockdb.Blocks); i++ {

		currentBlock := &blockdb.Blocks[i]

		h := sha256.New()

		payload, err := json.Marshal(currentBlock.Value.Payload)

		if err != nil {
			log.Fatal(err)
		}

		if i > 0 {
			currentHash = blockdb.Blocks[i-1].Key

		}

		// Write the current hash
		// Append the hash for the current block data state
		h.Write(append(currentHash[:], payload...))

		checksum := h.Sum(nil)

		if !bytes.Equal(checksum, currentBlock.Key[:]) {
			fmt.Println("Checksum does not match ", i, " => ", checksum, currentBlock.Key)
		}

	}

	blockdb.Mu.RUnlock()

	fmt.Printf(" done\n")
	return

}

// JSON RPC methods

// Return the latest message block in our stack
func (blockdb *BlockDB) GetLatestBlock() (block *BlockKV) {

	if len(blockdb.Blocks) == 0 {
		return &BlockKV{}
	}

	len := len(blockdb.Blocks) - 1

	return &blockdb.Blocks[len]

}

// Search for a specified block starting from a specified `from` sequence
func (blockdb *BlockDB) Sync(from []byte) (id int) {

	blockdb.Mu.RLock()
	defer blockdb.Mu.RUnlock()

	//fmt.Println("Searching for ", from)

	for i := 0; i < len(blockdb.Blocks); i++ {
		//fmt.Println("Searching ", i)

		if bytes.Equal(blockdb.Blocks[i].Value.Header.Parent[:], from[:]) {
			//fmt.Println("WE HAVE A MATCH!", i)
			return i
		}

	}

	return

}
