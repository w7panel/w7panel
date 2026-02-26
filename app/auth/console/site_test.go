// nolint
package console

import (
	"net/http"
	"testing"
	"time"
)

func TestSite_checkTLSHandshake(t *testing.T) {
	// Setup test TLS server

	tests := []struct {
		name    string
		timeout time.Duration
		want    bool
	}{
		{
			name:    "successful handshake",
			timeout: 10 * time.Second,
			want:    true,
		},
		{
			name:    "timeout",
			timeout: 1 * time.Nanosecond,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Site{}
			oldTimeout := sitero.Host
			sitero.Host = "cai.fan.b2.sz.w7.com"
			defer func() { sitero.Host = oldTimeout }()

			// Override default timeout for test case
			if tt.timeout > 0 {
				oldTimeout := http.DefaultClient.Timeout
				http.DefaultClient.Timeout = tt.timeout
				defer func() { http.DefaultClient.Timeout = oldTimeout }()
			}

			if got := s.checkTLSHandshake(); got != tt.want {
				t.Errorf("checkTLSHandshake() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSite_checkTLSHandshake_Failure(t *testing.T) {
	// Set invalid host
	oldHost := sitero.Host
	sitero.Host = "invalid.host"
	defer func() { sitero.Host = oldHost }()

	s := Site{}
	if got := s.checkTLSHandshake(); got != false {
		t.Errorf("checkTLSHandshake() = %v, want false", got)
	}
}
