package audit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHTTPObserver_Notify(t *testing.T) {
	type fields struct {
		url string
	}
	type args struct {
		e Event
	}

	var received Event
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name   string
		fields fields
		args   args
		want   Event
	}{
		{
			name:   "Send event",
			fields: fields{url: server.URL},
			args:   args{e: Event{TS: 1, Action: "login", UserID: "user1", URL: "/login"}},
			want:   Event{TS: 1, Action: "login", UserID: "user1", URL: "/login"},
		},
		{
			name:   "Send event without user_id",
			fields: fields{url: server.URL},
			args:   args{e: Event{TS: 2, Action: "view", URL: "/products"}},
			want:   Event{TS: 2, Action: "view", URL: "/products"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			received = Event{}
			h := &HTTPObserver{
				url: tt.fields.url,
			}
			h.Notify(tt.args.e)

			if !reflect.DeepEqual(received, tt.want) {
				t.Errorf("Notify sent wrong event, got %v, want %v", received, tt.want)
			}
		})
	}
}

func TestNewHTTPObserver(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want *HTTPObserver
	}{
		{
			name: "Basic creation",
			args: args{url: "http://example.com"},
			want: &HTTPObserver{url: "http://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHTTPObserver(tt.args.url); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHTTPObserver() = %v, want %v", got, tt.want)
			}
		})
	}
}
