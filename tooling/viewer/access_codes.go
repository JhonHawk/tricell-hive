package viewer

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"time"
)

const accessCodeTTL = 60 * time.Second
const maxAccessCodes = 64

// pruneAccessCodes runs only while accessMu is held.
func (s *Server) pruneAccessCodes(now time.Time) {
	for code, expires := range s.accessCodes {
		if !now.Before(expires) {
			delete(s.accessCodes, code)
		}
	}
}

func (s *Server) issueAccessCode(w http.ResponseWriter, r *http.Request) {
	if !allow(w, r, http.MethodPost) || !s.authorize(w, r, true) {
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1))
	if err != nil || len(body) != 0 {
		httpError(w, 400, "bad request")
		return
	}
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		httpError(w, 500, "internal error")
		return
	}
	code := hex.EncodeToString(random[:])
	s.accessMu.Lock()
	now := s.accessNow()
	s.pruneAccessCodes(now)
	if len(s.accessCodes) >= maxAccessCodes {
		s.accessMu.Unlock()
		httpError(w, 429, "too many pending opens")
		return
	}
	s.accessCodes[code] = now.Add(accessCodeTTL)
	s.accessMu.Unlock()
	writeJSON(w, struct {
		Code string `json:"code"`
	}{code})
}

func (s *Server) exchangeAccessCode(w http.ResponseWriter, r *http.Request) {
	if !allow(w, r, http.MethodPost) {
		return
	}
	if !s.originOK(r, true) {
		httpError(w, 403, "forbidden")
		return
	}
	ct, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || ct != "application/json" {
		httpError(w, 415, "unsupported media type")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var input struct {
		Code string `json:"code"`
	}
	if dec.Decode(&input) != nil || dec.Decode(new(any)) != io.EOF {
		httpError(w, 400, "bad request")
		return
	}
	decoded, err := hex.DecodeString(input.Code)
	if err != nil || len(input.Code) != 64 || len(decoded) != 32 {
		httpError(w, 403, "forbidden")
		return
	}
	s.accessMu.Lock()
	s.pruneAccessCodes(s.accessNow())
	_, ok := s.accessCodes[input.Code]
	delete(s.accessCodes, input.Code)
	s.accessMu.Unlock()
	if !ok {
		httpError(w, 403, "forbidden")
		return
	}
	writeJSON(w, struct {
		Token string `json:"token"`
	}{string(s.token)})
}
