package main

import (
	"embed"
	"io/fs"
)

//go:embed frontend/dist
var embeddedAssets embed.FS

func assetsFS() fs.FS {
	sub, err := fs.Sub(embeddedAssets, "frontend/dist")
	if err != nil {
		panic(err)
	}
	return sub
}
