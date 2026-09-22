package main

import "fmt"

const prismSourceRef = "v1.5.27"

func prismRawURL(file string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/%s/%s", prismSourceRef, file)
}
