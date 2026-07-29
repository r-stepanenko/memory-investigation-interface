package main

import "flag"

type CLIConfig struct {
	Addr        string
	DatasetsDir string
}

func ParseCLI() CLIConfig {
	addr := flag.String(
		"addr",
		":8080",
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