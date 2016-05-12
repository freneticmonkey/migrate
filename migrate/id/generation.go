package id

import (
	"crypto/md5"
	"fmt"
	"io"
	"time"
)

// Gen Generate an MD5 hash of the seed input data.  If no seed is provided
// then the current unix timestamp is used
func Gen(seed string) (hash string) {
	h := md5.New()
	if len(seed) == 0 {
		seed = fmt.Sprintf("%d", time.Now().Unix())
	}
	io.WriteString(h, seed)
	hash = fmt.Sprintf("%x", h.Sum(nil))

	return hash
}
