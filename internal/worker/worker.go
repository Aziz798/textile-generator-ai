package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"textile-generator-ai/internal/catalog"
	"textile-generator-ai/internal/config"
)

type ImageGenerator interface {
	ProductName(context.Context, string) (string, error)
	Generate(context.Context, string, string, string, string) error
}

var rooms = []string{"living room", "bedroom", "dining room", "kitchen", "home office", "hotel room"}

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

	var imagePaths []string
	for _, imagePath := range images {
		if catalog.IsImageFile(imagePath) {
			imagePaths = append(imagePaths, imagePath)
		} else {
			log.Printf("ignoring non-image file: %s", imagePath)
		}
	}

	log.Printf("found %d image file(s) in input directory", len(imagePaths))
	registry := catalog.NewNameRegistry()
	concurrency := w.settings.GenerationConcurrency
	if concurrency > len(imagePaths) {
		concurrency = len(imagePaths)
	}
	if concurrency == 0 {
		log.Printf("batch complete: processed=0 outputs=0")
		return nil
	}

	jobs := make(chan int)
	var waitGroup sync.WaitGroup
	var registryMu sync.Mutex
	var outputsMu sync.Mutex
	outputs := 0
	for workerIndex := 0; workerIndex < concurrency; workerIndex++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for imageIndex := range jobs {
				fabricPath := imagePaths[imageIndex]
				log.Printf("processing image %d: %s", imageIndex+1, fabricPath)

				name, err := w.generator.ProductName(ctx, fabricPath)
				if err != nil {
					log.Printf("skip %s: %v", fabricPath, err)
					continue
				}

				registryMu.Lock()
				outputPath := filepath.Join(w.settings.OutputDir, registry.OutputName(name)+".jpg")
				registryMu.Unlock()
				if err := w.generator.Generate(ctx, fabricPath, w.settings.ReferencePath, outputPath, rooms[imageIndex%len(rooms)]); err != nil {
					log.Printf("skip %s: %v", fabricPath, err)
					continue
				}

				outputsMu.Lock()
				outputs++
				outputsMu.Unlock()
				log.Printf("generated %s", outputPath)
			}
		}()
	}
	for imageIndex := range imagePaths {
		jobs <- imageIndex
	}
	close(jobs)
	waitGroup.Wait()

	log.Printf("batch complete: processed=%d outputs=%d", len(imagePaths), outputs)
	return nil
}

func ensureOutputDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	return nil
}
