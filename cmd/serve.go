package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/perrychain/perry/pkg/http"
	"github.com/perrychain/perry/pkg/p2pnet"
	"github.com/spf13/cobra"
)

const walletLocation = "path"
const dbLocation = "db"
const rpcIP = "rpcip"
const rpcPort = "rpcport"
const p2pIP = "p2pip"
const p2pPort = "p2pport"

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launch an instance of the Perry messaging blockchain",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		rpc_ip, _ := cmd.Flags().GetString(rpcIP)
		rpc_port, _ := cmd.Flags().GetUint16(rpcPort)
		p2p_ip, _ := cmd.Flags().GetString(p2pIP)
		p2p_port, _ := cmd.Flags().GetUint16(p2pPort)

		walletPath, _ := cmd.Flags().GetString(walletLocation)
		dbPath, _ := cmd.Flags().GetString(dbLocation)

		// Check wallet path
		if _, err := os.Stat(walletPath); err != nil {
			log.Fatalf("Wallet %s could not be opened (%s)", walletPath, err)
		}

		fmt.Printf("Launching RPC service on %s:%d\n", rpc_ip, rpc_port)

		rpc_node := p2pnet.Node{Port: rpc_port, Host: rpc_ip}
		p2p_node := p2pnet.Node{Port: p2p_port, Host: p2p_ip}

		http := http.New(http.HTTP{
			RPC_Node:   rpc_node,
			P2P_Node:   p2p_node,
			WalletPath: walletPath,
			DBPath:     dbPath,
		})

		http.Serve()

	},
}

func init() {

	rootCmd.PersistentFlags().String(dbLocation, ".blockchain-db.json", "Filename for blockchain DB")
	rootCmd.PersistentFlags().String(walletLocation, ".wallet.json", "Filename for wallet location to launch peer")
	rootCmd.PersistentFlags().String(rpcIP, "127.0.0.1", "exposed IP for communication with RPC peers")
	rootCmd.PersistentFlags().Uint16(rpcPort, 24816, "exposed HTTP port for communication with RPC peers")

	// 224.0.0.1 for all nodes, broadcast/multicast test
	rootCmd.PersistentFlags().String(p2pIP, "127.0.0.1", "exposed IP for communication with P2P peers")
	rootCmd.PersistentFlags().Uint16(p2pPort, 16842, "exposed HTTP port for communication with P2P peers")

	rootCmd.AddCommand(serveCmd)

}
