package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func selectionOffers() []providerOffer {
	return []providerOffer{
		{ID: "agent-browser", Name: "agent-browser", Source: "official", Unversioned: true},
		{ID: "codex-question-card", Name: "codex-question-card", Source: "Codex", DefaultOn: true, Effect: "enables Codex's question card in Default mode"},
		{ID: "engram", Name: "engram", Source: "official", ManualOnly: true},
	}
}

func selectionTerminal(input string, interactive bool) (installTerminal, *bytes.Buffer) {
	var out bytes.Buffer
	return installTerminal{reader: bufio.NewReader(strings.NewReader(input)), out: &out, interactive: interactive}, &out
}

func TestSelectProviderRequestsEnterAcceptsDefaultOffers(t *testing.T) {
	terminal, out := selectionTerminal("\n", true)
	requests, ok, err := selectProviderRequests(terminal, selectionOffers())
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if len(requests) != 1 || requests[0] != (providerRequest{ID: "codex-question-card"}) {
		t.Fatalf("requests = %v", requests)
	}
	for _, want := range []string{
		"(selected by default)",
		"Press Enter to accept the default selection, or type none to skip.\n",
		"Enter capability numbers separated by commas: ",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q in %q", want, out.String())
		}
	}
	if strings.Contains(out.String(), "Exact version") {
		t.Fatalf("default offers must not ask for a version: %q", out.String())
	}
}

func TestSelectProviderRequestsNoneSkipsEverything(t *testing.T) {
	terminal, _ := selectionTerminal("none\n", true)
	requests, ok, err := selectProviderRequests(terminal, selectionOffers())
	if err != nil || !ok || len(requests) != 0 {
		t.Fatalf("requests=%v ok=%v err=%v", requests, ok, err)
	}
}

func TestSelectProviderRequestsEnterWithoutDefaultOffersSkips(t *testing.T) {
	terminal, out := selectionTerminal("\n", true)
	offers := []providerOffer{selectionOffers()[0], selectionOffers()[2]}
	requests, ok, err := selectProviderRequests(terminal, offers)
	if err != nil || !ok || len(requests) != 0 {
		t.Fatalf("requests=%v ok=%v err=%v", requests, ok, err)
	}
	if !strings.Contains(out.String(), "or press Enter to skip: ") || strings.Contains(out.String(), "selected by default") {
		t.Fatalf("legacy prompt expected: %q", out.String())
	}
}

func TestSelectProviderRequestsNonInteractiveSelectsNothing(t *testing.T) {
	terminal, _ := selectionTerminal("\n", false)
	requests, ok, err := selectProviderRequests(terminal, selectionOffers())
	if err != nil || !ok || len(requests) != 0 {
		t.Fatalf("requests=%v ok=%v err=%v", requests, ok, err)
	}
}

func TestSelectProviderRequestsChosenEffectOfferAsksNoVersion(t *testing.T) {
	terminal, out := selectionTerminal("2\n", true)
	requests, ok, err := selectProviderRequests(terminal, selectionOffers())
	if err != nil || !ok || len(requests) != 1 || requests[0] != (providerRequest{ID: "codex-question-card"}) {
		t.Fatalf("requests=%v ok=%v err=%v out=%q", requests, ok, err, out.String())
	}
}

func TestCapsViewChecksDefaultOffersInitially(t *testing.T) {
	v := newCapsView(selectionOffers(), nil, nil)
	th := newAppTheme(true, false)
	out := stripANSI(v.View(viewCtx{Width: 100, Height: 21, Theme: &th}))
	for _, want := range []string{
		"[ ] agent-browser",
		"[x] codex-question-card",
		"[ ] engram",
		"The checked ones are selected by default; uncheck any you do not want.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Version:") {
		t.Fatalf("an effect offer opens no version field:\n%s", out)
	}
	var got []providerRequest
	v.onDone = func(r []providerRequest) (tea.Cmd, action) { got = r; return nil, action{} }
	v.proceed()
	if len(got) != 1 || got[0] != (providerRequest{ID: "codex-question-card"}) {
		t.Fatalf("requests = %v", got)
	}
}

func TestCapsViewKeepsLegacyHelpWithoutDefaultOffers(t *testing.T) {
	v := newCapsView([]providerOffer{selectionOffers()[0]}, nil, nil)
	th := newAppTheme(true, false)
	out := stripANSI(v.View(viewCtx{Width: 100, Height: 21, Theme: &th}))
	if !strings.Contains(out, "None is fine: leave them unchecked and continue.") {
		t.Fatalf("legacy help expected:\n%s", out)
	}
}

func TestProviderOfferLabelUsesEffect(t *testing.T) {
	got := providerOfferLabel(selectionOffers()[1])
	if !strings.Contains(got, "enables Codex's question card in Default mode") || strings.Contains(got, "installs the latest version") {
		t.Fatalf("label = %q", got)
	}
	if got := providerOfferLabel(selectionOffers()[2]); !strings.Contains(got, "manual instructions only") {
		t.Fatalf("manual label = %q", got)
	}
}

// The text installer, with the Codex offer present and accepted by Enter,
// keeps every fixed line within 80 columns.
func TestInstallTextWithCodexQuestionCardOfferStaysWithin80Columns(t *testing.T) {
	home := t.TempDir()
	adapter := &testOnboardingAdapter{
		offers: []providerOffer{selectionOffers()[1]},
		runner: testExternalRunner{},
	}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.AdapterFactory = func(onboardingInput) (onboardingAdapter, error) { return adapter, nil }
	var out bytes.Buffer
	if err := installWithDependencies(installArgs(t, home), strings.NewReader("\ny\n"), &out, true, dependencies); err != nil {
		t.Fatal(err)
	}
	if len(adapter.requests) != 1 || adapter.requests[0].ID != "codex-question-card" {
		t.Fatalf("requests = %v", adapter.requests)
	}
	if !strings.Contains(out.String(), "(selected by default)") {
		t.Fatalf("offer not marked: %s", out.String())
	}
	assertTerminalWidth(t, out.String(), home)
}
