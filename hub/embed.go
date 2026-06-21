package couchdev

import (
	"embed"
	"io/fs"
)

//go:embed web/dist
var webDist embed.FS

var WebFS fs.FS

func init() {
	var err error
	WebFS, err = fs.Sub(webDist, "web/dist")
	if err != nil {
		panic(err)
	}
}
