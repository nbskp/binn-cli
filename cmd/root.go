package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "binn-cli",
	Short: "binn-cli is a client for binn",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	rootCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(getCmd())
	rootCmd.AddCommand(initCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
