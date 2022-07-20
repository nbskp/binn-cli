package client

import (
	"io"
	"fmt"
	"time"
	"bufio"
	"bytes"
	"strings"
	"net/http"
	"encoding/json"
)

type Message struct {
	Text string `json:"text"`
}

type requestBottle struct {
	ID        string     `json:"id"`
	Message   *Message   `json:"message"`
	ExpiredAt *time.Time `json:"expired_at"`
}

type responseBottle struct {
	ID        string     `json:"id"`
	Message   *Message   `json:"message"`
	ExpiredAt *time.Time `json:"expired_at"`
}

type Client struct {
	url        string
	httpClient *http.Client
}

func NewClient(url string, timeout time.Duration) *Client {
	return &Client{
		url:        url,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
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
			em.Event = strings.TrimRight(field[n + 7:], "\n")
		} else if n := strings.Index(field, "data: "); n >= 0 {
			em.Data = strings.TrimRight(field[n + 6:], "\n")
		}
	}

	return &em
}

func (c *Client) Get() (chan *responseBottle, chan error) {
	ch := make(chan *responseBottle)
	errCh := make(chan error)

	go func() {
		req, err := http.NewRequest("GET", c.url, nil)
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Connection", "keep-alive")
		if err != nil {
			errCh <- err
			return
		}
		
		resp, err := c.httpClient.Do(req)
		if err != nil {
			errCh <- err
			return
		}
		scanner := NewEventStreamScanner(resp.Body)

	Loop:
		for {
			select {
			case <-req.Context().Done():
				break Loop
			default:
				if scanner.Scan() {
					data := scanner.Bytes()
					em := ParseEventMessage(data)
					if em == nil {
						break
					}
					var res responseBottle
					if err := json.Unmarshal([]byte(em.Data), &res); err != nil {
						errCh <- err
						break
					}
					ch <- &res
					break
				}
				if err := scanner.Err(); err != nil {
					errCh <- err
					return
				}
				break
			}
		}
	}()
	
	return ch, errCh
}

func (c *Client) Post(id string, text string) error {
	rb := &requestBottle{
		ID:        id,
		Message:   &Message{
			Text: text,
		},
	}

	if byte_, err := json.Marshal(rb); err != nil {
		return fmt.Errorf("%w", err)
	} else {
		payload := bytes.NewBuffer(byte_)
		_, err := c.httpClient.Post(c.url, "application/json", payload)
		return err
	}
}
