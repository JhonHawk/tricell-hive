package distribution

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"tricell-hive/tooling/version"
)

const releaseOrigin = "https://github.com/JhonHawk/tricell-hive"
const releaseLatest = "https://api.github.com/repos/JhonHawk/tricell-hive/releases/latest"
const releaseTimeout = 2 * time.Minute

// ReleaseClient always addresses the official public release origin. Transport
// permits the embedding application's network policy and external-boundary tests;
// redirect validation and deadlines cannot be replaced by it.
type ReleaseClient struct{ Transport http.RoundTripper }
type ReleaseTarget struct {
	ProductVersion, Platform, Label, ArchiveURL string
	Downgrade                                   bool
}

func allowedReleaseURL(u *url.URL) bool {
	if u.Scheme != "https" || u.User != nil || u.Opaque != "" || u.Fragment != "" || (u.Port() != "" && u.Port() != "443") {
		return false
	}
	h := strings.ToLower(u.Hostname())
	return h == "github.com" || h == "api.github.com" || h == "githubusercontent.com" || strings.HasSuffix(h, ".githubusercontent.com")
}
func (c ReleaseClient) fetch(ctx context.Context, address string, limit int64) ([]byte, error) {
	u, err := url.Parse(address)
	if err != nil || !allowedReleaseURL(u) {
		return nil, fmt.Errorf("release URL is not permitted")
	}
	ctx, cancel := context.WithTimeout(ctx, releaseTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Hive-release-update")
	client := http.Client{Transport: c.Transport, Timeout: releaseTimeout, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many release redirects")
		}
		if !allowedReleaseURL(r.URL) {
			return fmt.Errorf("release redirect is not permitted")
		}
		return nil
	}}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("release download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("release download returned HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > limit {
		return nil, fmt.Errorf("release download exceeds limit")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("release download incomplete: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("release download exceeds limit")
	}
	if resp.ContentLength > 0 && int64(len(data)) != resp.ContentLength {
		return nil, fmt.Errorf("release download truncated")
	}
	return data, nil
}

// Resolve chooses latest stable or an exact product version. An exact older
// version is marked explicitly; compatibility is checked before execution.
func (c ReleaseClient) Resolve(ctx context.Context, current, requested, platform string) (ReleaseTarget, error) {
	var result ReleaseTarget
	if !version.Valid(current) || current == "dev" {
		return result, fmt.Errorf("development build: use hive update --source <checkout>")
	}
	switch platform {
	case "darwin/arm64", "linux/amd64", "linux/arm64":
	default:
		return result, fmt.Errorf("unsupported release platform")
	}
	target := requested
	if target == "" {
		data, err := c.fetch(ctx, releaseLatest, 1<<20)
		if err != nil {
			return result, err
		}
		var metadata struct {
			Tag        string `json:"tag_name"`
			Draft      bool   `json:"draft"`
			Prerelease bool   `json:"prerelease"`
		}
		if err := json.Unmarshal(data, &metadata); err != nil {
			return result, fmt.Errorf("invalid release metadata")
		}
		target = strings.TrimPrefix(metadata.Tag, "v")
		if metadata.Draft || metadata.Prerelease || strings.Contains(target, "-") {
			return result, fmt.Errorf("latest release is not stable")
		}
	}
	if !version.Valid(target) || target == "dev" {
		return result, fmt.Errorf("invalid release version")
	}
	downgrade := compareReleaseVersions(target, current) < 0
	if requested == "" && downgrade {
		return result, fmt.Errorf("latest release would downgrade %s to %s; select --version explicitly", current, target)
	}
	label := "hive-" + target + "-" + strings.ReplaceAll(platform, "/", "-")
	return ReleaseTarget{ProductVersion: target, Platform: platform, Label: label, ArchiveURL: releaseOrigin + "/releases/download/v" + target + "/" + label + ".tar.gz", Downgrade: downgrade}, nil
}

func compareNumeric(a, b string) int {
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return strings.Compare(a, b)
}
func compareReleaseVersions(a, b string) int {
	ac, ap, _ := strings.Cut(a, "-")
	bc, bp, _ := strings.Cut(b, "-")
	aa, bb := strings.Split(ac, "."), strings.Split(bc, ".")
	for i := range aa {
		if d := compareNumeric(aa[i], bb[i]); d != 0 {
			return d
		}
	}
	if ap == bp {
		return 0
	}
	if ap == "" {
		return 1
	}
	if bp == "" {
		return -1
	}
	aa, bb = strings.Split(ap, "."), strings.Split(bp, ".")
	numeric := func(v string) bool {
		for _, r := range v {
			if r < '0' || r > '9' {
				return false
			}
		}
		return true
	}
	for i := 0; i < len(aa) && i < len(bb); i++ {
		an, bn := numeric(aa[i]), numeric(bb[i])
		var d int
		switch {
		case an && bn:
			d = compareNumeric(aa[i], bb[i])
		case an:
			d = -1
		case bn:
			d = 1
		default:
			d = strings.Compare(aa[i], bb[i])
		}
		if d != 0 {
			return d
		}
	}
	if len(aa) < len(bb) {
		return -1
	}
	if len(aa) > len(bb) {
		return 1
	}
	return 0
}
