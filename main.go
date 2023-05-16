package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/urfave/cli/v2"
)

var (
	noSpeakFlag = &cli.BoolFlag{
		Name:    "nospeak",
		Aliases: []string{"ns"},
		Usage:   "Disable speaking the answer aloud",
	}

	app = &cli.App{
		Name:    "screanswer",
		Version: "v1.0.7",
		Usage: "Capture a text on the screen, copy it to the clipboard, " +
			"answer it automatically using ChatGPT, " +
			"and read the output aloud in a system voice",
		Commands: []*cli.Command{
			&captureCommand,
			&answerCommand,
			&speakCommand,
		},
		Flags: []cli.Flag{
			answerClipFlag,
			noSpeakFlag,
		},
		Action: func(ctx *cli.Context) error {
			var wg sync.WaitGroup

			captureClient, err := NewCaptureClient(ctx.Context)
			if err != nil {
				return err
			}
			defer captureClient.client.Close()

			answerQueue := make(chan string, 10)
			defer close(answerQueue)

			answerClient, err := NewAnswerClient(ctx.Context, ctx.Bool(answerClipFlag.Name))
			if err != nil {
				return err
			}

			if ctx.Bool(noSpeakFlag.Name) {
				wg.Add(2)
				go captureClient.CaptureWithQueue(&wg, answerQueue)
				go answerClient.AnswerWithAnswerQueue(&wg, answerQueue)
				wg.Wait()

				return nil
			}

			speakQueue := make(chan SpeakRequest, 10)
			defer close(speakQueue)

			speakClient, err := NewSpeakClient(ctx.Context, 48000)
			if err != nil {
				return err
			}
			defer speakClient.client.Close()

			wg.Add(3)
			go captureClient.CaptureWithQueue(&wg, answerQueue)
			go answerClient.AnswerWithQueues(&wg, answerQueue, speakQueue)
			go speakClient.SpeakWithQueue(&wg, speakQueue)
			wg.Wait()

			return nil
		},
	}
)

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
