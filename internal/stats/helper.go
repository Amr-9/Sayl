package stats

import (
	"regexp"
)

var (
	// Regex to strip ephemeral ports from error messages
	// Matches: IP:PORT->IP:PORT (e.g., 127.0.0.1:54321->127.0.0.1:80)
	// and IP:PORT (e.g. dial tcp 127.0.0.1:5432)
	rePortPair   = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d+->\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d+`)
	reSinglePort = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d+`)
)

// copyMapStringInt returns a shallow copy of a map[string]int.
// Used by Snapshot() to prevent callers from holding a reference to the
// pre-allocated snap maps that get cleared on the next Snapshot() call.
func copyMapStringInt(src map[string]int) map[string]int {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]int, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func sanitizeError(err string) string {
	// First, try to replace source->dest pairs
	err = rePortPair.ReplaceAllString(err, "[CONN_TUPLE]")
	// Then, replace single IP:PORT occurrences (e.g. strict dependency or dial)
	err = reSinglePort.ReplaceAllString(err, "[IP]:[PORT]")
	return err
}
