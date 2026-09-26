package providers

import "testing"

func TestCatalogEntriesCarrySourceAndReason(t *testing.T) {
	for _, provider := range Catalog() {
		if provider.ID == "" {
			t.Fatalf("provider with empty ID: %#v", provider)
		}
		if len(provider.Hosts) == 0 {
			t.Fatalf("%s advertises no hosts", provider.ID)
		}
		if provider.Source == "" {
			t.Fatalf("%s has no official source", provider.ID)
		}
		if provider.Reason == "" {
			t.Fatalf("%s has no reason it is manual-only", provider.ID)
		}
	}
}
