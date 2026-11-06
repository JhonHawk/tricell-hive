package distribution

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type releaseTransport func(*http.Request) (*http.Response, error)

func (f releaseTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func releaseResponse(r *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: r}
}

func TestReleaseResolutionBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, body, current, requested string
		bad                            bool
	}{
		{"stable", `{"tag_name":"v1.2.3"}`, "1.2.2", "", false},
		{"implicit downgrade", `{"tag_name":"v1.2.3"}`, "2.0.0", "", true},
		{"prerelease", `{"tag_name":"v1.2.3-rc.1","prerelease":true}`, "1.0.0", "", true},
		{"draft", `{"tag_name":"v1.2.3","draft":true}`, "1.0.0", "", true},
		{"development", "", "dev", "", true},
		{"exact downgrade", "", "2.0.0", "1.2.3", false},
		{"invalid exact", "", "1.0.0", "../../other", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			c := ReleaseClient{Transport: releaseTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				deadline, ok := r.Context().Deadline()
				remaining := time.Until(deadline)
				if !ok || remaining <= 0 || remaining > 2*time.Minute {
					t.Fatal("release request must have an active deadline bounded to two minutes")
				}
				if r.URL.String() != "https://api.github.com/repos/JhonHawk/tricell-hive/releases/latest" {
					t.Fatalf("unexpected origin: %s", r.URL)
				}
				return releaseResponse(r, 200, tc.body), nil
			})}
			got, err := c.Resolve(context.Background(), tc.current, tc.requested, "darwin/arm64")
			if (err != nil) != tc.bad {
				t.Fatalf("Resolve = %+v, %v", got, err)
			}
			if tc.current == "dev" && calls != 0 {
				t.Fatal("development downloaded")
			}
			if !tc.bad && got.Downgrade != (tc.name == "exact downgrade") {
				t.Fatalf("downgrade = %v", got.Downgrade)
			}
		})
	}
}
func TestReleaseRedirectBoundary(t *testing.T) {
	for _, location := range []string{"http://github.com/unsafe", "https://evil.example/x", "https://github.com.evil.example/x", "https://user@github.com/x", "https://github.com:444/x"} {
		t.Run(location, func(t *testing.T) {
			calls := 0
			c := ReleaseClient{Transport: releaseTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				v := releaseResponse(r, 302, "")
				v.Header.Set("Location", location)
				return v, nil
			})}
			if _, err := c.Resolve(context.Background(), "1.0.0", "", "darwin/arm64"); err == nil {
				t.Fatal("accepted redirect")
			}
			if calls != 1 {
				t.Fatalf("requested forbidden redirect: %d", calls)
			}
		})
	}
}

func TestReleaseDownloadBoundsAndCancellation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		respond releaseTransport
	}{
		{"oversized", func(r *http.Request) (*http.Response, error) {
			v := releaseResponse(r, 200, "")
			v.ContentLength = 2 << 20
			return v, nil
		}},
		{"stream oversized", func(r *http.Request) (*http.Response, error) {
			return releaseResponse(r, 200, strings.Repeat("x", (1<<20)+1)), nil
		}},
		{"truncated", func(r *http.Request) (*http.Response, error) {
			v := releaseResponse(r, 200, `{"tag_name":"v1.2.3"}`)
			v.ContentLength = 100
			return v, nil
		}},
		{"status", func(r *http.Request) (*http.Response, error) { return releaseResponse(r, 404, "missing"), nil }},
		{"canceled", func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			if tc.name == "canceled" {
				cancel()
			}
			defer cancel()
			c := ReleaseClient{Transport: tc.respond}
			if _, err := c.Resolve(ctx, "1.0.0", "", "darwin/arm64"); err == nil {
				t.Fatal("accepted invalid download")
			}
		})
	}
}

func TestReleasePermittedRedirectAndLoop(t *testing.T) {
	for _, loop := range []bool{false, true} {
		calls := 0
		c := ReleaseClient{Transport: releaseTransport(func(r *http.Request) (*http.Response, error) {
			calls++
			if calls == 1 || loop {
				v := releaseResponse(r, 302, "")
				v.Header.Set("Location", "https://release-assets.githubusercontent.com/metadata")
				return v, nil
			}
			return releaseResponse(r, 200, `{"tag_name":"v1.2.3"}`), nil
		})}
		_, err := c.Resolve(context.Background(), "1.0.0", "", "darwin/arm64")
		if (err != nil) != loop {
			t.Fatalf("loop %v: %v", loop, err)
		}
		if calls > 5 {
			t.Fatal("redirect loop unbounded")
		}
	}
}
