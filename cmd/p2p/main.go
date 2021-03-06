package main

import (
	"flag"
	"fmt"
	"os/user"
	"time"

	"github.com/perrychain/perry/pkg/p2pnet"
	"github.com/perrychain/perry/pkg/wallet"
	log "github.com/sirupsen/logrus"
)

func main() {

	//poh := poh_hash.New("")
	//p2p.POH = &poh

	var server = flag.Int("s", 0, "Launch server")
	var num = flag.Int("n", 1_000, "Number of packets to send")

	var p2p_host = flag.String("p2phost", "127.0.0.1", "P2P Host")
	var p2p_port = flag.Uint("p2pport", 16842, "P2P Port")

	usr, _ := user.Current()
	defaultHomeDir := fmt.Sprintf("%s/.perry", usr.HomeDir)
	defaultWalletPath := fmt.Sprintf("%s/.wallet.json", defaultHomeDir)

	var walletPath = flag.String("wallet", defaultWalletPath, "Specify wallet path")

	flag.Parse()

	p2p := p2pnet.New(p2pnet.P2P{
		P2P_Node: p2pnet.Node{
			Host: *p2p_host,
			Port: uint16(*p2p_port),
		},
	})

	if *server > 0 {

		log.Info("Launching server")

		go func() {
			p2p.Listen(p2p.MsgHandler)
		}()

		go func() {
			//poh.GeneratePOH(100_000_000_000)
		}()

		for {

		}

	} else {

		log.Info("Launching client")

		senderwallet, err := wallet.Load(*walletPath)

		if err != nil {
			log.Warn("No wallet specified, generating new one\n")
			senderwallet.GenerateWallet()
		}

		for i := 0; i < *num; i++ {
			p2p.Send(senderwallet, fmt.Sprintf("%d", i))
			time.Sleep(1 * time.Millisecond)
		}

	}

}
