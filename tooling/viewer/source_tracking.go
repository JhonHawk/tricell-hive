package viewer

import (
	"context"
	"os"
	"path/filepath"
	"sort"
)

// SourceClassification describes current Git tracking, never document authorship.
// Reasons are fixed codes; external diagnostics must not reach the index.
type SourceClassification struct {
	Status string
	Reason string
}

// SourceTrackingReader classifies absolute manifest paths in a single batch.
// An omitted result is unknown. A local result asserts a complete successful
// Git query, or a verified absence of a containing repository.
type SourceTrackingReader func(context.Context, []string) map[string]SourceClassification

func sessionSourcePaths(d *indexData) []string {
	paths := make([]string, 0, len(d.Sessions))
	for _, sess := range d.Sessions {
		paths = append(paths, filepath.Join(sess.Source, manifestFile))
	}
	return paths
}

func (s *Store) classifySources(ctx context.Context, paths []string) map[string]SourceClassification {
	unique := map[string]bool{}
	for _, path := range paths {
		unique[path] = true
	}
	paths = paths[:0]
	for path := range unique {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var results map[string]SourceClassification
	if s.tracking != nil && len(paths) > 0 {
		results = s.tracking(ctx, paths)
	}
	out := make(map[string]SourceClassification, len(paths))
	for _, path := range paths {
		c := results[path]
		switch c.Status {
		case "tracked":
			c.Reason = "git_index"
		case "local":
			if info, err := os.Lstat(path); err != nil || !info.Mode().IsRegular() {
				c = SourceClassification{Status: "unknown", Reason: "source_unavailable"}
			} else {
				c.Reason = ""
			}
		default:
			c.Status = "unknown"
			switch c.Reason {
			case "git_query_failed", "git_output_incomplete", "source_unavailable":
			default:
				c.Reason = "tracking_unavailable"
			}
		}
		out[path] = c
	}
	return out
}

func (s *Store) classifySessions(d *indexData, classification map[string]SourceClassification) {
	for i := range d.Sessions {
		c, ok := classification[filepath.Join(d.Sessions[i].Source, manifestFile)]
		if !ok {
			c = SourceClassification{Status: "unknown", Reason: "tracking_unavailable"}
		}
		d.Sessions[i].SourceStatus, d.Sessions[i].SourceReason = c.Status, c.Reason
	}
}
