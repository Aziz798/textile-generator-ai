package main

import (
	"context"
	"log"
	"os"

	"textile-generator-ai/internal/config"
	"textile-generator-ai/internal/imagegen"
	"textile-generator-ai/internal/worker"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env not found")
	}

	log.Println("Textile AI generator starting...")
	settings := config.Load(
		os.Getenv("AZURE_OPENAI_API_KEY"),
		os.Getenv("AZURE_OPENAI_ENDPOINT"),
		os.Getenv("AZURE_OPENAI_API_VERSION"),
		os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT"),
		os.Getenv("AZURE_OPENAI_IMAGE_DEPLOYMENT"),
	)

	if settings.APIKey == "" || settings.Endpoint == "" || settings.APIVersion == "" || settings.ChatDeployment == "" || settings.ImageDeployment == "" {
		log.Fatal("AZURE_OPENAI_API_KEY, AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_API_VERSION, AZURE_OPENAI_CHAT_DEPLOYMENT, and AZURE_OPENAI_IMAGE_DEPLOYMENT must be set")
	}

	generator := imagegen.New(settings.Endpoint, settings.APIKey, settings.APIVersion, settings.ChatDeployment, settings.ImageDeployment)
	if err := worker.New(generator, settings).Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
