package viewer

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func accessRequest(s *Server, path, method, token, origin, host, contentType, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Host = host
	r.Header.Set(tokenHeader, token)
	r.Header.Set("Origin", origin)
	r.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}
func issueCode(t *testing.T, s *Server) string {
	t.Helper()
	w := accessRequest(s, "/api/access-codes", "POST", testToken, "http://"+testHost, testHost, "", "")
	if w.Code != 200 {
		t.Fatalf("issue status = %d", w.Code)
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("code must not be cached")
	}
	var v struct {
		Code string `json:"code"`
	}
	if json.Unmarshal(w.Body.Bytes(), &v) != nil || len(v.Code) != 64 || v.Code == testToken {
		t.Fatal("invalid issued code")
	}
	return v.Code
}
func exchangeCode(s *Server, code string) *httptest.ResponseRecorder {
	return accessRequest(s, "/api/access-codes/exchange", "POST", "", "http://"+testHost, testHost, "application/json", `{"code":"`+code+`"}`)
}
func TestAccessCodesSingleUseAndRestart(t *testing.T) {
	s := newSrvEnv(t).srv
	code := issueCode(t, s)
	w := exchangeCode(s, code)
	if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("exchange status = %d", w.Code)
	}
	var v struct {
		Token string `json:"token"`
	}
	if json.Unmarshal(w.Body.Bytes(), &v) != nil || v.Token != testToken {
		t.Fatal("exchange did not return token")
	}
	if exchangeCode(s, code).Code != 403 {
		t.Fatal("code reused")
	}
	code = issueCode(t, s)
	restarted, err := NewServer(ServerOptions{Store: s.store, Token: testToken, Port: testPort})
	if err != nil {
		t.Fatal(err)
	}
	if exchangeCode(restarted, code).Code != 403 {
		t.Fatal("pending code survived restart")
	}
}
func TestAccessCodesRejectInvalidRequests(t *testing.T) {
	s := newSrvEnv(t).srv
	code := issueCode(t, s)
	for _, path := range []string{"/api/access-codes", "/api/access-codes/exchange"} {
		for _, tc := range []struct{ name, token, origin, host string }{
			{"no origin", testToken, "", testHost}, {"null", testToken, "null", testHost}, {"foreign origin", testToken, "http://evil.example", testHost}, {"foreign host", testToken, "http://evil.example", "evil.example"},
		} {
			t.Run(path+tc.name, func(t *testing.T) {
				w := accessRequest(s, path, "POST", tc.token, tc.origin, tc.host, "application/json", `{"code":"`+code+`"}`)
				if w.Code != 403 {
					t.Fatalf("status = %d", w.Code)
				}
			})
		}
	}
	if w := accessRequest(s, "/api/access-codes", "POST", "", "http://"+testHost, testHost, "", ""); w.Code != 403 {
		t.Fatalf("unauthenticated issue = %d", w.Code)
	}
	for _, tc := range []struct {
		method, ct, body string
		status           int
	}{
		{"GET", "application/json", `{}`, 405}, {"POST", "text/plain", `{}`, 415}, {"POST", "application/json", `{}`, 403}, {"POST", "application/json", `{"code":"bad"}`, 403}, {"POST", "application/json", `{"code":3}`, 400}, {"POST", "application/json", `{"extra":true}`, 400}, {"POST", "application/json", `{} {}`, 400}, {"POST", "application/json", strings.Repeat(" ", 1025) + `{}`, 400},
	} {
		w := accessRequest(s, "/api/access-codes/exchange", tc.method, "", "http://"+testHost, testHost, tc.ct, tc.body)
		if w.Code != tc.status {
			t.Errorf("invalid exchange status = %d, want %d", w.Code, tc.status)
		}
	}
	if w := accessRequest(s, "/api/access-codes", "GET", testToken, "http://"+testHost, testHost, "", ""); w.Code != 405 {
		t.Fatalf("issue method = %d", w.Code)
	}
	if w := accessRequest(s, "/api/access-codes", "POST", testToken, "http://"+testHost, testHost, "", "x"); w.Code != 400 {
		t.Fatalf("issue body = %d", w.Code)
	}
	if exchangeCode(s, strings.Repeat("0", 64)).Code != 403 {
		t.Fatal("unknown code accepted")
	}
	if exchangeCode(s, code).Code != 200 {
		t.Fatal("invalid requests consumed valid code")
	}
}
func TestAccessCodesConcurrentExchange(t *testing.T) {
	s := newSrvEnv(t).srv
	code := issueCode(t, s)
	var wg sync.WaitGroup
	results := make(chan int, 16)
	for range 16 {
		wg.Go(func() { results <- exchangeCode(s, code).Code })
	}
	wg.Wait()
	close(results)
	success := 0
	for status := range results {
		if status == 200 {
			success++
		} else if status != 403 {
			t.Errorf("unexpected status %d", status)
		}
	}
	if success != 1 {
		t.Fatalf("successful exchanges = %d", success)
	}
}
func TestAccessCodesExpireAndBoundPending(t *testing.T) {
	s := newSrvEnv(t).srv
	now := time.Now()
	s.accessNow = func() time.Time { return now }
	code := issueCode(t, s)
	now = now.Add(59 * time.Second)
	if exchangeCode(s, code).Code != 200 {
		t.Fatal("code expired before 60 seconds")
	}
	code = issueCode(t, s)
	now = now.Add(60 * time.Second)
	if exchangeCode(s, code).Code != 403 {
		t.Fatal("code valid at expiry")
	}
	code = issueCode(t, s)
	for range 63 {
		issueCode(t, s)
	}
	if w := accessRequest(s, "/api/access-codes", "POST", testToken, "http://"+testHost, testHost, "", ""); w.Code != 429 {
		t.Fatalf("pending limit = %d", w.Code)
	}
	if exchangeCode(s, code).Code != 200 {
		t.Fatal("limit invalidated existing code")
	}
	issueCode(t, s)
	now = now.Add(60 * time.Second)
	issueCode(t, s)
}
