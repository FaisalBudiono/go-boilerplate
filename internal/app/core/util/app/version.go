package app

import (
	"embed"
	"errors"
	"io/fs"
	"log"
	"strings"
)

//go:embed ver/*
var versionEmbed embed.FS

var versionFileName = "ver/version.txt"

var versionDefault = "latest"

func Version() string {
	ver, err := versionEmbed.ReadFile(versionFileName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return versionDefault
		}

		log.Fatalf("failed to read version file: %s", err)
	}

	return strings.Join(strings.Fields(string(ver)), " ")
}
