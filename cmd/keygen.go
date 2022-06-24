package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// keygenCmd represents the keygen command
var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate keygen pair",
	Long:  `Generate a public/private keypair chain`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Require `new` or `verify` argument")
	},
}

func init() {

	rootCmd.AddCommand(keygenCmd)
}
