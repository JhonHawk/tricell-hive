package codex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFeatureEnabled(t *testing.T) {
	const key = "default_mode_request_user_input"
	ptr := func(s string) *string { return &s }
	cases := []struct {
		name    string
		config  *string
		want    bool
		wantErr bool
	}{
		{name: "no config file", want: false},
		{name: "empty file", config: ptr(""), want: false},
		{name: "no features table", config: ptr("model = \"gpt\"\n[profiles.x]\nmodel = \"y\"\n"), want: false},
		{name: "table without the key", config: ptr("[features]\nother = true\n"), want: false},
		{name: "table true", config: ptr("[features]\ndefault_mode_request_user_input = true\n"), want: true},
		{name: "table false", config: ptr("[features]\ndefault_mode_request_user_input = false\n"), want: false},
		{name: "table true with comment", config: ptr("[features] # flags\ndefault_mode_request_user_input = true # on\n"), want: true},
		{name: "table among others", config: ptr("model = \"a\"\n\n[features]\nfoo = false\ndefault_mode_request_user_input = true\n\n[tui]\nx = 1\n"), want: true},
		{name: "key in another table ignored", config: ptr("[profiles.p]\ndefault_mode_request_user_input = true\n[features]\nfoo = true\n"), want: false},
		{name: "key under profiles features ignored", config: ptr("[profiles.p.features]\ndefault_mode_request_user_input = true\n"), want: false},
		{name: "dotted root true", config: ptr("features.default_mode_request_user_input = true\n"), want: true},
		{name: "dotted root false", config: ptr("features.default_mode_request_user_input = false\n"), want: false},
		{name: "dotted root before a table", config: ptr("features.default_mode_request_user_input = true\n[tui]\nx = 1\n"), want: true},
		{name: "both forms same value", config: ptr("features.default_mode_request_user_input = true\n[features]\ndefault_mode_request_user_input = true\n"), want: true},
		{name: "both forms different values", config: ptr("features.default_mode_request_user_input = true\n[features]\ndefault_mode_request_user_input = false\n"), wantErr: true},
		{name: "duplicate table", config: ptr("[features]\nfoo = true\n[features]\ndefault_mode_request_user_input = true\n"), wantErr: true},
		{name: "inline table", config: ptr("features = { default_mode_request_user_input = true }\n"), wantErr: true},
		{name: "array of tables", config: ptr("[[features]]\ndefault_mode_request_user_input = true\n"), wantErr: true},
		{name: "duplicate key", config: ptr("[features]\ndefault_mode_request_user_input = true\ndefault_mode_request_user_input = false\n"), wantErr: true},
		{name: "non boolean value", config: ptr("[features]\ndefault_mode_request_user_input = \"yes\"\n"), wantErr: true},
		{name: "dotted subtable of features", config: ptr("[features.default_mode_request_user_input]\nx = 1\n"), wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			if tc.config != nil {
				if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(*tc.config), 0600); err != nil {
					t.Fatal(err)
				}
			}
			got, err := FeatureEnabled(home, key)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Fatalf("enabled = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFeatureEnabledRejectsUnreadableConfig(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, "config.toml"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := FeatureEnabled(home, "default_mode_request_user_input"); err == nil {
		t.Fatal("a config.toml that cannot be read as a file returned no error")
	}
}
