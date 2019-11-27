package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type slackMessage struct {
	Text string `json:"text"`
}

func sendSlackMessage(message, url string) error {
	// Slack escapes
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	message = replacer.Replace(message)

	payload := &slackMessage{Text: message}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Slack returned HTTP status %v", resp.StatusCode)
	}

	return nil
}

func runUpdater(webhook string, c <-chan string, done <-chan bool, ticker *time.Ticker, e *error) {
	var lastReport string
	var sendReport bool
	for {
		select {
		case <-ticker.C:
			if sendReport {
				if err := sendSlackMessage(lastReport, webhook); err != nil {
					*e = err
					log.Printf("error: could not send message to Slack: %v", err)
				} else {
					sendReport = false
				}
			}
		case newReport := <-c:
			if lastReport != newReport {
				lastReport = newReport
				sendReport = true
			}
		case <-done:
			return
		}
	}
}

type Reporter struct {
	c      chan<- string
	done   chan<- bool
	E      error
	ticker *time.Ticker
}

// Update sets the report to be updated next. If the runUpdater goroutine is not able to receive the update, the message is dropped so this never blocks.
func (r *Reporter) Update(message string) error {
	select {
	case r.c <- message:
		return nil
	default:
		return fmt.Errorf("warning: Reporter buffer is full")
	}
}

// Stop releases associated resources
func (r *Reporter) Stop() {
	r.ticker.Stop()
	r.done <- true
}

// NewReporter returns a new instance which will trigger the webhook periodically after the first call to Update.
func NewReporter(webhook string, d time.Duration) *Reporter {
	c := make(chan string, 1)
	done := make(chan bool)
	ticker := time.NewTicker(d)
	r := &Reporter{c: c, done: done, ticker: ticker}
	go runUpdater(webhook, c, done, ticker, &r.E)
	return r
}

func main() {
	// Replace using your previously configured webhook
	slackHook := "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"

	r := NewReporter(slackHook, 3*time.Second)
	defer r.Stop()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	done := make(chan bool)
	go func() {
		time.Sleep(10 * time.Second)
		done <- true
	}()

	for {
		if r.E != nil {
			log.Fatalf("error: slackUpdater error field is not nil: %v", r.E)
		}
		select {
		case <-done:
			return
		case ts := <-ticker.C:
			message := "[DEBUG] Running Unit Test: " + ts.String()
			if err := r.Update(message); err != nil {
				fmt.Println(err)
			}
		}
	}
}
