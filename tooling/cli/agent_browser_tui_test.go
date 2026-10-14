package main

import (
	"fmt"
	"testing"
)

// An unversioned offer in the Hosts view opens no version field, says it
// installs the latest version, and requests the capability with no version.
func TestHostsViewUnversionedCapabilityAsksForNoVersion(t *testing.T) {
	source := minimalTestSource(t)
	adapter := &testOnboardingAdapter{
		offers: []providerOffer{{ID: "agent-browser", Name: "agent-browser", Source: "official fixture", Unversioned: true}},
		runner: testExternalRunner{},
	}
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	deps.AdapterFactory = func(onboardingInput) (onboardingAdapter, error) { return adapter, nil }
	home, stateDir := newHostsTestHome(t)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "codex")
	d.key("a")
	d.mustShow("Optional capabilities", "[ ] agent-browser", "installs the latest version")
	d.key("space")
	d.mustShow("[x] agent-browser")
	d.mustNotShow("Version:")
	d.key("enter")
	d.mustShow("[Apply]")
	d.key("enter")
	d.mustShow("Hive installed and verified")
	if fmt.Sprint(adapter.requests) != fmt.Sprint([]providerRequest{{ID: "agent-browser"}}) {
		t.Fatalf("requests = %v", adapter.requests)
	}
}

func TestCatalogSourceMatchesDoctorSource(t *testing.T) {
	if got := providerSource("agent-browser"); got != agentBrowserSource {
		t.Fatalf("catalog source %q, doctor source %q", got, agentBrowserSource)
	}
}
