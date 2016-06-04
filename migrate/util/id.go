package util

import (
	"crypto/md5"
	"fmt"
	"io"
	"time"
)

// PropertyIDGen Generate an MD5 hash of the seed input data.  If no seed is provided
// then the current unix timestamp is used
func PropertyIDGen(seed string) (hash string) {
	h := md5.New()

	seed += fmt.Sprintf("%d", time.Now().Nanosecond())

	io.WriteString(h, seed)
	hash = fmt.Sprintf("%x", h.Sum(nil))

	return hash
}
