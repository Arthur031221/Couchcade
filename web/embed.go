// Package web holds the browser pages that ship inside the binary.
package web

import "embed"

//go:embed index.html play.html logo.svg css js locales
var FS embed.FS
