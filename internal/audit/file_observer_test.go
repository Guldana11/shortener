package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestFileObserver_Notify(t *testing.T) {
	type fields struct {
		path string
	}
	type args struct {
		e Event
	}

	tmpFileSingle := "test_single.log"
	tmpFileMultiple := "test_multiple.log"
	invalidFile := "/invalid_dir/test.log"

	tests := []struct {
		name   string
		fields fields
		args   args
		verify func() error
	}{
		{
			name:   "Write single event",
			fields: fields{path: tmpFileSingle},
			args:   args{e: Event{TS: 1, Action: "login", UserID: "user1", URL: "/login"}},
			verify: func() error {
				defer os.Remove(tmpFileSingle)
				file, err := os.Open(tmpFileSingle)
				if err != nil {
					return err
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)
				if !scanner.Scan() {
					return os.ErrNotExist
				}

				var got Event
				if err := json.Unmarshal(scanner.Bytes(), &got); err != nil {
					return err
				}
				want := Event{TS: 1, Action: "login", UserID: "user1", URL: "/login"}
				if !reflect.DeepEqual(got, want) {
					return os.ErrInvalid
				}
				return nil
			},
		},
		{
			name:   "Append multiple events",
			fields: fields{path: tmpFileMultiple},
			args:   args{e: Event{TS: 2, Action: "click", UserID: "user2", URL: "/home"}},
			verify: func() error {
				defer os.Remove(tmpFileMultiple)
				f := NewFileObserver(tmpFileMultiple)
				events := []Event{
					{TS: 2, Action: "click", UserID: "user2", URL: "/home"},
					{TS: 3, Action: "view", UserID: "user3", URL: "/products"},
					{TS: 4, Action: "logout", UserID: "user4", URL: "/logout"},
				}

				for _, e := range events[1:] {
					f.Notify(e)
				}

				file, err := os.Open(tmpFileMultiple)
				if err != nil {
					return err
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)
				i := 0
				for scanner.Scan() {
					var got Event
					if err := json.Unmarshal(scanner.Bytes(), &got); err != nil {
						return err
					}
					if !reflect.DeepEqual(got, events[i]) {
						return os.ErrInvalid
					}
					i++
				}
				if i != len(events) {
					return os.ErrInvalid
				}
				return nil
			},
		},
		{
			name:   "Invalid path",
			fields: fields{path: invalidFile},
			args:   args{e: Event{TS: 5, Action: "fail", URL: "/fail"}},
			verify: func() error {
				if _, err := os.Stat(invalidFile); !os.IsNotExist(err) {
					return os.ErrExist
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileObserver{
				path: tt.fields.path,
			}
			f.Notify(tt.args.e)
			if tt.verify != nil {
				if err := tt.verify(); err != nil {
					t.Errorf("verification failed: %v", err)
				}
			}
		})
	}
}

func TestNewFileObserver(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want *FileObserver
	}{
		{
			name: "Basic creation",
			args: args{path: "dummy.log"},
			want: &FileObserver{path: "dummy.log"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFileObserver(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFileObserver() = %v, want %v", got, tt.want)
			}
		})
	}
}
