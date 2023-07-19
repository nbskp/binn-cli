package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/nbskp/binn-cli/client"
	"github.com/nbskp/binn-cli/config"
	"github.com/spf13/cobra"
	"nhooyr.io/websocket"
)

func getCmd() *cobra.Command {
	ws := false
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get bottles",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := config.Load()
			if err != nil {
				fmt.Printf("failed to load config: %v\n", err)
				return
			}

			if ws {
				handleWebsocket(cmd, c)
				return
			}

			handleStream(cmd, c)
		},
	}
	cmd.PersistentFlags().BoolVarP(&ws, "ws", "", false, "use websocket")
	return cmd
}

func handleStream(cmd *cobra.Command, c *config.Config) {
	cli := client.NewClient(fmt.Sprintf("http://%s", c.Host))
	ch, errCh := cli.RunGetByText(cmd.Context())
	for {
		select {
		case b := <-ch:
			printBottle(b)
		case err := <-errCh:
			fmt.Println(err)
			return
		}
	}
}

func handleWebsocket(cmd *cobra.Command, c *config.Config) {
	ws, _, err := websocket.Dial(cmd.Context(), fmt.Sprintf("ws://%s/bottles/ws", c.Host), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close(websocket.StatusInternalError, "disconnected")

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	go func() {
		select {
		case <-ch:
			ws.Close(websocket.StatusNormalClosure, "disconnected")
			return
		}
	}()

	for {
		_, rd, err := ws.Reader(cmd.Context())
		if err != nil {
			fmt.Println(err)
			return
		}
		dec := json.NewDecoder(rd)
		for dec.More() {
			var b client.Bottle
			if err := dec.Decode(&b); err == io.EOF {
				break
			} else if err != nil {
				fmt.Println(err)
				return
			}
			printBottle(&b)
		}
		//ws.CloseRead(cmd.Context())
	}
}

func printBottle(b *client.Bottle) {
	fmt.Printf("ID: %s, Msg: \"%s\", ExpiredAt: %s\n", b.ID, b.Msg, time.Unix(b.ExpiredAt, 0))
}
