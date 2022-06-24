package poh_hash

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/alitto/pond"
	"github.com/gin-gonic/gin"
	"github.com/perrychain/perry/pkg/blockdb"
	"github.com/perrychain/perry/pkg/wallet"

	log "github.com/sirupsen/logrus"
)

var genisis_hash string

type POH_Entry struct {
	Hash      []byte
	Data      []byte
	Signature []byte
	Seq       uint64
}

type POH_Epoch struct {
	Entry []POH_Entry
	Block uint64
	Epoch uint32
}

type POH_Block struct {
	Block uint64
}

type POH struct {
	POH                   []POH_Epoch
	QueueSync             QueueSync
	HashRate              uint32
	VerifyHashRate        uint32
	VerifyHashRatePerCore uint32
	TickRate              uint64
	Mu                    sync.RWMutex
	Wallet                wallet.Wallet
	BlockDB               blockdb.BlockDB
	currentBlock          blockdb.Block
}

type QueueSync struct {
	State []blockdb.TxPayload
}

/*
type QueueData struct {
	Data      []byte
	Sender    []byte
	Recipient []byte
	Block     uint64
}
*/

type SyncState struct {
	Entry []POH_Entry
	Len   int
}

// Create a new PoH using a specified wallet file
func New(wallet_path, db_path string) (this POH) {

	this.TickRate = 1_000_000

	if wallet_path == "" {
		wallet_path = ".perry-wallet.json"
	}

	var err error
	this.Wallet, err = wallet.Load(wallet_path)

	if err != nil {
		err = this.Wallet.GenerateWallet()

		if err != nil {
			log.Fatal("Could not create wallet %s", err)
		}

		err = this.Wallet.Save(wallet_path, false)

		if err != nil {
			log.Fatal("Could not save wallet file %s (%s)", wallet_path, err)
		}

	}

	// Specify the blockchain database
	this.BlockDB = blockdb.New(db_path)

	return

}

// Push data waiting in the queue to the current PoH block calculation
func (poh *POH) FetchDataState(block uint64) (payload blockdb.TxPayload, chk bool) {

	for i := 0; i < len(poh.QueueSync.State); i++ {

		if poh.QueueSync.State[i].Block > 0 {
			continue
		} else {
			poh.Mu.Lock()
			poh.QueueSync.State[i].Block = block
			poh.Mu.Unlock()
			return poh.QueueSync.State[i], true
		}

	}

	return

}

func (poh *POH) GeneratePOH(count uint64) {

	start := time.Now()

	//tickcount := count / this.TickRate
	poh.POH = append(poh.POH, POH_Epoch{Epoch: 1})

	h := sha256.New()
	var prevhash []byte

	err := poh.BlockDB.Open()

	if err != nil {
		log.Fatal(fmt.Sprintf("Could not open BlockDB: %s", err))
	}

	err = poh.BlockDB.Verify()

	if err != nil {
		log.Fatal(fmt.Sprintf("Could not verify BlockDB: %s", err))
	}

	poh.currentBlock.Payload = make([]blockdb.TxPayload, 0)

	if len(poh.BlockDB.Blocks) > 0 {
		// Get the last hash from the previous block
		key := poh.BlockDB.Blocks[len(poh.BlockDB.Blocks)-1].Key
		h.Write(key[:])

		log.Debug("Using last block hash => ", key)

	} else {
		// TODO: Geneisis + wallet public-key + timestamp + rand number
		genisis_hash := fmt.Sprintf("GENESIS_HASH-%d-%s", time.Now().UnixNano(), base64.StdEncoding.EncodeToString(poh.Wallet.PrivateKey))
		h.Write([]byte(genisis_hash))

		log.Debug("Using Genesis Hash => ", genisis_hash)

	}

	prevhash = h.Sum(nil)

	poh.Mu.Lock()
	poh.POH[0].Entry = append(poh.POH[0].Entry, POH_Entry{Hash: prevhash, Seq: 0})
	poh.Mu.Unlock()

	// Spawn go routine for block confirmation thread
	pohBlock := make(chan POH_Block)
	go poh.BlockConfirmation(pohBlock)

	// Add a channel ticket to write blocks every 500ms to disk
	var blockid int

	ticker := time.NewTicker(500 * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				pohBlock <- POH_Block{Block: uint64(blockid)}
				log.Debug("Tick for Block ID", blockid)
				blockid++
			}
		}
	}()

	// Loop generating a PoH for a specified period
	for i := uint64(1); i < uint64(count); i++ {

		h := sha256.New()

		// TODO: Optimise, periodically push events published off the stack
		block, chk := poh.FetchDataState(i)
		t := i % uint64(poh.TickRate)

		// Create a new hash from a data request
		if chk {
			h.Write(append(prevhash, block.Data...))
			prevhash = h.Sum(nil)

			// Sign the transaction from the validator
			signature, _ := poh.Wallet.Sign(block.Data)

			poh.Mu.Lock()
			poh.POH[0].Entry = append(poh.POH[0].Entry, POH_Entry{Hash: prevhash, Data: block.Data, Seq: i, Signature: signature})
			poh.Mu.Unlock()

			payload := blockdb.TxPayload{}

			payload.Data = block.Data
			payload.Signature = block.Signature
			payload.Recipient = block.Recipient
			payload.Sender = block.Sender

			poh.currentBlock.Payload = append(poh.currentBlock.Payload, payload)

		} else {
			// Hash the latest output, hash of a hash for POH
			h.Write(prevhash)
			prevhash = h.Sum(nil)

		}

		if t == 0 {
			// Only save a state every X events (based on TickSize) to reduce memory allocation
			poh.Mu.Lock()
			poh.POH[0].Entry = append(poh.POH[0].Entry, POH_Entry{Hash: prevhash, Seq: i})
			poh.Mu.Unlock()

		}

	}

	timer := time.Now()
	elapsed := timer.Sub(start)

	poh.HashRate = uint32(float64(count) * (1 / elapsed.Seconds()))

}

func (poh *POH) VerifyPOH(cpu_cores int) (err error) {

	start := time.Now()

	// TODO: Revise - Keep one CPU core available for other tasks and scheduling, benchmark improvement.
	//if cpu_cores > 4 {
	//	cpu_cores -= 1
	//}

	/*
		panicHandler := func(p interface{}) {
			fmt.Printf("Task panicked: %v", p)
		}
	*/

	pool := pond.New(cpu_cores-1, cpu_cores*2, pond.Strategy(pond.Eager())) //, pond.MinWorkers(cpu_cores), pond.PanicHandler(panicHandler))

	// Distribute jobs on each core for the specified sequence
	tasks := uint64(len(poh.POH[0].Entry))

	var error_abort bool

	for i := uint64(1); i < tasks; i++ {
		log.Debug("Job started => ", i, tasks)
		n := i
		pool.Submit(func() {

			log.Debug("Job fork => ", n)

			seqstart := poh.POH[0].Entry[n-1].Seq
			seqend := poh.POH[0].Entry[n].Seq

			var prevhash []byte

			for a := seqstart; a <= seqend; a++ {

				h := sha256.New()

				// Confirm if the beginning block
				if a == seqstart {
					prevhash = poh.POH[0].Entry[n-1].Hash
				} else {
					// Hash the hash
					h.Write([]byte(prevhash))
					prevhash = h.Sum(nil)
				}

				// Confirm the sequence matches our sync state
				if a == seqend {

					// Hash the data block if specified
					if len(poh.POH[0].Entry[n].Data) > 0 {
						h.Write(poh.POH[0].Entry[n].Data)

						// Verify the signature matches the publickey
						verify, _ := poh.Wallet.Verify(poh.POH[0].Entry[n].Data, poh.POH[0].Entry[n].Signature)

						if !verify {
							error_abort = true
							log.Warn("POH data signature failed, sequence ID %d - Calculated (%s)", poh.POH[0].Entry[n].Seq, poh.POH[0].Entry[n].Signature)
						}

					}

					// Read lock
					poh.Mu.RLock()
					orig := poh.POH[0].Entry[n].Hash
					poh.Mu.RUnlock()

					// Compare the proof to the original
					compare := bytes.Compare(h.Sum(nil), orig)

					if compare != 0 {
						error_abort = true
						log.Warn("POH Verification failed, sequence ID %d - Calculated (%s) vs Reference (%s)", poh.POH[0].Entry[n].Seq, base64.RawStdEncoding.EncodeToString(h.Sum(nil)), base64.RawStdEncoding.EncodeToString(orig))
						// TODO: Improve stopping existing workers in the queue, revise.

					}

				}

			}

		})
	}

	// Stop the pool and wait for all submitted tasks to complete
	pool.StopAndWait()

	timer := time.Now()
	elapsed := timer.Sub(start)

	// Calculate the verification hashrate
	lastSeq := poh.POH[0].Entry[len(poh.POH[0].Entry)-1]
	poh.VerifyHashRate = uint32(float64(lastSeq.Seq) * (1 / elapsed.Seconds()))
	poh.VerifyHashRatePerCore = poh.VerifyHashRate / uint32(cpu_cores)

	log.Debug("VerifyPOH > VerifyHashRate = %d\n", poh.VerifyHashRate)
	log.Debug("VerifyPOH > VerifyHashRatePerCore = %d\n", poh.VerifyHashRatePerCore)

	if error_abort {
		return errors.New("POH validation failed")
	}

	return
}

// Confirm block PoH is valid prior to publishing to the blockchain
func (poh *POH) BlockConfirmation(block chan POH_Block) {

	// Return if no path specified
	if poh.BlockDB.Filename == "" {
		return
	}

	for {

		// Wait for a job to be pushed to the stack to create a new block
		current_block := <-block

		if len(poh.currentBlock.Payload) == 0 {
			log.Debug("No blocks to write to disk ...")
			continue
		}

		start := time.Now()
		blockLen := len(poh.currentBlock.Payload)

		log.Info("Writing block (%d) to disk for (%d) TX's ... ", current_block, blockLen)

		poh.Mu.Lock()
		payload, err := json.Marshal(poh.currentBlock.Payload)

		if err != nil {
			log.Fatal(err)
		}

		payload = poh.CreateBlock(payload, false)

		// Append the new block to disk
		err = poh.BlockDB.Append(payload)

		if err != nil {
			log.Fatal(err)
		}

		// Reset the state and unlock the mutex
		poh.currentBlock = blockdb.Block{}
		poh.Mu.Unlock()

		timer := time.Now()
		elapsed := timer.Sub(start)

		txRate := uint32(float64(blockLen) * (1 / elapsed.Seconds()))

		log.Info("done in %s, %d per sec\n", elapsed, txRate)

	}

}

func (poh *POH) CreateBlock(payload []byte, direct bool) (newpayload []byte) {

	blockJson := blockdb.BlockKV{}

	if !direct {

		hash := sha256.New()
		var previousHash blockdb.Hash
		var currentSeqID uint64

		if len(poh.BlockDB.Blocks) > 0 {
			// Find the previous block hash
			previousHash = poh.BlockDB.Blocks[len(poh.BlockDB.Blocks)-1].Key
			// Increment the block sequenceID
			currentSeqID = poh.BlockDB.Blocks[len(poh.BlockDB.Blocks)-1].Value.Header.SeqID

		} else {
			currentSeqID = 0

		}

		// Append the new hash based on the previous hash + new payload
		hash.Write(append(previousHash[:], payload...))
		currentHash := hash.Sum(nil)
		copy(blockJson.Key[:], currentHash)

		// Append the Sequence time
		blockJson.Value.Header.SeqTime = time.Now()

		// Add the parent hash
		blockJson.Value.Header.Parent = previousHash

		// Increment the block sequenceID
		blockJson.Value.Header.SeqID = currentSeqID + 1

		// Append the new TX records
		blockJson.Value.Payload = append(blockJson.Value.Payload, poh.currentBlock.Payload...)

	} else {

		if err := json.Unmarshal(payload, &blockJson); err != nil {
			panic(err)
		}

	}

	// Append to our local state
	poh.BlockDB.Mu.Lock()
	poh.BlockDB.Blocks = append(poh.BlockDB.Blocks, blockJson)
	poh.BlockDB.Mu.Unlock()

	// Prepare the JSON to write to disk
	newpayload, err := json.Marshal(blockJson)

	if err != nil {
		log.Fatal(err)
	}

	return

}

// JSON RPC methods
func (poh *POH) Index(c *gin.Context) {

	c.JSON(200, gin.H{"status": "OK"})
}

func (poh *POH) Syncstate(c *gin.Context) {

	//this.Mu.RLock()

	var records []POH_Entry
	var len = len(poh.POH[0].Entry) - 1

	start := 1
	if len > 10 {
		start = len - 10
	}

	records = append(records, poh.POH[0].Entry[0])

	for a := start; a < len; a++ {
		records = append(records, poh.POH[0].Entry[a])
	}
	c.JSON(200, SyncState{Entry: records, Len: len})

	//poh.Mu.RUnlock()

}

func (poh *POH) Syncdatastate(c *gin.Context) {

	// TODO: Find more efficient way to handle returning entries, gin-tonic limitation it seems for c.JSON
	c.Data(200, "application/json; charset=utf-8", []byte(fmt.Sprintf("{\"PublicKey\": \"%s\", \"Data\": [", base64.StdEncoding.EncodeToString(poh.Wallet.PublicKey))))

	//c.Data(200, "application/json; charset=utf-8", []byte("["))

	// Print the first
	poh.Mu.RLock()
	c.JSON(200, &poh.POH[0].Entry[0])

	len := len(poh.POH[0].Entry)

	for a := 1; a < len; a++ {

		if poh.POH[0].Entry[a].Data != nil {
			c.Data(200, "application/json; charset=utf-8", []byte(","))

			c.JSON(200, &poh.POH[0].Entry[a])

		}

	}

	poh.Mu.RUnlock()

	c.Data(200, "application/json; charset=utf-8", []byte("]}"))

}

// TODO: REplace with p2pnet.Sync
func (poh *POH) Verify(c *gin.Context) {

	start := time.Now()

	host, _ := c.GetQuery("host")

	resp, err := http.Get(fmt.Sprintf("http://%s/syncdata", host))

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	log.Debug("Response status:", resp.Status)

	timer := time.Now()
	elapsed := timer.Sub(start)

	log.Debug("Fetch syncdata %s\n", elapsed)

	var valid struct {
		PublicKey string
		Data      []POH_Entry
	}

	validator := valid

	json.NewDecoder(resp.Body).Decode(&validator)

	timer = time.Now()
	elapsed = timer.Sub(start)

	log.Debug("Decode syncdata %s\n", elapsed)

	confirmation := POH{}
	pubkey, _ := base64.StdEncoding.DecodeString(validator.PublicKey)

	confirmation.Wallet.PublicKey = pubkey

	confirmation.POH = append(confirmation.POH, POH_Epoch{Epoch: 1})
	confirmation.POH[0].Entry = validator.Data

	timer = time.Now()
	elapsed = timer.Sub(start)

	log.Debug("Prepare syncdata %s\n", elapsed)

	err = confirmation.VerifyPOH(runtime.NumCPU())

	timer = time.Now()
	elapsed = timer.Sub(start)

	log.Info("Verify syncdata %s\n", elapsed)

	if err == nil {
		c.JSON(200, gin.H{"Status": "OK", "Stats": fmt.Sprintf("Hash Rate (all cores) %d - Hash Rate per core %d", confirmation.VerifyHashRate, confirmation.VerifyHashRatePerCore)})
	} else {
		c.JSON(500, gin.H{"Status": "fail"})
		log.Warn("VerifyPOH error =>", err)
	}

}

func (poh *POH) Pushstate(c *gin.Context) {

	data, _ := c.GetQuery("data")
	sender, _ := c.GetQuery("sender")

	poh.Mu.Lock()
	queuedata := blockdb.TxPayload{Data: []byte(data), Sender: []byte(sender)}
	poh.QueueSync.State = append(poh.QueueSync.State, queuedata)
	poh.Mu.Unlock()

	c.JSON(200, queuedata)

}

func (poh *POH) State(c *gin.Context) {

	c.JSON(200, poh.QueueSync)

}
