package config

type Settings struct {
	APIKey          string
	Endpoint        string
	APIVersion      string
	ChatDeployment  string
	ImageDeployment string
	InputDir        string
	OutputDir       string
	ReferencePath   string
}

func Load(apiKey, endpoint, apiVersion, chatDeployment, imageDeployment string) Settings {
	return Settings{
		APIKey:          apiKey,
		Endpoint:        endpoint,
		APIVersion:      apiVersion,
		ChatDeployment:  chatDeployment,
		ImageDeployment: imageDeployment,
		InputDir:        "input",
		OutputDir:       "output",
		ReferencePath:   "reference/catalog-style.png",
	}
}
