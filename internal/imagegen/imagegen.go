package imagegen

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/azure"
)

type Generator struct {
	client          *openai.Client
	chatDeployment  string
	imageDeployment string
}

type imageReader struct {
	io.Reader
	filename  string
	mediaType string
}

func (r imageReader) Filename() string    { return r.filename }
func (r imageReader) ContentType() string { return r.mediaType }

func (g *Generator) ProductName(ctx context.Context, fabricPath string) (string, error) {
	started := time.Now()
	log.Printf("reading product name from %s", fabricPath)

	data, err := os.ReadFile(fabricPath)
	if err != nil {
		return "", fmt.Errorf("read fabric image: %w", err)
	}

	mediaType := mime.TypeByExtension(filepath.Ext(fabricPath))
	if mediaType == "" {
		return "", fmt.Errorf("unsupported image format")
	}

	result, err := g.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: g.chatDeployment,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
				openai.TextContentPart("Read the product or fabric name printed inside this image. Return only that name, with no explanation. If no name is visible, return UNKNOWN."),
				openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
					URL: "data:" + mediaType + ";base64," + base64.StdEncoding.EncodeToString(data),
				}),
			}),
		},
	})
	if err != nil {
		return "", fmt.Errorf("read product name: %w", err)
	}
	if len(result.Choices) == 0 || strings.TrimSpace(result.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("OpenAI returned no product name")
	}

	name := strings.TrimSpace(result.Choices[0].Message.Content)
	if strings.EqualFold(name, "UNKNOWN") {
		return "", fmt.Errorf("no product name visible in image")
	}

	filename := sanitizeFilename(name)
	log.Printf("product name for %s: %q (took %s)", fabricPath, filename, time.Since(started).Round(time.Millisecond))
	return filename, nil
}

func sanitizeFilename(name string) string {
	var builder strings.Builder
	for _, character := range strings.TrimSpace(name) {
		if unicode.IsLetter(character) || unicode.IsNumber(character) {
			builder.WriteRune(character)
		} else if builder.Len() > 0 && !strings.HasSuffix(builder.String(), "-") {
			builder.WriteRune('-')
		}
	}

	return strings.Trim(builder.String(), "-")
}

func New(endpoint, apiKey, apiVersion, chatDeployment, imageDeployment string) *Generator {
	client := openai.NewClient(
		azure.WithEndpoint(endpoint, apiVersion),
		azure.WithAPIKey(apiKey),
	)

	return &Generator{
		client:          &client,
		chatDeployment:  chatDeployment,
		imageDeployment: imageDeployment,
	}
}

func (g *Generator) Generate(
	ctx context.Context,
	fabricPath string,
	stylePath string,
	outputPath string,
) error {
	started := time.Now()
	log.Printf("starting image generation: fabric=%s style=%s output=%s", fabricPath, stylePath, outputPath)

	fabric, err := os.Open(fabricPath)
	if err != nil {
		return fmt.Errorf("open fabric image: %w", err)
	}
	defer fabric.Close()

	style, err := os.Open(stylePath)
	if err != nil {
		return fmt.Errorf("open style image: %w", err)
	}
	defer style.Close()

	fabricType := mime.TypeByExtension(filepath.Ext(fabricPath))
	styleType := mime.TypeByExtension(filepath.Ext(stylePath))
	if fabricType == "" || styleType == "" {
		return fmt.Errorf("unsupported image format")
	}

	prompt := `
Create a professional textile product catalog image.

The FIRST image is the actual textile that must be presented.
The SECOND image is the visual style and layout reference.

IMPORTANT:

- Preserve the actual textile's color, weave, pattern and visual character.
- Do not invent a different textile.
	- Create a photorealistic interior showing this textile used as curtains.
	- Choose the room and background that best suit the textile's color, pattern, texture and mood.
	- Do not always use a living room. Select an appropriate room such as a living room, bedroom, dining room, kitchen, home office or hotel room when it better fits the textile.
	- Use the textile's visual character to guide the interior: lighter or delicate fabrics may suit a bright bedroom or refined living space, while practical, warm or highly textured fabrics may suit a dining room, kitchen or contemporary space.
	- Vary the room architecture, furniture, lighting and decor between images when the textiles are different. Do not reuse the same background by default.
	- Keep the chosen room realistic and premium, with a coherent color palette that complements rather than changes the supplied textile.
- Create two curtain layers:
  1. a lighter sheer/inner curtain
  2. the supplied textile as the darker outer curtain
- The darker curtain must frame and hug the lighter curtain from BOTH sides.
- The composition should look like a premium Italian interior-design catalog.
	- Follow the overall visual composition, spacing, elegance and presentation style of the reference image, but do not copy its exact room, furniture or background.
- Include fabric close-ups and a fabric sample area.
- Do not invent technical specifications.
- Do not invent composition percentages.
- Do not invent product information.
- Do not add fake logos or brands.
- Keep the visual identity consistent with the reference catalog.
`

	result, err := g.client.Images.Edit(ctx, openai.ImageEditParams{
		Model: g.imageDeployment,
		Image: openai.ImageEditParamsImageUnion{
			OfFileArray: []io.Reader{
				imageReader{Reader: fabric, filename: filepath.Base(fabricPath), mediaType: fabricType},
				imageReader{Reader: style, filename: filepath.Base(stylePath), mediaType: styleType},
			},
		},
		Prompt:       prompt,
		Size:         openai.ImageEditParamsSize1536x1024,
		OutputFormat: openai.ImageEditParamsOutputFormatJPEG,
	})

	if err != nil {
		return fmt.Errorf("generate image: %w", err)
	}
	log.Printf("image API returned %d result(s) for %s", len(result.Data), fabricPath)

	if len(result.Data) == 0 {
		return fmt.Errorf("OpenAI returned no image")
	}

	if result.Data[0].B64JSON == "" {
		return fmt.Errorf("OpenAI returned empty image data")
	}

	data := result.Data[0].B64JSON

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("decode generated image: %w", err)
	}

	if err := os.WriteFile(outputPath, decoded, 0644); err != nil {
		return fmt.Errorf("write output image: %w", err)
	}
	log.Printf("finished image generation: output=%s bytes=%d took=%s", outputPath, len(decoded), time.Since(started).Round(time.Millisecond))

	return nil
}
