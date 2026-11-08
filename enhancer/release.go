package main

import "fmt"

const uiVersion = "1.5.32"
const prismSourceRef = "v" + uiVersion

func prismRawURL(file string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/xcxcadc/chenfei-Glass-Prism-dns/refs/tags/%s/%s", prismSourceRef, file)
}
