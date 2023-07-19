package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Bottle struct {
	ID        string `json:"id"`
	Msg       string `json:"msg"`
	Token     string `json:"token"`
	ExpiredAt int64  `json:"expired_at"`
}

type Client struct {
	url        string
	httpClient *http.Client
}

func NewClient(url string) *Client {
	return &Client{
		url:        url,
		httpClient: &http.Client{},
	}
}

func buildGetStreamEndpointPath(base string) (string, error) {
	return url.JoinPath(base, "/bottles/stream")
}

func (cli *Client) RunGetByText(ctx context.Context) (chan *Bottle, chan error) {
	ch := make(chan *Bottle, 0)
	errCh := make(chan error, 0)
	go func() {
		p, err := buildGetStreamEndpointPath(cli.url)
		if err != nil {
			errCh <- fmt.Errorf("failed to build endpoint path: %w", err)
			return
		}
		req, err := http.NewRequest(http.MethodGet, p, nil)
		if err != nil {
			errCh <- fmt.Errorf("failed to build http request: %w", err)
			return
		}
		resp, err := cli.httpClient.Do(req)
		if err != nil {
			errCh <- fmt.Errorf("failed to request: %w", err)
			return
		}
		defer resp.Body.Close()
		dec := json.NewDecoder(resp.Body)
		for dec.More() {
			select {
			case <-ctx.Done():
				return
			default:
				var b *Bottle
				if err := dec.Decode(&b); err != nil {
					errCh <- fmt.Errorf("failed to decode json: %w", err)
					return
				}
				ch <- b
			}
		}
	}()
	return ch, errCh
}

func buildPostBottleEndpointPath(base string) (string, error) {
	return url.JoinPath(base, "/bottles/")
}

func (cli *Client) PostBottle(ctx context.Context, b *Bottle) (ok bool, err error) {
	p, err := buildPostBottleEndpointPath(cli.url)
	if err != nil {
		return false, err
	}
	data, err := json.Marshal(&b)
	if err != nil {
		return false, err
	}
	r := bytes.NewReader(data)
	req, err := http.NewRequest(http.MethodPost, p, r)
	if err != nil {
		return false, err
	}
	resp, err := cli.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, nil
	}
	return true, nil
}

func NewEventStreamScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024), 1024)
	split := func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.Index(data, []byte("\n\n")); i >= 0 {
			return i + 2, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	}
	scanner.Split(split)

	return scanner
}

type EventMessage struct {
	Event string
	Data  string
}

func ParseEventMessage(b []byte) *EventMessage {
	if len(b) == 0 {
		return nil
	}

	fields := strings.Split(string(b), "\n")
	var em EventMessage
	for _, field := range fields {
		if n := strings.Index(field, "event: "); n >= 0 {
			em.Event = strings.TrimRight(field[n+7:], "\n")
		} else if n := strings.Index(field, "data: "); n >= 0 {
			em.Data = strings.TrimRight(field[n+6:], "\n")
		}
	}

	return &em
}
