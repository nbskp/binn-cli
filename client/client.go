package client

import (
	"fmt"
	"time"
	"bytes"
	"net/http"
	"io/ioutil"
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

func (c *Client) Get() (*responseBottle, error) {
	resp, err := c.httpClient.Get(c.url)
	if err != nil {
		return nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var res responseBottle
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("%s (%s)", err, body)
	}
	
	return &res, nil
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
