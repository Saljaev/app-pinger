package verifier

import (
	"app-pinger/backend/internal/api/utilapi"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVerifier_Verify(t *testing.T) {
	type limiterForTest struct {
		countReq int
		timeWait time.Duration
	}

	type fieldsVerifier struct {
		apiKey string
		rL     int
		rT     time.Duration
	}

	tests := []struct {
		name        string
		verifier    fieldsVerifier
		apiKey      string
		limiterTest limiterForTest
		want        interface{}
	}{
		{
			name: "Valid request",
			verifier: fieldsVerifier{
				apiKey: "super-secret-key",
				rL:     5,
				rT:     time.Second * 5,
			},
			apiKey: "super-secret-key",
			want:   http.StatusOK,
		},
		{
			name: "Invalid request (Not valid key)",
			verifier: fieldsVerifier{
				apiKey: "super-secret-key",
				rL:     5,
				rT:     time.Second * 5,
			},
			apiKey: "not-super-secret-key",
			limiterTest: limiterForTest{
				countReq: 4,
				timeWait: 0,
			},
			want: http.StatusUnauthorized,
		},
		{
			name: "Invalid request (With no key)",
			verifier: fieldsVerifier{
				apiKey: "super-secret-key",
				rL:     5,
				rT:     time.Second * 5,
			},
			apiKey: "",
			want:   http.StatusUnauthorized,
		},
		{
			name: "DoS",
			verifier: fieldsVerifier{
				apiKey: "super-secret-key",
				rL:     5,
				rT:     time.Second * 5,
			},
			apiKey: "super-secret-key",
			limiterTest: limiterForTest{
				countReq: 6,
				timeWait: 0,
			},
			want: http.StatusTooManyRequests,
		},
		{
			name: "DoS with wait",
			verifier: fieldsVerifier{
				apiKey: "super-secret-key",
				rL:     5,
				rT:     time.Second * 5,
			},
			apiKey: "super-secret-key",
			limiterTest: limiterForTest{
				countReq: 6,
				timeWait: time.Second * 6,
			},
			want: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := NewVerifier([]string{tt.verifier.apiKey}, tt.verifier.rL, tt.verifier.rT)

			r := utilapi.NewRouter(slog.Default())
			r.Handle("/", verifier.Verify)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-API-Key", tt.apiKey)

			w := httptest.NewRecorder()

			for i := 0; i <= tt.limiterTest.countReq-1; i++ {
				r.ServeHTTP(w, req)
			}

			time.Sleep(tt.limiterTest.timeWait)
			r.ServeHTTP(w, req)

			if w.Code != tt.want {
				t.Errorf("expected: %v get: %v", tt.want, w.Code)
			}
		})
	}
}
