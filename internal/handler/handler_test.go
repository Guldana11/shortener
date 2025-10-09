package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestNewURLHandler(t *testing.T) {
	tests := []struct {
		name string
		want *URLHandler
	}{
		{
			name: "Успешное создание хэндлера",
			want: &URLHandler{
				store: map[string]string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewURLHandler()
			if got == nil {
				t.Fatal("NewURLHandler() вернул nil, ожидался валидный объект")
			}

			if got.store == nil {
				t.Error("store должен быть инициализирован, но равен nil")
			}

			if len(got.store) != 0 {
				t.Errorf("ожидался пустой store, получили длину %d", len(got.store))
			}

			got2 := NewURLHandler()
			if &got.store == &got2.store {
				t.Error("два экземпляра NewURLHandler() ссылаются на одну и ту же map")
			}

			if !reflect.DeepEqual(got.store, tt.want.store) {
				t.Errorf("store не совпадает: got %v, want %v", got.store, tt.want.store)
			}

		})
	}
}

func TestURLHandler_GetHandler(t *testing.T) {
	type fields struct {
		store map[string]string
		mu    sync.Mutex
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		expectedStatus   int
		expectedBody     string
		expectedLocation string
	}{
		{
			name: "Существующий короткий URL",
			fields: fields{
				store: map[string]string{
					"id": "https://example.com",
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodGet, "/id", nil),
			},
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://example.com",
		},
		{
			name: "Несуществующий короткий URL",
			fields: fields{
				store: make(map[string]string),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodGet, "/id", nil),
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "bad request",
		},
		{
			name: "Пустой id в URL",
			fields: fields{
				store: make(map[string]string),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodGet, "/", nil),
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "bad request",
		},
		{
			name: "Неверный формат пути",
			fields: fields{
				store: make(map[string]string),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodGet, "/abc/extra", nil),
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "bad request",
		},
		{
			name: "Параллельный доступ",
			fields: fields{
				store: map[string]string{
					"id1": "https://example.com/1",
					"id2": "https://example.com/2",
				},
			},
			expectedStatus: http.StatusTemporaryRedirect,
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			h := &URLHandler{
				store: tt.fields.store,
				mu:    tt.fields.mu,
			}

			if tt.args.r != nil && tt.args.w != nil {
				h.GetHandler(tt.args.w, tt.args.r)
				rec := tt.args.w.(*httptest.ResponseRecorder)
				res := rec.Result()
				defer res.Body.Close()

				if res.StatusCode != tt.expectedStatus {
					t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
				}

				if tt.expectedLocation != "" {
					loc := res.Header.Get("Location")
					if loc != tt.expectedLocation {
						t.Errorf("expected Location header %q, got %q", tt.expectedLocation, loc)
					}
				}

				if tt.expectedBody != "" {
					body, _ := io.ReadAll(res.Body)
					if !strings.Contains(string(body), tt.expectedBody) {
						t.Errorf("expected body to contain %q, got %q", tt.expectedBody, string(body))
					}
				}
			}

			if tt.name == "Параллельный доступ" {
				var wg sync.WaitGroup
				for j := 0; j < 5; j++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						w := httptest.NewRecorder()
						r := httptest.NewRequest(http.MethodGet, "/id1", nil)
						h.GetHandler(w, r)
						res := w.Result()
						defer res.Body.Close()
						if res.StatusCode != tt.expectedStatus && res.StatusCode != http.StatusBadRequest {
							t.Errorf("unexpected status: %d", res.StatusCode)
						}
					}()
				}
				wg.Wait()
			}
		})
	}

}

func TestURLHandler_PostHandler(t *testing.T) {
	type fields struct {
		store map[string]string
		mu    sync.Mutex
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "действительный запрос POST",
			fields: fields{
				store: make(map[string]string),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://example.com")),
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/",
		},
		{
			name: "недопустимый метод GET",
			fields: fields{
				store: make(map[string]string),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodGet, "/", nil),
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "bad request",
		},
		{
			name: "пустое тело",
			fields: fields{
				store: make(map[string]string),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("")),
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "bad request",
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			h := &URLHandler{
				store: tt.fields.store,
				mu:    tt.fields.mu,
			}
			h.PostHandler(tt.args.w, tt.args.r)

			rec := tt.args.w.(*httptest.ResponseRecorder)
			res := rec.Result()
			defer res.Body.Close()

			body, _ := io.ReadAll(res.Body)

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}
			if !strings.Contains(string(body), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, string(body))
			}
		})
	}

}

func Test_generateID(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Генерация случайного ID — корректная длина и символы",
		},
		{
			name: "Генерация разных ID — уникальность результата",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "Генерация случайного ID — корректная длина и символы":
				id := generateID()

				if len(id) != 8 {
					t.Errorf("длина ID = %d, ожидалось 8", len(id))
				}

				for _, r := range id {
					if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
						t.Errorf("найден недопустимый символ: %q", r)
					}
				}

			case "Генерация разных ID — уникальность результата":
				ids := make(map[string]bool)
				for i := 0; i < 1000; i++ {
					id := generateID()
					if ids[id] {
						t.Errorf("найден повторяющийся ID: %s", id)
					}
					ids[id] = true
				}
			}
		})
	}
}
