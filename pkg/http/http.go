package http

import (
	"context"
	"fmt"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/perrychain/perry/pkg/p2pnet"
	"github.com/perrychain/perry/pkg/poh_hash"
)

type HTTP struct {
	P2P_Node   p2pnet.Node
	RPC_Node   p2pnet.Node
	WalletPath string
	DBPath     string
}

func New(h HTTP) HTTP {

	return h

}

// Server the JSON RPC endpoint
func (http HTTP) Serve() {

	router := gin.Default()

	// Enable gzip compression by default
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	poh := poh_hash.New(http.WalletPath, http.DBPath)

	p2p := p2pnet.New(p2pnet.P2P{
		RPC_Node: p2pnet.Node{
			Host: http.RPC_Node.Host,
			Port: http.RPC_Node.Port,
		},

		P2P_Node: p2pnet.Node{
			Host: http.P2P_Node.Host,
			Port: http.P2P_Node.Port,
		},
	})

	p2p.POH = &poh

	// Launch the UDP packet receiver
	go func() {
		p2p.Listen(p2p.MsgHandler)
	}()

	// Launch peer query
	go func() {

		p2p.QueryPeers(context.Background())

	}()

	// Launch the PoH go routine
	go func() {
		// TODO: Loop forever
		poh.GeneratePOH(100_000_000_000)
	}()

	// TODO: Use JSON RPC style endpoints
	router.GET("/sync", poh.Syncstate)

	router.GET("/syncdata", poh.Syncdatastate)

	router.GET("/verify", poh.Verify)

	router.GET("/push", poh.Pushstate)

	router.GET("/state", poh.State)

	router.GET("/", poh.Index)

	// p2p state
	router.GET("/p2p/status", p2p.Status)

	// Sync state from specified point
	router.GET("/p2p/sync", p2p.Sync)

	// TODO: Add P2P support
	//router.GET("/p2p/nodes", poh.state)
	//router.GET("/p2p/status", poh.state)

	router.Run(fmt.Sprintf("%s:%d", http.RPC_Node.Host, http.RPC_Node.Port))

}
