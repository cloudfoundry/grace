package handlers

import (
	"net/http"
	"time"
)

type sleep struct {
}

func (p *sleep) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	durationStr := r.URL.Query().Get(":duration")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	time.Sleep(duration)
	w.WriteHeader(http.StatusOK)
}
