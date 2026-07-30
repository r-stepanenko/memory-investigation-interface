package main

import (
	"flag"
	"os"
)

type CLIConfig struct {
	Addr        string
	DatasetsDir string
}

func ParseCLI() CLIConfig {

	defaultAddr := ":8080"

	if port := os.Getenv("PORT"); port != "" {
		defaultAddr = ":" + port
	}

	addr := flag.String(
		"addr",
		defaultAddr,
		"server address",
	)

	datasets := flag.String(
		"datasets",
		"./internal/datasets",
		"datasets directory",
	)

	flag.Parse()

	return CLIConfig{
		Addr:        *addr,
		DatasetsDir: *datasets,
	}
}