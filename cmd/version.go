package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Perry version details",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(PerryVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

}
