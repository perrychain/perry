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
	Short: "View a specified wallet public key",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		walletPath, _ := cmd.Flags().GetString("wallet")

		mywallet, _ := wallet.Load(walletPath)

		fmt.Printf("%s\n", base64.StdEncoding.EncodeToString(mywallet.PublicKey))

	},
}

func init() {
	keygenCmd.AddCommand(pubkeyCmd)

}
