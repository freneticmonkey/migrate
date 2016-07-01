package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func main() {
	paths := []string{
		"/tmp/test/working/animals",
		"/tmp/test/something",
		"/tmp/test/working/../../something",
	}
	base := "/tmp/test/working"

	fmt.Println("On Unix:")
	for _, p := range paths {
		rel, err := filepath.Rel(base, p)
		fmt.Printf("%q: %q %v\n", p, rel, err)

		if strings.HasPrefix(rel, "..") {
			fmt.Printf("DENIED MF! %s\n", rel)
		} else {
			fmt.Printf("G2G! %s\n", rel)
		}
	}

}
