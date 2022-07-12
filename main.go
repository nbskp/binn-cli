package main

import (
	"os"
	"fmt"
	"flag"
	"time"
	"bufio"
	"strings"

	"github.com/binn-client/client"
)

const (
	GET_COMMAND = "get"
	POST_COMMAND = "post"
)

func getHandler(cli *client.Client) {
	ch, errCh := cli.Get()
	for {
		select {
		case rb := <-ch:
			fmt.Printf("==================================================================\n")
			fmt.Printf("id        : %s\n", rb.ID)
			fmt.Printf("message   : \n%s\n", rb.Message.Text)
			fmt.Printf("expired_at: %s\n", rb.ExpiredAt.Format("Mon, 02 Jan 2006 15:04:05"))
			break
		case err := <-errCh:
			fmt.Printf("%s\n", err)
		default:
			break
		}
	}
}

func postHandler(cli *client.Client, id string) {
	lines := []string{}
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		if err := s.Err(); err != nil {
			fmt.Printf("%s", err);
		}
		lines = append(lines, s.Text())
	}

	message := strings.Join(lines, "\n")
	if err := cli.Post(id, message); err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf("success\n")
	}
}

func helpText() string {
	return `
Usage:

     binn get <endpoint>
     binn post <endpoint> <id>

`
}

func main() {
	flag.Parse()
	
	if flag.NArg() < 2 {
		fmt.Printf(helpText())
		return
	}

	args := flag.Args()

	cli := client.NewClient(args[1], time.Duration(24) * time.Hour)

	switch args[0] {
	case GET_COMMAND:
		getHandler(cli)
	case POST_COMMAND:
		if flag.NArg() < 3 {
			fmt.Printf(helpText())
			return
		}
		postHandler(cli, args[2])
	default:
		fmt.Printf(helpText())
		return
	}
}
