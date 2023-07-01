package cmd

import (
	"fmt"

	"github.com/nbskp/binn-cli/client"
	"github.com/nbskp/binn-cli/config"
	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send [id] [token] [msg]",
	Short: "Send a bottle",
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.Load()
		if err != nil {
			fmt.Printf("failed to load config: %v\n", err)
			return
		}

		cli := client.NewClient(fmt.Sprintf("http://%s", c.Host))
		ok, err := cli.PostBottle(cmd.Context(), &client.Bottle{
			ID:    args[0],
			Token: args[1],
			Msg:   args[2],
		})
		if err != nil {
			fmt.Printf("failed to send a bottle: %v\n", err)
			return
		}
		if ok {
			fmt.Println("success")
		} else {
			fmt.Println("failed")
		}
	},
}
