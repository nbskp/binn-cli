package client

import (
	"testing"
	"net/http"
	"net/http/httptest"
	
	"github.com/stretchr/testify/assert"
)

func mockPostHandlerFunc() http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {}
}

func mockGetHandlerFunc() http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"1c7a8201-cdf7-11ec-a9b3-0242ac110004","message":{"text":"a test message"},"expiredAt":0}`))
	}
}

func TestPostBottle(t *testing.T) {
	ts := httptest.NewServer(mockPostHandlerFunc())
	defer ts.Close()
	cli := NewClient(ts.URL, 0)
	err := cli.Post("1c7a8201-cdf7-11ec-a9b3-0242ac110004", "a test message")

	assert.Nil(t, err)
}

func TestGetBottle(t *testing.T) {
	ts := httptest.NewServer(mockGetHandlerFunc())
	defer ts.Close()
	cli := NewClient(ts.URL, 0)
	b, _ := cli.Get()

	assert.Equal(t, "a test message", b.Message.Text)
	assert.Equal(t, "1c7a8201-cdf7-11ec-a9b3-0242ac110004", b.ID)
}
