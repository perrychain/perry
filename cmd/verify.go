package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("verify called")
	},
}

func init() {
	keygenCmd.AddCommand(verifyCmd)

}
