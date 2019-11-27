package slack

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

type confFile struct {
	WebhookURL string
}

func getWebhookURL() (string, error) {
	file, err := os.Open("conf.json")
	if err != nil {
		return "", err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)

	cf := &confFile{}
	err = decoder.Decode(cf)
	if err != nil {
		return "", err
	}

	return cf.WebhookURL, nil
}

func TestSendMessage(t *testing.T) {
	slackHook, err := getWebhookURL()
	if err != nil {
		t.Fatalf("error: could not load webhook URL from config file: %v", err)
	}
	if err := sendSlackMessage("[DEBUG] Running Unit Test", slackHook); err != nil {
		t.Fatalf("error: sending message to Slack failed: %v", err)
	}
}

func TestReporter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	slackHook, err := getWebhookURL()
	if err != nil {
		t.Fatalf("error: could not load webhook URL from config file: %v", err)
	}

	r := NewReporter(slackHook, 3*time.Second)
	defer r.Stop()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	done := make(chan bool)
	go func() {
		time.Sleep(4 * time.Second)
		done <- true
	}()

	for {
		if r.E != nil {
			t.Fatalf("error: slackUpdater error field is not nil: %v", r.E)
		}
		select {
		case <-done:
			return
		case ts := <-ticker.C:
			message := "[DEBUG] Running Unit Test: " + ts.String()
			if err := r.Update(message); err != nil {
				t.Error(err)
			}
		}
	}
}
