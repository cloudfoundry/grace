package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	interval = time.Second
)

type TimeStream struct {
}

func (t *TimeStream) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	notifyCh := w.(http.CloseNotifier).CloseNotify()
	flusher := w.(http.Flusher)

	timer := time.NewTimer(0)

	for {
		select {
		case <-notifyCh:
			fmt.Fprintf(os.Stderr, "time-stream: client closed connection")
			return

		case t := <-timer.C:
			ts := t.UTC().Format(time.RFC3339Nano)
			fmt.Fprintf(os.Stderr, "time-stream: tick "+ts+"\n")
			w.Write([]byte(ts + "\n"))
			flusher.Flush()
			timer.Reset(interval)
		}
	}
}
