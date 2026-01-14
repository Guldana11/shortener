package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type HttpObserver struct {
	url string
}

func NewHttpObserver(url string) *HttpObserver {
	return &HttpObserver{url: url}
}

func (h *HttpObserver) Notify(e Event) {
	data, _ := json.Marshal(e)
	_, _ = http.Post(h.url, "application/json", bytes.NewBuffer(data))
}
