package client

import (
	"time"
	"strings"
	"testing"
	"net/http"
	"encoding/json"
	"net/http/httptest"
	
	"github.com/stretchr/testify/assert"
)

func mockPostHandlerFunc() http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {}
}

func mockGetHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")

		unfilledTicker := time.NewTicker(time.Duration(10) * time.Millisecond)
		filledTicker := time.NewTicker(time.Duration(100) * time.Millisecond)

		flusher, ok := w.(http.Flusher)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	Loop:
		for {
			select {
			case <- r.Context().Done():
				break Loop
			case _ = <- unfilledTicker.C:
				if _, err := w.Write([]byte("\n\n")); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					break Loop
				}
				flusher.Flush()
				break
			case _ = <- filledTicker.C:
				if _, err := w.Write([]byte("event: bottle\ndata: {\"id\":\"1c7a8201-cdf7-11ec-a9b3-0242ac110004\",\"message\":{\"text\":\"a test message\"},\"expired_at\":\"2020-07-07T14:04:23Z\"}\n\n")); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					break Loop
				}
				flusher.Flush()
				break
			default:
				break
			}
		}
		return
	}
}

func TestPostBottle(t *testing.T) {
	ts := httptest.NewServer(mockPostHandlerFunc())
	defer ts.Close()
	cli := NewClient(ts.URL, 0)
	err := cli.Post("1c7a8201-cdf7-11ec-a9b3-0242ac110004", "a test message")

	assert.Nil(t, err)
}

func TestEventStreamScanner(t *testing.T) {
	eventStreamReader := strings.NewReader("event: bottle\ndata: {\"id\":\"1c7a8201-cdf7-11ec-a9b3-0242ac110004\",\"message\":{\"text\":\"a test message\"}}\n\n")
	scanner := NewEventStreamScanner(eventStreamReader)
	if scanner.Scan() {
		eventMessage := ParseEventMessage(scanner.Bytes())
		var b responseBottle
		if err := json.Unmarshal([]byte(eventMessage.Data), &b); err != nil {
			assert.Fail(t, "failed parse json")
			return
		}
		assert.Equal(t, "1c7a8201-cdf7-11ec-a9b3-0242ac110004", b.ID)
		assert.Equal(t, "a test message", b.Message.Text)
	} else {
		assert.Fail(t, "failed scan")
	}
}

func TestGetBottle(t *testing.T) {
	ts := httptest.NewServer(mockGetHandlerFunc())
	defer ts.Close()

	cli := NewClient(ts.URL, time.Duration(200) * time.Millisecond)
	ch, errCh := cli.Get()

	var b *responseBottle
Loop:
	for {
		select {
		case err := <-errCh:
			assert.Failf(t, "failed", "%w", err)
			break Loop
		case b = <-ch:
			assert.Equal(t, "a test message", b.Message.Text)
			assert.Equal(t, "1c7a8201-cdf7-11ec-a9b3-0242ac110004", b.ID)
			break Loop
		default:
			break
		}
	}
}
