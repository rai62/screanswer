package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:    "screanswer",
	Version: "v1.0.6",
	Usage: "Capture a text on the screen, copy it to the clipboard, " +
		"answer it automatically using ChatGPT, " +
		"and read the output aloud in a system voice",
	Commands: []*cli.Command{
		&captureCommand,
		&answerCommand,
		&speakCommand,
	},
	Flags: []cli.Flag{answerClipFlag},
	Action: func(ctx *cli.Context) error {
		answerQueue := make(chan string, 10)
		defer close(answerQueue)

		speakQueue := make(chan SpeakRequest, 10)
		defer close(speakQueue)

		captureClient, err := NewCaptureClient(ctx.Context)
		if err != nil {
			return err
		}
		defer captureClient.client.Close()

		answerClient, err := NewAnswerClient(ctx.Context, ctx.Bool(answerClipFlag.Name))
		if err != nil {
			return err
		}

		speakClient, err := NewSpeakClient(ctx.Context, 48000)
		if err != nil {
			return err
		}
		defer speakClient.client.Close()

		var wg sync.WaitGroup
		wg.Add(3)
		go captureClient.CaptureWithQueue(&wg, answerQueue)
		go answerClient.AnswerWithQueue(&wg, answerQueue, speakQueue)
		go speakClient.SpeakWithQueue(&wg, speakQueue)
		wg.Wait()

		return nil
	},
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
