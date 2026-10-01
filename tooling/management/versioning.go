package management

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/version"
)

const stateVersion = 6

// hexDigest64 matches a lowercase SHA-256 hex digest; validateProduct
// checks both a product's artifact and release ID against it.
var hexDigest64 = regexp.MustCompile(`^[0-9a-f]{64}$`)

type ProductIdentity struct {
	Version    string `json:"version"`
	ArtifactID string `json:"artifact_id"`
	ReleaseID  string `json:"release_id"`
}
type InstalledResource struct {
	Path       string `json:"path"`
	Hash       string `json:"hash"`
	Mode       uint32 `json:"mode,omitempty"`
	LinkTarget string `json:"link_target,omitempty"`
}
type InstallationReceipt struct {
	Consumer  Consumer            `json:"consumer"`
	Product   ProductIdentity     `json:"product"`
	Resources []InstalledResource `json:"resources"`
}

func validProductVersion(v string) bool { return version.Valid(v) }

// productFromSource identifies o.Source's product version, preferring a
// packaged release.json (via distribution.ReadManifestData, the same manifest
// packaging and installation verify) over a bare VERSION file (via
// version.ReadSourceFile). A source tree with neither is historical/
// unversioned: identity stays nil, not "dev", so it is never mistaken for a
// resolved version.
func productFromSource(o Options, r Release, s State) (*ProductIdentity, error) {
	if o.ReleaseID != "" {
		return nil, nil
	} // Historical snapshots retain content identity only.
	productVersion, artifact := "", ""
	m, b, err := distribution.ReadManifestData(o.Source)
	if err == nil {
		productVersion = m.ProductVersion
		artifact = hash(b)
	} else if os.IsNotExist(err) {
		value, present, rerr := version.ReadSourceFile(o.Source)
		if rerr != nil {
			return nil, rerr
		}
		if !present {
			return nil, nil
		}
		productVersion = value
		artifact = hash(append([]byte(productVersion+"\n"), encode(r)...))
	} else {
		return nil, err
	}
	if productVersion == "" {
		return nil, nil
	}
	p := &ProductIdentity{productVersion, artifact, r.ID}
	if err = validateProduct(p, s, r.ID); err != nil {
		return nil, err
	}
	return p, nil
}
func validateProduct(p *ProductIdentity, s State, release string) error {
	if p == nil {
		return nil
	}
	if !validProductVersion(p.Version) || len(p.ArtifactID) != 64 || p.ReleaseID != release || len(p.ReleaseID) != 64 || !hexDigest64.MatchString(p.ArtifactID) || !hexDigest64.MatchString(p.ReleaseID) {
		return fmt.Errorf("invalid product identity")
	}
	if old, ok := s.Versions[p.Version]; p.Version != "dev" && ok && old != *p {
		return fmt.Errorf("published version %s already identifies a different artifact", p.Version)
	}
	return nil
}
func receiptResource(r Record) InstalledResource {
	return InstalledResource{r.Target.Path, hash(r.Managed), r.Mode, r.Target.LinkTarget}
}
func updateProductState(next *State, old State, p Plan) bool {
	next.Versions = map[string]ProductIdentity{}
	for k, v := range old.Versions {
		next.Versions[k] = v
	}
	next.Installations = map[string]InstallationReceipt{}
	for k, v := range old.Installations {
		next.Installations[k] = v
	}
	context := p.Config.Home
	if p.Config.Scope == "project" {
		context = p.Config.Root
	}
	for _, host := range p.Hosts {
		c := Consumer{host, p.Config.Scope, context}
		key := consumerKey(c)
		if p.Action == "remove" || p.Product == nil {
			delete(next.Installations, key)
			continue
		}
		receipt := InstallationReceipt{Consumer: c, Product: *p.Product}
		for _, ch := range p.Changes {
			if ch.After != nil && hasConsumer(ch.After.Consumers, c) {
				receipt.Resources = append(receipt.Resources, receiptResource(*ch.After))
			}
		}
		sort.Slice(receipt.Resources, func(i, j int) bool { return receipt.Resources[i].Path < receipt.Resources[j].Path })
		next.Installations[key] = receipt
	}
	if p.Product != nil {
		next.Versions[p.Product.Version] = *p.Product
	}
	if len(next.Versions) == 0 {
		next.Versions = nil
	}
	if len(next.Installations) == 0 {
		next.Installations = nil
	}
	return !reflect.DeepEqual(next.Versions, old.Versions) || !reflect.DeepEqual(next.Installations, old.Installations)
}
func installationVersion(s State, c Consumer) (string, string) {
	receipt, ok := s.Installations[consumerKey(c)]
	if !ok {
		return "", "legacy"
	}
	expected := map[string]InstalledResource{}
	for _, r := range receipt.Resources {
		expected[r.Path] = r
	}
	count := 0
	for _, r := range s.Records {
		if hasConsumer(r.Consumers, c) {
			count++
			if x, ok := expected[r.Target.Path]; !ok || x != receiptResource(r) {
				return receipt.Product.Version, "partial"
			}
			cur, err := readResource(r.Target, false)
			if err != nil || owned(cur, r, hiveMarkers) != nil {
				return receipt.Product.Version, "drift"
			}
		}
	}
	if count != len(expected) {
		return receipt.Product.Version, "partial"
	}
	return receipt.Product.Version, "verified"
}

// RequiredHosts returns the explicit cohort needed to change shared payloads.
// It never writes or adds hosts to a plan on the caller's behalf.
func RequiredHosts(o Options) ([]string, error) {
	hosts, err := validateHosts(o.Hosts)
	if err != nil {
		return nil, err
	}
	c, dir, err := normalize(o)
	if err != nil {
		return nil, err
	}
	s, _, err := readState(dir)
	if err != nil {
		return nil, err
	}
	r, err := loadRelease(o, dir)
	if err != nil {
		return nil, err
	}
	context := c.Home
	if c.Scope == "project" {
		context = c.Root
	}
	all := map[string]bool{}
	for _, h := range hosts {
		all[h] = true
	}
	for {
		selected := []string{}
		for h := range all {
			selected = append(selected, h)
		}
		sort.Strings(selected)
		groups, err := desiredResources(c, selected, s, "install", &r)
		if err != nil {
			return nil, err
		}
		added := false
		for _, g := range groups {
			old, ok := s.Records[g.Target.Path]
			if !ok {
				continue
			}
			current, err := readResource(g.Target, g.Replaces != nil)
			if err != nil {
				return nil, err
			}
			p := Plan{Action: "install", Config: c, Release: &r}
			// Preview the new bytes with all known consumers to avoid the conflict gate.
			expanded := g
			expanded.Consumers = sortedConsumers(append(append([]Consumer{}, g.Consumers...), old.Consumers...))
			after, err := nextRecord(p, expanded, &old, current, false)
			if err != nil {
				return nil, err
			}
			if after == nil || !reflect.DeepEqual(old.Managed, after.Managed) || old.Mode != after.Mode || old.Target.LinkTarget != after.Target.LinkTarget {
				for _, co := range old.Consumers {
					if co.Scope != c.Scope || co.Context != context {
						return nil, fmt.Errorf("shared resource crosses scopes; explicit coordinated operation required")
					}
					if !all[co.Host] {
						all[co.Host] = true
						added = true
					}
				}
			}
		}
		if !added {
			return selected, nil
		}
	}
}

func validateProductState(s State) error {
	for label, p := range s.Versions {
		if label != p.Version {
			return fmt.Errorf("invalid version index")
		}
		if err := validateProduct(&p, State{}, p.ReleaseID); err != nil {
			return err
		}
	}
	for key, r := range s.Installations {
		if key != consumerKey(r.Consumer) {
			return fmt.Errorf("invalid installation receipt")
		}
		if _, err := validateHosts([]string{r.Consumer.Host}); err != nil {
			return err
		}
		if r.Consumer.Scope != "user" && r.Consumer.Scope != "project" {
			return fmt.Errorf("invalid receipt scope")
		}
		if !filepath.IsAbs(r.Consumer.Context) {
			return fmt.Errorf("invalid receipt context")
		}
		if err := validateProduct(&r.Product, State{}, r.Product.ReleaseID); err != nil {
			return err
		}
		for _, entry := range r.Resources {
			if !filepath.IsAbs(entry.Path) || len(entry.Hash) != 64 {
				return fmt.Errorf("invalid receipt resource")
			}
		}
	}
	return nil
}
