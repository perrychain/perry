package cmd

import (
	"encoding/base64"
	"fmt"

	"github.com/perrychain/perry/pkg/wallet"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new wallet",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		force, _ := cmd.Flags().GetBool("force")
		walletPath, _ := cmd.Flags().GetString("wallet")
		name, _ := cmd.Flags().GetString("name")

		if walletPath == "" {
			walletPath = "~/.perry-wallet.json"
		}

		mywallet := wallet.New()
		mywallet.Name = name

		err := mywallet.GenerateWallet()

		if err != nil {
			log.Fatal(err)
		}

		err = mywallet.Save(walletPath, force)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("New wallet saved in %s\n", walletPath)
		fmt.Println("Your wallet public key is:")
		fmt.Printf("%s\n", base64.StdEncoding.Strict().EncodeToString(mywallet.PublicKey))

	},
}

func init() {
	keygenCmd.AddCommand(newCmd)

	newCmd.Flags().BoolP("force", "f", false, "Force overwrite")
	newCmd.Flags().StringP("name", "n", "My Wallet", "Wallet identifier name (My wallet)")

}
