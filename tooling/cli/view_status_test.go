package main

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestViewStatusRejectsLegacyDirectoryWithoutCreatingState(t *testing.T) {
	state := filepath.Join(t.TempDir(), "unused-state")
	var out bytes.Buffer
	err := runView([]string{"status", "--dir", filepath.Join(t.TempDir(), "walkthrough.json"), "--state-dir", state}, &out, viewDeps{})
	if err == nil || !strings.Contains(err.Error(), "unsupported_protocol") {
		t.Fatal("source-directory status did not fail cutoff", err)
	}
	if out.Len() != 0 {
		t.Fatal("rejected status printed a confirmation")
	}
	if _, err = os.Stat(state); !os.IsNotExist(err) {
		t.Fatal("rejected status created viewer state")
	}
}

func TestViewStatusWithoutDirectoryKeepsOutputAndDoesNotCreateStore(t *testing.T) {
	state := filepath.Join(t.TempDir(), "unused-state")
	var out bytes.Buffer
	if err := runView([]string{"status", "--state-dir", state}, &out, viewDeps{}); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got["running"] != false {
		t.Fatal("normal offline output changed")
	}
	if _, err := os.Stat(state); !os.IsNotExist(err) {
		t.Fatalf("status created state: %v", err)
	}
}

func TestViewStatusIndexDoesNotRedirectCredentials(t *testing.T) {
	targetCalls := 0
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { targetCalls++ }))
	defer target.Close()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer srv.Close()
	if _, err := viewIndex(serverInfo{Port: srv.Listener.Addr().(*net.TCPAddr).Port, Token: "fixture-status-access-token"}); err == nil {
		t.Fatal("index redirect accepted")
	}
	if targetCalls != 0 {
		t.Fatal("index followed redirect")
	}
}

func TestViewStatusHelpDescribesV2WithoutLegacyConfirmation(t *testing.T) {
	var out bytes.Buffer
	if err := runView([]string{"--help"}, &out, viewDeps{}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "current source revision") || strings.Contains(out.String(), "hive view status [--dir DIR]") || !strings.Contains(out.String(), "protocol v2") {
		t.Fatal("help must describe v2 status instead of manifest confirmation")
	}
}
