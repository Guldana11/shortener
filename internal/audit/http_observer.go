package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type HTTPObserver struct {
	url string
}

func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{url: url}
}

func (h *HTTPObserver) Notify(e Event) {
	data, _ := json.Marshal(e)

	resp, err := http.Post(h.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
