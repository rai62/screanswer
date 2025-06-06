package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v2"
	"golang.design/x/clipboard"
)

var (
	answerClipFlag = &cli.BoolFlag{
		Name:    "clip",
		Aliases: []string{"c"},
		Usage:   "Enable to write an answer to the clipboard",
	}

	answerCommand = cli.Command{
		Name:      "answer",
		Aliases:   []string{"a"},
		Usage:     "Answer a text using chat AI",
		ArgsUsage: "<text>",
		Flags:     []cli.Flag{answerClipFlag},
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Len() != 1 {
				return errors.New("incorrect usage: not enough command line arguments")
			}

			answerClient, err := NewAnswerClient(ctx.Context, ctx.Bool(answerClipFlag.Name))
			if err != nil {
				return err
			}

			answerClient.Answer(ctx.Args().Get(0))

			return nil
		},
	}
)

// AnswerClient is a client for answering a text using chat AI.
type AnswerClient struct {
	client    *openai.Client
	ctx       context.Context
	isClipped bool
}

// NewAnswerClient creates a new AnswerClient.
func NewAnswerClient(ctx context.Context, isClipped bool) (*AnswerClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable is not set")
	}

	return &AnswerClient{
		client:    openai.NewClient(apiKey),
		ctx:       ctx,
		isClipped: isClipped,
	}, nil
}

// Answer answers a text using chat AI.
func (c *AnswerClient) Answer(text string) {
	if err := c.answer(text, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// AnswerWithAnswerQueue answers a text from the answerQueue using chat AI.
func (c *AnswerClient) AnswerWithAnswerQueue(wg *sync.WaitGroup, answerQueue <-chan string) {
	for text := range answerQueue {
		if err := c.answer(text, nil); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	wg.Done()
}

// AnswerWithQueues answers a text from the answerQueue using chat AI and sends the answer to the speakQueue.
func (c *AnswerClient) AnswerWithQueues(wg *sync.WaitGroup, answerQueue <-chan string, speakQueue chan<- SpeakRequest) {
	for text := range answerQueue {
		if err := c.answer(text, speakQueue); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	wg.Done()
}

func (c *AnswerClient) answer(text string, speakQueue chan<- SpeakRequest) error {
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			},
		},
		Stream: true,
	}
	stream, err := c.client.CreateChatCompletionStream(c.ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create the answer stream: %v", err)
	}
	defer stream.Close()

	var sentences, sentence []byte
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Printf("\n\n%s\n", "Stream finished")
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive the answer: %v", err)
		}

		character := response.Choices[0].Delta.Content
		fmt.Printf(character)

		sentence = append(sentence, character...)

		if speakQueue != nil {
			switch character {
			case ".", "!", "?":
				sentences = append(sentences, sentence...)
				sendSpeakRequest(speakQueue, "en-US", &sentence)
			case "。", "！", "？":
				sentences = append(sentences, sentence...)
				sendSpeakRequest(speakQueue, "ja-JP", &sentence)
			}
		}
	}

	if c.isClipped {
		sentences = append(sentences, sentence...)
		clipboard.Write(clipboard.FmtText, sentences)
	}

	return nil
}

func sendSpeakRequest(sQueue chan<- SpeakRequest, langCode string, s *[]byte) {
	sQueue <- SpeakRequest{
		languageCode: langCode,
		text:         string(*s),
	}

	*s = nil
}
