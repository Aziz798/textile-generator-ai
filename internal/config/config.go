package config

import (
	"os"
	"strconv"
)

type Settings struct {
	APIKey                string
	InputDir              string
	OutputDir             string
	ReferencePath         string
	GenerationConcurrency int
}

func Load(apiKey string) Settings {
	concurrency := 3
	if value, err := strconv.Atoi(os.Getenv("GENERATION_CONCURRENCY")); err == nil && value > 0 {
		concurrency = value
	}

	return Settings{
		APIKey:                apiKey,
		InputDir:              "prestige collection_input",
		OutputDir:             "prestige collection_output_second",
		ReferencePath:         "reference/trasparenze_collection_reference.png",
		GenerationConcurrency: concurrency,
	}
}
