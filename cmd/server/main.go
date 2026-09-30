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

	settings := config.Load(os.Getenv("OPENAI_API_KEY"))

	log.Println("Textile AI generator starting...")

	if settings.APIKey == "" {
		log.Fatal("OPENAI_API_KEY is not set")
	}

	generator := imagegen.New(settings.APIKey)
	if err := worker.New(generator, settings).Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
