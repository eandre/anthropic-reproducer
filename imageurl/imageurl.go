package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/anthropics/anthropic-sdk-go/vertex"
	"golang.org/x/oauth2/google"
)

var _ = vertex.WithCredentials // to keep import

var vertexProjectID = flag.String("vertex-project-id", "", "The Vertex AI project ID to use")

func main() {
	flag.Parse()

	apiKey := os.Getenv("ANTHROPIC_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_KEY environment variable not set")
	}

	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	model := anthropic.ModelClaudeSonnet4_5_20250929

	// Enable Vertex AI, if requested.
	if *vertexProjectID != "" {
		creds, err := google.FindDefaultCredentials(context.Background(), "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			log.Fatalf("failed to get google default credentials: %v", err)
		}
		opts = append(opts, vertex.WithCredentials(context.Background(), "global", *vertexProjectID, creds))
		model = anthropic.Model("claude-sonnet-4-5@20250929")
	}

	params := anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: 2000,
		System: []anthropic.TextBlockParam{
			{Text: "You are a helpful assistant that can analyze images."},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("What is the dominant color in this image?")),
			anthropic.NewAssistantMessage(anthropic.NewToolUseBlock(
				"toolu_read_image",
				json.RawMessage(`{"path": "/path/to/image.png"}`),
				"read_image",
			)),
			anthropic.NewUserMessage(
				anthropic.ContentBlockParamUnion{
					OfToolResult: &anthropic.ToolResultBlockParam{
						ToolUseID: "toolu_read_image",
						IsError:   param.NewOpt(false),
						Content: []anthropic.ToolResultBlockParamContentUnion{
							{
								OfImage: &anthropic.ImageBlockParam{
									Type: "image",
									Source: anthropic.ImageBlockParamSourceUnion{
										OfURL: &anthropic.URLImageSourceParam{
											Type: "url",
											URL:  "https://images.unsplash.com/photo-1506905925346-21bda4d32df4",
										},
									},
								},
							},
						},
					},
				}),
		},
	}

	log.Println("Making request with image URL in tool response...")

	client := anthropic.NewClient(opts...)
	stream := client.Messages.NewStreaming(context.Background(), params)
	var resp anthropic.Message
	for stream.Next() {
		event := stream.Current()
		if err := resp.Accumulate(event); err != nil {
			log.Fatalf("failed to accumulate response: %v", err)
		}
	}

	if err := stream.Err(); err != nil {
		log.Fatalf("Stream error: %v", err)
	}

	log.Printf("got response: %+v", resp)

	log.Println("Image in tool response test completed successfully")
}
