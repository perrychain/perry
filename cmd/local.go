package cmd

import (
	"fmt"
	"runtime"

	"github.com/perrychain/perry/pkg/poh_hash"
	"github.com/spf13/cobra"
)

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Benchmark and simulate a local Perry instance using proof-of-history",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		walletPath, _ := cmd.Flags().GetString(walletLocation)
		dbPath, _ := cmd.Flags().GetString(dbLocation)

		fmt.Println("Proof of History example")

		cpu_cores := runtime.NumCPU()
		fmt.Printf("CPU Cores: %d\n", cpu_cores)

		poh := poh_hash.New(walletPath, dbPath)
		poh.GeneratePOH(100_000_000)

		fmt.Printf("Generate Hashrate %d p/sec (1-core)\n", poh.HashRate)

		poh.VerifyPOH(cpu_cores)

		fmt.Printf("Verify Hashrate %d p/sec (%d-cores)\n", poh.VerifyHashRate, cpu_cores)
		fmt.Printf("Verify Hashrate %d p/core\n", poh.VerifyHashRatePerCore)

	},
}

func init() {
	rootCmd.AddCommand(localCmd)

}
