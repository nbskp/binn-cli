package cmd

import (
	"fmt"
	"time"

	"github.com/nbskp/binn-cli/client"
	"github.com/nbskp/binn-cli/config"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get bottles",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.Load()
		if err != nil {
			fmt.Printf("failed to load config: %v\n", err)
			return
		}

		cli := client.NewClient(fmt.Sprintf("http://%s", c.Host))
		ch, errCh := cli.RunGetByText(cmd.Context())
		for {
			select {
			case b := <-ch:
				fmt.Printf("ID: %s, Msg: \"%s\", ExpiredAt: %s\n", b.ID, b.Msg, time.Unix(b.ExpiredAt, 0))
			case err := <-errCh:
				fmt.Println(err)
				return
			}
		}
	},
}
