package service

import (
	"net/http"
	"strings"
	"testing"
)

func TestGenerateUserCookie(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Generate valid user cookie"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateUserCookie()
			if got == nil {
				t.Fatal("GenerateUserCookie() returned nil") // безопасно выходим, если nil
			}

			if got.Name != cookieName {
				t.Errorf("GenerateUserCookie() Name = %v, want %v", got.Name, cookieName)
			}

			parts := strings.Split(got.Value, "|")
			if len(parts) != 2 {
				t.Errorf("GenerateUserCookie() Value format invalid, got: %v", got.Value)
			}
			if len(parts[0]) == 0 {
				t.Error("GenerateUserCookie() userID part is empty")
			}
			if len(parts[1]) == 0 {
				t.Error("GenerateUserCookie() signature part is empty")
			}
		})
	}
}

func TestGenerateUserCookieSafety(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Generate another valid cookie"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateUserCookie()
			if got == nil {
				t.Fatal("GenerateUserCookie() returned nil")
			}

			if got.Value == "" {
				t.Error("GenerateUserCookie() Value is empty")
			}

			parts := strings.Split(got.Value, "|")
			if len(parts) != 2 {
				t.Errorf("GenerateUserCookie() Value format invalid, got: %v", got.Value)
			}
		})
	}
}

func TestValidateUserCookie(t *testing.T) {
	cookie := GenerateUserCookie()
	req, _ := http.NewRequest("GET", "/", nil)
	req.AddCookie(cookie)

	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Valid cookie",
			args:    args{r: req},
			want:    strings.Split(cookie.Value, "|")[0],
			wantErr: false,
		},
		{
			name: "Invalid cookie",
			args: args{
				r: func() *http.Request {
					req, _ := http.NewRequest("GET", "/", nil)
					req.AddCookie(&http.Cookie{Name: cookieName, Value: "invalid"})
					return req
				}(),
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateUserCookie(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUserCookie() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateUserCookie() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sign(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Sign test",
			args: args{data: "testdata"},
			want: sign("testdata"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sign(tt.args.data); got != tt.want {
				t.Errorf("sign() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_verify(t *testing.T) {
	data := "testdata"
	sig := sign(data)

	type args struct {
		data string
		sig  string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Correct signature",
			args: args{data: data, sig: sig},
			want: true,
		},
		{
			name: "Incorrect signature",
			args: args{data: data, sig: sig + "wrong"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := verify(tt.args.data, tt.args.sig); got != tt.want {
				t.Errorf("verify() = %v, want %v", got, tt.want)
			}
		})
	}
}
