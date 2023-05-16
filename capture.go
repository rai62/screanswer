package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"cloud.google.com/go/vision/apiv1"
	"github.com/urfave/cli/v2"
	"golang.design/x/clipboard"
)

var captureCommand = cli.Command{
	Name:    "capture",
	Aliases: []string{"c"},
	Usage:   "Capture a text on the screen and copy it to the clipboard",
	Action: func(ctx *cli.Context) error {
		captureClient, err := NewCaptureClient(ctx.Context)
		if err != nil {
			return err
		}
		defer captureClient.client.Close()

		captureClient.Capture()

		return nil
	},
}

// CaptureClient is a client for capturing a text on the screen
type CaptureClient struct {
	client *vision.ImageAnnotatorClient
	ctx    context.Context
}

// NewCaptureClient creates a new CaptureClient
func NewCaptureClient(ctx context.Context) (*CaptureClient, error) {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		return nil, errors.New("GOOGLE_APPLICATION_CREDENTIALS environment variable is not set")
	}

	if err := clipboard.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize the clipboard: %v", err)
	}

	imageClient, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the image client: %v", err)
	}

	return &CaptureClient{
		client: imageClient,
		ctx:    ctx,
	}, nil
}

// Capture captures a text on the screen and copies it to the clipboard
func (c *CaptureClient) Capture() {
	c.capture(nil)
}

// CaptureWithQueue captures a text on the screen, copies it to the clipboard and sends it to the answerQueue
func (c *CaptureClient) CaptureWithQueue(wg *sync.WaitGroup, answerQueue chan<- string) {
	c.capture(answerQueue)
	wg.Done()
}

func (c *CaptureClient) capture(answerQueue chan<- string) {
	fmt.Println("On macOS, use Ctrl+Shift+Cmd+4")

	// Watch the clipboard for a screenshot
	ch := clipboard.Watch(c.ctx, clipboard.FmtImage)
	for data := range ch {
		fmt.Printf("\n%s\n\n", "Captured a screenshot")

		image, err := vision.NewImageFromReader(bytes.NewReader(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create the image reader: %v", err)
			continue
		}

		texts, err := c.client.DetectTexts(c.ctx, image, nil, 10)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to detect texts: %v", err)
			continue
		}

		if texts != nil {
			text := texts[0].Description
			fmt.Println(text)

			if answerQueue != nil {
				answerQueue <- text
			}

			clipboard.Write(clipboard.FmtText, []byte(text))
		}
	}
}
