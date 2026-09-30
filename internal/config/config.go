package config

type Settings struct {
	APIKey        string
	InputDir      string
	OutputDir     string
	ReferencePath string
}

func Load(apiKey string) Settings {
	return Settings{
		APIKey:        apiKey,
		InputDir:      "input",
		OutputDir:     "output",
		ReferencePath: "reference/catalog-style.png",
	}
}
