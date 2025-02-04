package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var (
	Version    string
	CommitHash string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version",
	Long:  "Print pricefeeder version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nCommit hash: %s\n", Version, CommitHash)
	},
}
