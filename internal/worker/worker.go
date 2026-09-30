package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"textile-generator-ai/internal/catalog"
	"textile-generator-ai/internal/config"
)

type ImageGenerator interface {
	ProductName(context.Context, string) (string, error)
	Generate(context.Context, string, string, string) error
}

type Worker struct {
	generator ImageGenerator
	settings  config.Settings
}

func New(generator ImageGenerator, settings config.Settings) *Worker {
	return &Worker{generator: generator, settings: settings}
}

func (w *Worker) Run(ctx context.Context) error {
	images, err := filepath.Glob(filepath.Join(w.settings.InputDir, "*"))
	if err != nil {
		return fmt.Errorf("find input images: %w", err)
	}
	if err := ensureOutputDir(w.settings.OutputDir); err != nil {
		return err
	}

	log.Printf("found %d file(s) in input directory", len(images))
	registry := catalog.NewNameRegistry()
	processed := 0
	outputs := 0
	for _, fabricPath := range images {
		if !catalog.IsImageFile(fabricPath) {
			log.Printf("ignoring non-image file: %s", fabricPath)
			continue
		}
		processed++
		log.Printf("processing image %d: %s", processed, fabricPath)

		name, err := w.generator.ProductName(ctx, fabricPath)
		if err != nil {
			log.Printf("skip %s: %v", fabricPath, err)
			continue
		}

		outputPath := filepath.Join(w.settings.OutputDir, registry.OutputName(name)+".jpg")
		if err := w.generator.Generate(ctx, fabricPath, w.settings.ReferencePath, outputPath); err != nil {
			log.Printf("skip %s: %v", fabricPath, err)
			continue
		}

		outputs++
		log.Printf("generated %s", outputPath)
	}
	log.Printf("batch complete: processed=%d outputs=%d", processed, outputs)
	return nil
}

func ensureOutputDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	return nil
}
