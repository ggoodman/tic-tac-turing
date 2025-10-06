// Package content embeds markdown content files for the website.
package content

import "embed"

//go:embed *.md
var Files embed.FS
