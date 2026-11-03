package viewer

import (
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// maxCapture bounds a stored capture; segmentPattern is the shape of one
// path segment; deliveryAuthorityPattern is the lexical form of an https
// delivery authority.
const maxCapture = 20 << 20

var (
	segmentPattern           = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	deliveryAuthorityPattern = regexp.MustCompile(`^[Hh][Tt][Tt][Pp][Ss]://([A-Za-z0-9._-]+|\[[0-9A-Fa-f:.]+\])(:[0-9]+)?([/?#]|$)`)
)

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

// safeDeliveryURL uses the lexical checks before parsing, so the parser cannot
// normalize away a forbidden input character.
func safeDeliveryURL(raw string) bool {
	// Apply the ASCII authority contract before either parser can normalize it.
	if !deliveryAuthorityPattern.MatchString(raw) {
		return false
	}
	for _, r := range raw {
		if r == '\\' || unicode.IsControl(r) || unicode.IsSpace(r) {
			return false
		}
	}
	for i := 0; i < len(raw); i++ {
		if raw[i] == '%' {
			if i+2 >= len(raw) || !isHex(raw[i+1]) || !isHex(raw[i+2]) {
				return false
			}
			i += 2
		}
	}
	u, err := url.Parse(raw)
	if err != nil || !strings.EqualFold(u.Scheme, "https") || u.Host == "" || u.User != nil || u.Opaque != "" {
		return false
	}
	host := u.Hostname()
	if host == "" || strings.ContainsAny(u.Host, "%@") {
		return false
	}
	if strings.ContainsAny(u.Host, "[]") {
		if !strings.HasPrefix(u.Host, "[") || !strings.Contains(host, ":") || net.ParseIP(host) == nil {
			return false
		}
	} else if strings.Contains(host, ":") {
		return false
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 0 || n > 65535 {
			return false
		}
	} else if strings.HasSuffix(u.Host, ":") {
		return false
	}
	return true
}

func isHex(b byte) bool { return b >= '0' && b <= '9' || b >= 'a' && b <= 'f' || b >= 'A' && b <= 'F' }
