package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
	"github.com/urfave/cli/v2"
)

var speakCommand = cli.Command{
	Name:        "speak",
	Aliases:     []string{"s"},
	Usage:       "Read a text aloud in a system voice",
	ArgsUsage:   "<languageCode> <text>",
	Description: "For supported languageCode, please see https://cloud.google.com/text-to-speech/docs/voices.",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() != 2 {
			return errors.New("incorrect usage: not enough command line arguments")
		}

		speakClient, err := NewSpeakClient(ctx.Context, 48000)
		if err != nil {
			return err
		}
		defer speakClient.client.Close()

		speakClient.Speak(SpeakRequest{
			languageCode: ctx.Args().Get(0),
			text:         ctx.Args().Get(1),
		})

		return nil
	},
}

// SpeakClient is a client for reading a text aloud in a system voice.
type SpeakClient struct {
	client       *texttospeech.Client
	ctx          context.Context
	otoCtx       *oto.Context
	readyCh      chan struct{}
	samplingRate int32
}

// SpeakRequest is a request for reading a text aloud in a system voice.
type SpeakRequest struct {
	languageCode string
	text         string
}

// NewSpeakClient creates a new SpeakClient.
func NewSpeakClient(ctx context.Context, samplingRate int32) (*SpeakClient, error) {
	t2sClient, err := texttospeech.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the Text-to-Speech client: %v", err)
	}

	// Remember that you should **not** create more than one context
	otoCtx, ready, err := oto.NewContext(int(samplingRate), 2, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to create the audio context: %v", err)
	}

	return &SpeakClient{
		client:       t2sClient,
		ctx:          ctx,
		otoCtx:       otoCtx,
		readyCh:      ready,
		samplingRate: samplingRate,
	}, nil
}

// Speak reads a text aloud in a system voice.
func (c *SpeakClient) Speak(request SpeakRequest) {
	if err := c.speak(request); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// SpeakWithQueue reads a text from the speakQueue aloud in a system voice.
func (c *SpeakClient) SpeakWithQueue(wg *sync.WaitGroup, speakQueue <-chan SpeakRequest) {
	for request := range speakQueue {
		if err := c.speak(request); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	wg.Done()
}

func (c *SpeakClient) speak(request SpeakRequest) error {
	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{
				Text: request.text,
			},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: request.languageCode,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding:   texttospeechpb.AudioEncoding_MP3,
			SpeakingRate:    1.5,
			SampleRateHertz: c.samplingRate,
		},
	}
	resp, err := c.client.SynthesizeSpeech(c.ctx, req)
	if err != nil {
		return fmt.Errorf("failed to synthesize the speech: %v", err)
	}

	reader := bytes.NewReader(resp.AudioContent)

	decoder, err := mp3.NewDecoder(reader)
	if err != nil {
		return fmt.Errorf("failed to create the mp3 decoder: %v", err)
	}

	// It might take a bit for the hardware audio devices to be ready, so we wait on the channel.
	<-c.readyCh

	player := c.otoCtx.NewPlayer(decoder)
	defer player.Close()

	player.Play()

	for player.IsPlaying() {
		time.Sleep(time.Millisecond)
	}

	return nil
}
