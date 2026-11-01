package viewer

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTimelineHTTPMultipartReadbackAndScope(t *testing.T) {
	store, request := newTimelineFixture(t)
	v := createTimeline(t, store, request)
	server, err := NewServer(ServerOptions{Store: store, Port: 4483, Token: strings.Repeat("z", 32)})
	if err != nil {
		t.Fatal(err)
	}
	call := func(method, path string, raw []byte, ct string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://127.0.0.1:4483"+path, bytes.NewReader(raw))
		r.Header.Set(tokenHeader, strings.Repeat("z", 32))
		r.Header.Set("Origin", "http://127.0.0.1:4483")
		r.Header.Set("X-Hive-Action-Scope", "session")
		if ct != "" {
			r.Header.Set("Content-Type", ct)
		}
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	upload := func(path string, metadata any, content []byte) *httptest.ResponseRecorder {
		var buf bytes.Buffer
		m := multipart.NewWriter(&buf)
		meta, _ := json.Marshal(metadata)
		f, _ := m.CreateFormField("metadata")
		f.Write(meta)
		f, _ = m.CreateFormFile("file", "../../ignored-name")
		f.Write(content)
		m.Close()
		return call("POST", path, buf.Bytes(), m.FormDataContentType())
	}
	path := "/api/timeline/sessions/" + v.ViewerSessionID
	w := upload(path+"/documents", TimelineDocumentMetadata{SchemaVersion: 1, OperationID: "http-doc", MediaType: "text/markdown"}, []byte("# Document\n"))
	if w.Code != 200 {
		t.Fatalf("upload %d %s", w.Code, w.Body.String())
	}
	op, err := DecodeTimelineOperationResult(w.Body.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	w = call("GET", path+"/documents/"+op.DocumentID, nil, "")
	if w.Code != 200 {
		t.Fatalf("document read %d", w.Code)
	}
	if _, err = DecodeTimelineDocumentResponse(w.Body.Bytes()); err != nil {
		t.Fatal(err)
	}
	request.Source.NativeSessionID = "other"
	request.OperationID = "other"
	other := createTimeline(t, store, request)
	if w = call("GET", "/api/timeline/sessions/"+other.ViewerSessionID+"/documents/"+op.DocumentID, nil, ""); w.Code != 404 {
		t.Fatalf("cross session document %d", w.Code)
	}
	var imageBytes bytes.Buffer
	png.Encode(&imageBytes, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	meta := TimelineArtifactMetadata{SchemaVersion: 1, OperationID: "http-image", CapturedAt: "2026-01-01T00:00:00Z", Candidate: TimelineCandidate{Branch: "feature", Commit: "abc", Base: "development"}, Viewport: "desktop", Theme: "dark", MediaType: "image/png"}
	w = upload(path+"/artifacts", meta, imageBytes.Bytes())
	if w.Code != 200 {
		t.Fatalf("image %d %s", w.Code, w.Body.String())
	}
	w = call("GET", path, nil, "")
	response, err := DecodeTimelineSessionResponse(w.Body.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	for _, url := range response.ArtifactURLs {
		r := httptest.NewRequest("GET", "http://127.0.0.1:4483"+url, nil)
		out := httptest.NewRecorder()
		server.ServeHTTP(out, r)
		if out.Code != 200 || !bytes.Equal(out.Body.Bytes(), imageBytes.Bytes()) {
			t.Fatalf("capability read %d", out.Code)
		}
	}
	meta.OperationID = "invalid"
	if w = upload(path+"/artifacts", meta, []byte("not an image")); w.Code != 400 {
		t.Fatal("invalid signature accepted")
	}
	r := httptest.NewRequest("GET", "http://foreign.example/api/health", nil)
	r.Header.Set(tokenHeader, strings.Repeat("z", 32))
	out := httptest.NewRecorder()
	server.ServeHTTP(out, r)
	if out.Code != 403 {
		t.Fatal("foreign host accepted")
	}
	r = httptest.NewRequest("POST", "http://127.0.0.1:4483/api/timeline/sessions", strings.NewReader("{}"))
	r.Header.Set(tokenHeader, strings.Repeat("z", 32))
	r.Header.Set("Origin", "https://foreign.example")
	out = httptest.NewRecorder()
	server.ServeHTTP(out, r)
	if out.Code != 403 {
		t.Fatal("foreign origin accepted")
	}
	w = call("GET", "/api/health", nil, "")
	var health map[string]any
	json.Unmarshal(w.Body.Bytes(), &health)
	if health["retention"] != "manual" || health["ttl"] != "0s" || health["timeline_protocol_version"] != float64(1) {
		t.Fatalf("health contract %v", health)
	}
}
func TestTimelineHTTPRejectsOversizedAndDuplicateMultipart(t *testing.T) {
	store, request := newTimelineFixture(t)
	v := createTimeline(t, store, request)
	server, _ := NewServer(ServerOptions{Store: store, Port: 4483, Token: strings.Repeat("z", 32)})
	for _, test := range []string{"too-large", "duplicate"} {
		t.Run(test, func(t *testing.T) {
			var buf bytes.Buffer
			m := multipart.NewWriter(&buf)
			metadata, _ := json.Marshal(TimelineDocumentMetadata{SchemaVersion: 1, OperationID: "test", MediaType: "text/markdown"})
			part, _ := m.CreateFormField("metadata")
			part.Write(metadata)
			part, _ = m.CreateFormFile("file", "plan.md")
			if test == "too-large" {
				io.Copy(part, io.LimitReader(&zeroReader{}, TimelineMaxDocumentBytes+1))
			} else {
				part.Write([]byte("# Plan"))
				part, _ = m.CreateFormField("metadata")
				part.Write(metadata)
			}
			m.Close()
			r := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:4483/api/timeline/sessions/"+v.ViewerSessionID+"/documents", &buf)
			r.Header.Set("Content-Type", m.FormDataContentType())
			r.Header.Set(tokenHeader, strings.Repeat("z", 32))
			r.Header.Set("Origin", "http://127.0.0.1:4483")
			r.Header.Set("X-Hive-Action-Scope", "session")
			w := httptest.NewRecorder()
			server.ServeHTTP(w, r)
			if w.Code < 400 {
				t.Fatal("unbounded multipart accepted")
			}
		})
	}
}

type zeroReader struct{}

func (*zeroReader) Read(p []byte) (int, error) { clear(p); return len(p), nil }
