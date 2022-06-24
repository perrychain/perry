package p2pnet

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/perrychain/perry/pkg/blockdb"
	"github.com/perrychain/perry/pkg/poh_hash"
	"github.com/perrychain/perry/pkg/wallet"
	log "github.com/sirupsen/logrus"
)

// "Safe" UDP packet size up to 508 bytes
type Packet struct {
	Version            [1]byte
	Reserved           [3]byte
	SenderPublicKey    [32]byte // TODO: Consider Base36 (with ICAP) or Base56 encoding w/ unique identifier
	RecipientPublicKey [32]byte
	Payload            [376]byte
	SenderSignature    [64]byte
}

type P2P struct {
	P2P_Node        Node            `json:"p2p_node"`
	P2P_Peers       map[string]Node `json:"p2p_peers"`
	RPC_Node        Node            `json:"rpc_node"`
	RPC_Peers       map[string]Node `json:"rpc_peers"`
	POH             *poh_hash.POH
	maxDatagramSize int
}

// JSON RPC
type Status struct {
	Parent []byte `json:"parent"`
	Hash   []byte `json:"hash"`
	SeqID  uint64 `json:"seqid"`

	P2P_Node  Node            `json:"p2p_node"`
	P2P_Peers map[string]Node `json:"p2p_peers"`
	RPC_Node  Node            `json:"rpc_node"`
	RPC_Peers map[string]Node `json:"rpc_peers"`
}

// Known Peers
type Node struct {
	Host      string    `json:"host"`
	Port      uint16    `json:"port"`
	LastSeen  time.Time `json:"lastseen"`
	Version   uint8     `json:"version"`
	Bootstrap bool      `json:"bootstrap"`
}

func New(p P2P) *P2P {

	// Set defaults
	if p.P2P_Node.Version == 0 {
		p.P2P_Node.Version = 1
	}

	if p.maxDatagramSize == 0 {
		p.maxDatagramSize = 8192
	}

	p.RPC_Peers = make(map[string]Node)

	// Set our bootstrap node
	p.RPC_Peers["127.0.0.1:24816"] = Node{Host: "127.0.0.1", Port: 24816, Bootstrap: true}

	return &p
}

// Listen on the specified UDP port for P2P blockchain traffic
func (p2p *P2P) Listen(h func(*net.UDPAddr, int, []byte)) {

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", p2p.P2P_Node.Host, p2p.P2P_Node.Port))

	log.Debug("P2P.Listen => ", addr)

	if err != nil {
		log.Fatal(err)
	}

	var l *net.UDPConn

	if p2p.P2P_Node.Host == "224.0.0.1" {
		l, err = net.ListenMulticastUDP("udp", nil, addr)

	} else {
		l, err = net.ListenUDP("udp", addr)

	}

	if err != nil {
		log.Fatal(err)
	}

	l.SetReadBuffer(8192) //this.maxDatagramSize)

	for {
		b := make([]byte, p2p.maxDatagramSize)
		n, src, err := l.ReadFromUDP(b)
		if err != nil {
			log.Warn("ReadFromUDP failed:", err)
		}
		h(src, n, b)
	}

}

// Send a packet on the P2P network via UDP
func (p2p *P2P) Send(data string) {

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", p2p.P2P_Node.Host, p2p.P2P_Node.Port))

	if err != nil {
		log.Fatal(err)
	}

	c, err := net.DialUDP("udp", nil, addr)

	if err != nil {
		log.Fatal(err)
	}

	mywallet := wallet.New()
	mywallet.GenerateWallet()

	senderwallet := wallet.New()
	senderwallet.GenerateWallet()

	packet := Packet{}

	packet.Version = [1]byte{1}
	packet.Reserved = [3]byte{0, 0, 0}

	copy(packet.SenderPublicKey[:], mywallet.PublicKey)
	copy(packet.RecipientPublicKey[:], mywallet.PublicKey)

	msg := []byte(fmt.Sprintf("%s WOW A super important message here for you to consider %s", time.Now(), data))

	copy(packet.Payload[:], msg)

	signature, _ := mywallet.Sign(packet.Payload[:])

	copy(packet.SenderSignature[:], signature)

	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, packet); err != nil {
		fmt.Println(err)
		return
	}

	c.Write(buf.Bytes())

	defer c.Close()

}

// Process a UDP packet into the queue
func (p2p *P2P) MsgHandler(src *net.UDPAddr, n int, b []byte) {
	log.Debug(n, "bytes read from", src)

	packet := Packet{}

	r := bytes.NewReader(b)

	if err := binary.Read(r, binary.BigEndian, &packet); err != nil {
		log.Warn("failed to Read:", err)
		return
	}

	// Only accept blocks that are valid
	if len(packet.SenderPublicKey) == 0 {
		log.Warn("Ignoring packet, no sender:")
		return
	}

	if len(packet.RecipientPublicKey) == 0 {
		log.Warn("Ignoring packet, no recipient:")
		return
	}

	if len(packet.Payload) == 0 {
		log.Warn("Ignoring packet, no data:")
		return
	}

	// Confirm if validated
	mywallet := wallet.New()
	verify := mywallet.VerifyRaw(packet.SenderPublicKey[:], packet.Payload[:], packet.SenderSignature[:])

	// If signed and verified, push to the stack
	if verify {
		queuedata := blockdb.TxPayload{
			Data:      packet.Payload[:],
			Sender:    packet.SenderPublicKey[:],
			Recipient: packet.RecipientPublicKey[:],
			Signature: packet.SenderSignature[:],
		}
		p2p.POH.Mu.Lock()
		p2p.POH.QueueSync.State = append(p2p.POH.QueueSync.State, queuedata)
		p2p.POH.Mu.Unlock()

	} else {
		log.Warn("Ignoring packet, signature failure")
		return
	}

}

// Query peers at a specified interval, check block state and new peers
func (p2p *P2P) QueryPeers(ctx context.Context) {

	ticker := time.NewTicker(3 * time.Second)

	for {
		select {
		case <-ticker.C:
			p2p.doSync()

		case <-ctx.Done():
			ticker.Stop()
		}
	}

}

// Query each RCP peer discovered for new blocks and peers
func (p2p *P2P) doSync() {

	myNode := fmt.Sprintf("%s:%d", p2p.RPC_Node.Host, p2p.RPC_Node.Port)

	for host := range p2p.RPC_Peers {

		if myNode == host {
			log.Info("Host, skipping my node => ", host)

		} else {
			log.Info("Host, Query blocks etc => ", host)
			p2p.querySync(host)

		}

	}

}

// Query the status of the remote RCP peer, check if block higher then local for sync
func (p2p *P2P) querySync(hostname string) {

	start := time.Now()

	// Get our latest block
	latestBlock := p2p.POH.BlockDB.GetLatestBlock()

	// Query our external host

	resp, err := http.Get(fmt.Sprintf("http://%s/p2p/status?rpc_host=%s&rpc_port=%d", hostname, p2p.RPC_Node.Host, p2p.RPC_Node.Port))

	if err != nil {
		log.Warn("Error connecting for status => ", err)
		return
	}

	defer resp.Body.Close()

	log.Info("Response status:", resp.Status)

	timer := time.Now()
	elapsed := timer.Sub(start)

	log.Info("Fetch syncdata %s\n", elapsed)

	remoteStatus := Status{}

	json.NewDecoder(resp.Body).Decode(&remoteStatus)

	timer = time.Now()
	elapsed = timer.Sub(start)

	if bytes.Compare(latestBlock.Key[:], remoteStatus.Hash) == 0 && (latestBlock.Value.Header.SeqID == remoteStatus.SeqID) {

		log.Info("GetLatestBlock matches from local to remote, skipping sync.")
	} else if remoteStatus.SeqID > latestBlock.Value.Header.SeqID {
		log.Info("Remote has larger SeqID (%d) vs local (%d) - Syncing\n", remoteStatus.SeqID, latestBlock.Value.Header.SeqID)
		p2p.stateSync(hostname, latestBlock.Key)

	} else if latestBlock.Value.Header.SeqID > remoteStatus.SeqID {
		log.Info("Local has larger SeqID (%d) vs remote (%d) - Skipping\n", latestBlock.Value.Header.SeqID, remoteStatus.SeqID)

	} else {
		log.Info("querySync => Unknown state. Local (%d) Remote (%d)", latestBlock.Value.Header.SeqID, remoteStatus.SeqID)

	}

}

func (p2p *P2P) stateSync(hostname string, hash blockdb.Hash) {

	start := time.Now()

	hashB64 := base64.StdEncoding.EncodeToString(hash[:])

	// TODO: Pass correct JSON RPC, not GET
	resp, err := http.Get(fmt.Sprintf("http://%s/p2p/sync?from=%s", hostname, url.QueryEscape(hashB64)))

	if err != nil {
		log.Warn("Error connecting for status => ", err)
		return
	}

	defer resp.Body.Close()

	log.Debug("Response status:", resp.StatusCode)

	syncBlocks := blockdb.SyncBlocks{}

	json.NewDecoder(resp.Body).Decode(&syncBlocks)

	timer := time.Now()
	elapsed := timer.Sub(start)

	log.Info("Prepare stateSync %s\n", elapsed)

	blockLen := len(syncBlocks.Blocks)

	for i := 0; i < len(syncBlocks.Blocks); i++ {

		start = time.Now()

		// TODO: Push to final signature state, vs each block

		/*
			for i2 := 0; i2 < len(syncBlocks.Blocks[i].Value.Payload); i2++ {


				i3 := i2
				i4 := i

				go func() {

					if len(syncBlocks.Blocks[i4].Value.Payload[i3].Sender) == 0 {
						fmt.Println("Empty block at => ", i3)
						//continue
					}

					// Confirm if validated
					mywallet := wallet.New()
					verify := mywallet.VerifyRaw(syncBlocks.Blocks[i4].Value.Payload[i3].Sender, syncBlocks.Blocks[i4].Value.Payload[i3].Data, syncBlocks.Blocks[i4].Value.Payload[i3].Signature)

					// If signed and verified, push to the stack
					if !verify {
						fmt.Printf("Signature failed -- Skipping block (%d) payload ID (%d)", i, i2)
						//return
					}

				}()


			}
		*/

		// Prepare the JSON to write to disk
		payload, err := json.Marshal(syncBlocks.Blocks[i])

		if err != nil {
			log.Fatal(err)
		}

		// Append the new block to disk
		err = p2p.POH.BlockDB.Append(p2p.POH.CreateBlock(payload, true))

		if err != nil {
			log.Fatal(err)
		}

		timer = time.Now()
		elapsed = timer.Sub(start)

		txRate := uint32(float64(blockLen) * (1 / elapsed.Seconds()))

		log.Info("Sync Blocks done in %s, %d per sec\n", elapsed, txRate)

	}

}

// JSON RPC methods

// Return the latest status (latestblock)
func (p2p *P2P) Status(c *gin.Context) {

	// Query our end-point
	// TODO: Validate rpc_host and source ip
	rpc_host, _ := c.GetQuery("rpc_host")
	rpc_port, _ := c.GetQuery("rpc_port")

	if rpc_host != "" && rpc_port != "" {

		// TODO: Confirm correct end-point
		rpc_addr := fmt.Sprintf("%s:%s", rpc_host, rpc_port)

		log.Debug("Adding RPC Addr => ", rpc_addr)

		if p2p.RPC_Peers[rpc_addr].Host == "" {
			// New host, append to our stack
			p2p.RPC_Peers[rpc_addr] = Node{Host: rpc_host}

		}

	}

	// Return the latest block in the stack
	latestBlock := p2p.POH.BlockDB.GetLatestBlock()

	status := Status{
		Parent: latestBlock.Value.Header.Parent[:],
		SeqID:  latestBlock.Value.Header.SeqID,
		Hash:   latestBlock.Key[:],

		P2P_Node:  p2p.P2P_Node,
		P2P_Peers: p2p.P2P_Peers,

		RPC_Node:  p2p.RPC_Node,
		RPC_Peers: p2p.RPC_Peers,
	}

	c.JSON(200, status)
}

// Sync the state from a specified block
func (p2p *P2P) Sync(c *gin.Context) {

	from, _ := c.GetQuery("from")

	fromBytes, _ := base64.StdEncoding.DecodeString(from)

	log.Debug("fromBytes => ", fromBytes)

	blockID := p2p.POH.BlockDB.Sync(fromBytes)

	// TODO: Replace with struct and pointer, vs manually walking array
	c.Data(200, "application/json; charset=utf-8", []byte(fmt.Sprintf("{\"public_key\": \"%s\", \"blocks\": [", base64.StdEncoding.EncodeToString(p2p.POH.Wallet.PublicKey))))

	if len(p2p.POH.BlockDB.Blocks) > 0 {

		// Print the first block
		c.JSON(200, &p2p.POH.BlockDB.Blocks[blockID])

		for i := blockID + 1; i < len(p2p.POH.BlockDB.Blocks); i++ {
			c.Data(200, "application/json; charset=utf-8", []byte(","))
			c.JSON(200, &p2p.POH.BlockDB.Blocks[i])
		}

	}

	c.Data(200, "application/json; charset=utf-8", []byte("]}"))

}
