package cmd

import (
	"encoding/base64"
	"fmt"

	"github.com/perrychain/perry/pkg/wallet"
	"github.com/spf13/cobra"
)

// pubkeyCmd represents the pubkey command
var pubkeyCmd = &cobra.Command{
	Use:   "pubkey",
	Short: "View a wallets public key",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")

		mywallet, _ := wallet.Load(path)

		fmt.Printf("%s\n", base64.StdEncoding.EncodeToString(mywallet.PublicKey))

	},
}

func init() {
	keygenCmd.AddCommand(pubkeyCmd)

	pubkeyCmd.Flags().StringP("path", "p", ".wallet.json", "Wallet filename")
}
