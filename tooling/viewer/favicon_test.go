package viewer

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestServerServesHiveFaviconOverHTTP(t *testing.T) {
	s := newSrvEnv(t)
	server := httptest.NewUnstartedServer(nil)
	port := server.Listener.Addr().(*net.TCPAddr).Port
	handler, err := NewServer(ServerOptions{Store: s.store, Token: testToken, Port: port})
	if err != nil {
		t.Fatal(err)
	}
	server.Config.Handler = handler
	server.Start()
	t.Cleanup(server.Close)
	want, err := os.ReadFile("../../site/src/assets/hive-mark.svg")
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"/favicon.ico", "/assets/hive-mark.svg"} {
		for _, method := range []string{http.MethodGet, http.MethodHead} {
			req, err := http.NewRequest(method, server.URL+target, nil)
			if err != nil {
				t.Fatal(err)
			}
			res, err := server.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode != http.StatusOK || res.Header.Get("Content-Type") != "image/svg+xml" {
				t.Errorf("%s %s: status %d, MIME %q", method, target, res.StatusCode, res.Header.Get("Content-Type"))
			}
			if res.Header.Get("Content-Security-Policy") != contentSecurityPolicy || res.Header.Get("X-Content-Type-Options") != "nosniff" || res.Header.Get("Cross-Origin-Resource-Policy") != "same-origin" {
				t.Errorf("%s %s changed security headers", method, target)
			}
			if method == http.MethodGet && !bytes.Equal(body, want) {
				t.Errorf("%s differs from the existing Hive mark", target)
			}
			if method == http.MethodHead && (len(body) != 0 || res.Header.Get("Content-Length") != strconv.Itoa(len(want))) {
				t.Errorf("HEAD %s has incorrect body or length", target)
			}
		}
		if got := s.do(http.MethodPost, target, reqOpts{}); got.Code != http.StatusMethodNotAllowed {
			t.Errorf("POST %s = %d", target, got.Code)
		}
		if got := s.do(http.MethodGet, target, reqOpts{host: "evil.example"}); got.Code != http.StatusForbidden {
			t.Errorf("foreign host %s = %d", target, got.Code)
		}
	}
	entry := s.do(http.MethodGet, "/", reqOpts{})
	if !strings.Contains(entry.Body.String(), `<link rel="icon" type="image/svg+xml" href="/assets/hive-mark.svg">`) {
		t.Error("entrypoint does not declare the same-origin SVG favicon")
	}
	for _, target := range []string{"/assets/../hive-mark.svg", "/assets/favicon.ico", "/favicon.ico/", "/assets/hive-mark.svg/extra"} {
		if got := s.do(http.MethodGet, target, reqOpts{}); got.Code != http.StatusNotFound {
			t.Errorf("undeclared asset %s = %d", target, got.Code)
		}
	}
}
