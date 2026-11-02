package viewer

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

func validateSessionTitle(title string) (string, error) {
	if !utf8.ValidString(title) {
		return "", errors.New("session title must be valid UTF-8")
	}
	for _, r := range title {
		if unicode.IsControl(r) {
			return "", errors.New("session title must not contain control characters")
		}
	}
	title = strings.TrimSpace(title)
	if utf8.RuneCountInString(title) > maxNameRunes {
		return "", errors.New("session title exceeds 200 Unicode code points")
	}
	return title, nil
}
