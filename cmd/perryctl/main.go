package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"time"

	"github.com/perrychain/perry/pkg/blockdb"
	"github.com/perrychain/perry/pkg/p2pnet"
	"github.com/perrychain/perry/pkg/wallet"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {

	var cmd = flag.String("cmd", "verify", "verify blockchain DB")
	var dbpath = flag.String("dbpath", ".blockchain.db", "Path to blockchain DB")
	var msg = flag.String("msg", "Hello world", "Message to sign")
	var format = flag.String("format", "base64", "Format hex or base64 (default)")

	usr, _ := user.Current()
	defaultHomeDir := fmt.Sprintf("%s/.perry", usr.HomeDir)
	defaultWalletPath := fmt.Sprintf("%s/.wallet.json", defaultHomeDir)

	var walletPath = flag.String("wallet", defaultWalletPath, "Specify wallet path")

	flag.Parse()

	if *cmd == "verify" {
		verify(*dbpath)
	} else if *cmd == "sign" {
		sign(*dbpath, *msg, *walletPath, *format)
	}

}

func sign(dbpath, msg, walletPath, format string) {

	mywallet, err := wallet.Load(walletPath)

	if err != nil {
		log.Warn("No wallet specified, generating new one\n")
		mywallet.GenerateWallet()
	}

	sdata, err := mywallet.Sign([]byte(msg))

	if err != nil {
		log.Warn("Error signing data %s", err)
	}

	if format == "hex" {
		fmt.Printf("Signed Sig => %x\n", sdata)
		fmt.Printf("Public Key => %x\n", mywallet.PublicKey)
		fmt.Printf("Message => %s\n", msg)
		fmt.Printf("Private Key => %x\n", mywallet.PrivateKey)
	} else {

		pck := p2pnet.Packet{}

		copy(pck.Payload[:], msg) //[]byte(msg)
		copy(pck.SenderPublicKey[:], mywallet.PublicKey)
		copy(pck.SenderSignature[:], sdata)

		fmt.Printf("Signed Sig => %s\n", base64.StdEncoding.EncodeToString(pck.SenderSignature[:]))
		fmt.Printf("Public Key => %s\n", base64.StdEncoding.EncodeToString(pck.SenderPublicKey[:]))
		fmt.Printf("Message => %s\n", base64.StdEncoding.EncodeToString(pck.Payload[:]))
		fmt.Printf("Private Key => %s\n", base64.StdEncoding.EncodeToString(mywallet.PrivateKey))

	}

}

func verify(dbpath string) {

	start := time.Now()

	fmt.Println(fmt.Sprintf("Verfiying blockchain DB => %s", dbpath))

	db := blockdb.New(dbpath)
	db.Open()
	db.Verify()

	fmt.Println("Number of entries: ", len(db.Blocks))

	totalRows := 0
	mywallet := wallet.New()
	cpu_cores := runtime.NumCPU()

	//pool := pond.New(cpu_cores-1, cpu_cores*2, pond.Strategy(pond.Eager()))

	// Distribute jobs on each core for the specified sequence
	tasks := uint64(len(db.Blocks))

	//var error_abort bool

	for i := uint64(0); i < tasks; i++ {

		//pool.Submit(func() {

		totalRows += len(db.Blocks[i].Value.Payload)

		for i2 := 0; i2 < len(db.Blocks[i].Value.Payload); i2++ {

			packet := db.Blocks[i].Value.Payload[i2]

			// TODO: Match header names with wallet/p2p implementation
			fmt.Println(string(packet.Data))
			verify := mywallet.VerifyRaw(packet.Sender[:], packet.Data[:], packet.Signature[:])

			//log.Debug(fmt.Sprintf("\t(Block %d, Payload %d) Status %t : Verifying Sender => %s, Data => %s, Signature => %s\n", i, i2, verify, base64.StdEncoding.EncodeToString(packet.Sender[:]), base64.StdEncoding.EncodeToString(packet.Data[:]), base64.StdEncoding.EncodeToString(packet.Signature[:])))

			if !verify {
				log.Warn(fmt.Sprintf("Transaction verification failed! SeqID => %d Payload => %d", db.Blocks[i].Value.Header.SeqID, i2))
			}

		}

		//})
	}

	fmt.Println("Number of unique transactions: ", totalRows)

	// Stop the pool and wait for all submitted tasks to complete
	//pool.StopAndWait()

	timer := time.Now()
	elapsed := timer.Sub(start)

	VerifyHashRate := uint32(float64(totalRows) * (1 / elapsed.Seconds()))
	VerifyHashRatePerCore := VerifyHashRate / uint32(cpu_cores)

	fmt.Println(fmt.Sprintf("BlockchainDB > VerifySignature per sec: %d", VerifyHashRate))
	fmt.Println(fmt.Sprintf("BlockchainDB > VerifySignaturePerCore per sec: %d", VerifyHashRatePerCore))

}
