# slackreports

This is a basic tool to send at most one message every X seconds on a Slack channel using a webhook, useful to reduce the volume of messages when only the latest message is important. Only the latest message set by the `Update` method will be sent. Use the `Stop` method release associated resources.

## Example

```go
package main

import (
	"log"
	"time"

	"github.com/clement-fischer/slackreports"
)

func main() {
	// Replace using your previously configured webhook
	slackHook := "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"

	r := slackreports.NewReporter(slackHook, 3*time.Second)
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
			message := ts.String()
			if err := r.Update(message); err != nil {
				log.Println(err)
			}
		}
	}
}
```
