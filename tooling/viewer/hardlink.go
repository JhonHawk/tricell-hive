package viewer

import (
	"errors"
	"os"
	"syscall"
)

// validateSingleLink requires the Unix link count; missing metadata cannot
// establish that the file has a single link. Ownership is not restricted.
func validateSingleLink(info os.FileInfo) error {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || st == nil {
		return errors.New("cannot determine file hard link count")
	}
	if st.Nlink > 1 {
		return errors.New("file has more than one hard link")
	}
	return nil
}
