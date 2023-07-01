package cmd

import (
	"fmt"

	"github.com/nbskp/binn-cli/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [host]",
	Short: "initialize settings",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := config.Config{Host: args[0]}
		if err := c.Save(); err != nil {
			fmt.Printf("failed to initialize config file: %v\n", err)
		}
	},
}
