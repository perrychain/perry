package cmd

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/perrychain/perry/pkg/wallet"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new wallet",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		force, _ := cmd.Flags().GetBool("force")
		path, _ := cmd.Flags().GetString("path")
		name, _ := cmd.Flags().GetString("name")

		if path == "" {
			path = ".wallet.json"
		}

		mywallet := wallet.New()
		mywallet.Name = name

		err := mywallet.GenerateWallet()

		if err != nil {
			log.Fatal(err)
		}

		err = mywallet.Save(path, force)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("New wallet saved in %s\n", path)
		fmt.Println("Your wallet public key is:")
		fmt.Printf("%s\n", base64.StdEncoding.Strict().EncodeToString(mywallet.PublicKey))

	},
}

func init() {
	keygenCmd.AddCommand(newCmd)

	newCmd.Flags().BoolP("force", "f", false, "Force overwrite")
	newCmd.Flags().StringP("path", "p", ".wallet.json", "Wallet filename")
	newCmd.Flags().StringP("name", "n", "My Wallet", "Wallet identifier name (My wallet)")

}
